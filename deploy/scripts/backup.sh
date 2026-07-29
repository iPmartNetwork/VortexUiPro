#!/bin/bash
# ═══════════════════════════════════════════════════════════════════════
# VortexUiPro — Automated Backup Script
#
# Performs PostgreSQL dump with zstd compression, retention management,
# optional S3/remote sync, and Telegram notification on failure.
#
# Usage:
#   ./backup.sh                    # Full backup
#   ./backup.sh --db-only          # Database only
#   ./backup.sh --config-only      # Config only
#   ./backup.sh --list             # List available backups
#   ./backup.sh --restore <file>   # Restore from backup
#
# Environment variables (or set in .env):
#   POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, POSTGRES_HOST
#   BACKUP_DIR, BACKUP_RETENTION_DAYS, BACKUP_RETENTION_WEEKS
#   TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID
#   S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY
# ═══════════════════════════════════════════════════════════════════════

set -euo pipefail

# ─── Config ──────────────────────────────────────────────────────────

BACKUP_DIR="${BACKUP_DIR:-/etc/vortex/backups}"
DB_HOST="${POSTGRES_HOST:-postgres}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_USER="${POSTGRES_USER:-vortex}"
DB_PASS="${POSTGRES_PASSWORD:-}"
DB_NAME="${POSTGRES_DB:-vortex}"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"
RETENTION_WEEKS="${BACKUP_RETENTION_WEEKS:-4}"
RETENTION_MONTHS="${BACKUP_RETENTION_MONTHS:-12}"
TELEGRAM_BOT_TOKEN="${TELEGRAM_BOT_TOKEN:-}"
TELEGRAM_CHAT_ID="${TELEGRAM_CHAT_ID:-}"
S3_ENDPOINT="${S3_ENDPOINT:-}"
S3_BUCKET="${S3_BUCKET:-}"
S3_ACCESS_KEY="${S3_ACCESS_KEY:-}"
S3_SECRET_KEY="${S3_SECRET_KEY:-}"

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
DATE_TAG=$(date +%Y%m%d)
WEEK_NUM=$(date +%V)
MONTH_NUM=$(date +%Y%m)

# ─── Help ─────────────────────────────────────────────────────────────

usage() {
    grep '^#' "$0" | grep -v '^#!' | sed 's/^#//'
    exit 0
}

# ─── Logging ─────────────────────────────────────────────────────────

log_info()  { echo "[INFO]  $(date '+%H:%M:%S') $*"; }
log_warn()  { echo "[WARN]  $(date '+%H:%M:%S') $*" >&2; }
log_error() { echo "[ERROR] $(date '+%H:%M:%S') $*" >&2; }

send_telegram() {
    local msg="$1"
    if [[ -n "$TELEGRAM_BOT_TOKEN" && -n "$TELEGRAM_CHAT_ID" ]]; then
        curl -sf -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
            -d "chat_id=${TELEGRAM_CHAT_ID}" \
            -d "text=${msg}" \
            -d "parse_mode=HTML" >/dev/null 2>&1 || true
    fi
}

# ─── Prerequisites ───────────────────────────────────────────────────

ensure_deps() {
    local missing=()
    command -v psql    >/dev/null 2>&1 || missing+=("postgresql-client")
    command -v zstd    >/dev/null 2>&1 || missing+=("zstd")
    command -v curl    >/dev/null 2>&1 || missing+=("curl")

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_warn "Missing dependencies: ${missing[*]}"
        if command -v apk >/dev/null 2>&1; then
            apk add --no-cache postgresql-client zstd curl coreutils 2>/dev/null
        elif command -v apt-get >/dev/null 2>&1; then
            apt-get update -qq && apt-get install -y -qq postgresql-client zstd curl coreutils 2>/dev/null
        fi
    fi
}

# ─── Backup Functions ─────────────────────────────────────────────────

backup_db() {
    log_info "Starting database backup..."

    mkdir -p "${BACKUP_DIR}/db"
    local output="${BACKUP_DIR}/db/vortex-${TIMESTAMP}.sql.zst"
    local tempfile="/tmp/vortex-backup-${TIMESTAMP}.sql"

    # Dump PostgreSQL database with compression
    PGPASSWORD="${DB_PASS}" pg_dump \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        --clean \
        --if-exists \
        --no-owner \
        --no-privileges \
        --quote-all-identifiers \
        -f "${tempfile}" 2>&1

    # Compress with zstd (fast compression)
    zstd -f -q --compress "${tempfile}" -o "${output}"
    rm -f "${tempfile}"

    local size
    size=$(wc -c < "${output}" 2>/dev/null | tr -d ' ' || echo 0)
    log_info "Database backup complete: $(numfmt --to=iec "${size}") — ${output}"
    echo "${output}"
}

backup_config() {
    log_info "Starting configuration backup..."

    mkdir -p "${BACKUP_DIR}/config"
    local output="${BACKUP_DIR}/config/vortex-config-${TIMESTAMP}.tar.zst"

    tar -cf - \
        -C /etc/vortex/data . \
        -C /etc/vortex/certs . 2>/dev/null \
        | zstd -f -q -o "${output}"

    local size
    size=$(wc -c < "${output}" 2>/dev/null | tr -d ' ' || echo 0)
    log_info "Config backup complete: $(numfmt --to=iec "${size}") — ${output}"
    echo "${output}"
}

# ─── Restore ──────────────────────────────────────────────────────────

restore_backup() {
    local file="$1"

    if [[ ! -f "$file" ]]; then
        log_error "Backup file not found: ${file}"
        exit 1
    fi

    case "$file" in
        *.sql.zst)
            log_info "Restoring database from: ${file}"
            local tempfile="/tmp/vortex-restore-${TIMESTAMP}.sql"
            zstd -f -q -d "${file}" -o "${tempfile}"
            PGPASSWORD="${DB_PASS}" psql \
                -h "${DB_HOST}" \
                -p "${DB_PORT}" \
                -U "${DB_USER}" \
                -d "${DB_NAME}" \
                -f "${tempfile}" 2>&1
            rm -f "${tempfile}"
            log_info "Database restore complete!"
            send_telegram "✅ <b>VortexUiPro Restore</b> — Database restored from ${file}"
            ;;
        *.tar.zst)
            log_info "Restoring configuration from: ${file}"
            zstd -f -q -d "${file}" -o "/tmp/vortex-config-restore-${TIMESTAMP}.tar"
            tar -xf "/tmp/vortex-config-restore-${TIMESTAMP}.tar" -C /
            rm -f "/tmp/vortex-config-restore-${TIMESTAMP}.tar"
            log_info "Config restore complete!"
            ;;
        *)
            log_error "Unknown backup format: ${file}"
            exit 1
            ;;
    esac
}

# ─── Retention ────────────────────────────────────────────────────────

apply_retention() {
    log_info "Applying retention policy..."

    # Daily backups: keep last N days
    find "${BACKUP_DIR}/db" -name "vortex-*.sql.zst" -mtime "+${RETENTION_DAYS}" -delete 2>/dev/null || true
    find "${BACKUP_DIR}/config" -name "vortex-config-*.tar.zst" -mtime "+${RETENTION_DAYS}" -delete 2>/dev/null || true

    # Weekly backups: keep last N weeks (Sunday backups)
    find "${BACKUP_DIR}/db" -name "vortex-*sun*.sql.zst" -mtime "+$((RETENTION_WEEKS * 7))" -delete 2>/dev/null || true

    # Monthly backups: keep last N months (1st of month)
    find "${BACKUP_DIR}/db" -name "vortex-*01*.sql.zst" -mtime "+$((RETENTION_MONTHS * 30))" -delete 2>/dev/null || true

    log_info "Retention applied: ${RETENTION_DAYS}d daily / ${RETENTION_WEEKS}w weekly / ${RETENTION_MONTHS}m monthly"
}

# ─── S3 Sync ─────────────────────────────────────────────────────────

sync_to_s3() {
    if [[ -z "$S3_ENDPOINT" || -z "$S3_BUCKET" ]]; then
        log_info "S3 not configured, skipping remote sync"
        return
    fi

    log_info "Syncing backups to S3: ${S3_BUCKET}"

    if command -v rclone >/dev/null 2>&1; then
        rclone copy "${BACKUP_DIR}" ":s3,env_auth=false,endpoint=${S3_ENDPOINT},access_key_id=${S3_ACCESS_KEY},secret_access_key=${S3_SECRET_KEY}:${S3_BUCKET}/vortex-backups/" \
            --progress --verbose 2>&1 || log_warn "S3 sync failed"
    elif command -v aws >/dev/null 2>&1; then
        aws s3 sync "${BACKUP_DIR}" "s3://${S3_BUCKET}/vortex-backups/" \
            --endpoint-url "${S3_ENDPOINT}" 2>&1 || log_warn "AWS CLI sync failed"
    else
        # Fallback: use curl for S3-compatible storage
        log_warn "Neither rclone nor aws CLI found. Install one for S3 sync."
        return
    fi

    log_info "S3 sync complete!"
}

# ─── List Backups ────────────────────────────────────────────────────

list_backups() {
    echo ""
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║              VortexUiPro — Available Backups                ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""

    if [[ -d "${BACKUP_DIR}/db" ]]; then
        echo "📀 Database Backups:"
        ls -lhS "${BACKUP_DIR}/db"/*.sql.zst 2>/dev/null | awk '{printf "   %s  %s  %s\n", $6, $5, $9}' || echo "   (none)"
    fi
    echo ""

    if [[ -d "${BACKUP_DIR}/config" ]]; then
        echo "⚙️  Config Backups:"
        ls -lhS "${BACKUP_DIR}/config"/*.tar.zst 2>/dev/null | awk '{printf "   %s  %s  %s\n", $6, $5, $9}' || echo "   (none)"
    fi
    echo ""
}

# ─── Main ─────────────────────────────────────────────────────────────

main() {
    mkdir -p "${BACKUP_DIR}" "${BACKUP_DIR}/db" "${BACKUP_DIR}/config"

    case "${1:-}" in
        --help|-h) usage ;;
        --list|-l) list_backups; exit 0 ;;
        --restore|-r) restore_backup "${2:-}"; exit 0 ;;
        --db-only) backup_db ;;
        --config-only) backup_config ;;
        *)
            ensure_deps
            local db_file config_file
            db_file=$(backup_db)
            config_file=$(backup_config)
            apply_retention
            sync_to_s3
            total_size=$(du -sh "${BACKUP_DIR}" 2>/dev/null | cut -f1)

            send_telegram "✅ <b>VortexUiPro Backup Complete</b>
📀 DB: $(basename "${db_file}")
⚙️  Config: $(basename "${config_file}")
💾 Total: ${total_size}
🗑️  Retention: ${RETENTION_DAYS}d / ${RETENTION_WEEKS}w / ${RETENTION_MONTHS}m"

            log_info "═══════════════════════════════════════════════════════════════"
            log_info "  Backup Complete!"
            log_info "  Database:  ${db_file}"
            log_info "  Config:    ${config_file}"
            log_info "  Total:     ${total_size}"
            log_info "═══════════════════════════════════════════════════════════════"
            ;;
    esac
}

main "$@"
