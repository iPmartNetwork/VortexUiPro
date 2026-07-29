#!/bin/bash
# ═══════════════════════════════════════════════════════════════════════
# VortexUiPro — Server-Side Deployment Script
#
# Runs on the target server to pull, build, and restart the stack.
# Can be called from:
#   - GitHub Actions (CI/CD)
#   - Manual SSH session
#   - Cron-based auto-update
#
# Usage:
#   ./deploy.sh                            # Deploy latest (staging)
#   ./deploy.sh production                 # Deploy to production
#   ./deploy.sh staging v1.2.3             # Deploy specific version
#   ./deploy.sh production latest          # Deploy latest production
#   ./deploy.sh --rollback                 # Rollback to previous version
#   ./deploy.sh --status                   # Check deployment status
#   ./deploy.sh --cleanup                  # Clean up old images
#
# Environment variables:
#   VORTEX_ENV       : staging | production (default: staging)
#   VORTEX_VERSION   : Docker image tag (default: latest)
#   VORTEX_PROJECT   : Project directory (default: /opt/vortexuipro)
#   VORTEX_COMPOSE   : Compose file path (default: deploy/production.yml)
# ═══════════════════════════════════════════════════════════════════════

set -euo pipefail

# ─── Config ──────────────────────────────────────────────────────────

VORTEX_ENV="${VORTEX_ENV:-${1:-staging}}"
VORTEX_VERSION="${VORTEX_VERSION:-${2:-latest}}"
VORTEX_PROJECT="${VORTEX_PROJECT:-/opt/vortexuipro}"
VORTEX_COMPOSE="${VORTEX_COMPOSE:-deploy/production.yml}"
VORTEX_BACKUP="${VORTEX_BACKUP:-true}"
VORTEX_HEALTH_RETRIES="${VORTEX_HEALTH_RETRIES:-12}"
VORTEX_HEALTH_INTERVAL="${VORTEX_HEALTH_INTERVAL:-10}"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# ─── Colors ──────────────────────────────────────────────────────────

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log_info()  { echo -e "${BLUE}[INFO]${NC}  $(date '+%H:%M:%S') $*"; }
log_ok()    { echo -e "${GREEN}[OK]${NC}    $(date '+%H:%M:%S') $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $(date '+%H:%M:%S') $*" >&2; }
log_error() { echo -e "${RED}[ERROR]${NC} $(date '+%H:%M:%S') $*" >&2; }
log_step()  { echo -e "\n${CYAN}══════════════════════════════════════════════════${NC}"; echo -e "${CYAN}  ➜ $*${NC}"; echo -e "${CYAN}══════════════════════════════════════════════════${NC}"; }

# ─── Help ─────────────────────────────────────────────────────────────

usage() {
    grep '^#' "$0" | grep -v '^#!' | sed 's/^#//'
    exit 0
}

# ─── Prerequisites ───────────────────────────────────────────────────

check_prereqs() {
    local missing=0
    for cmd in docker git curl; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            log_error "Missing dependency: $cmd"
            missing=1
        fi
    done

    if ! command -v docker compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
        log_error "Docker Compose v2 is required"
        missing=1
    fi

    if [ "$missing" -eq 1 ]; then
        exit 1
    fi
}

# ─── Git Operations ──────────────────────────────────────────────────

git_pull() {
    log_step "Pulling latest code"

    if [ ! -d "${VORTEX_PROJECT}/.git" ]; then
        log_info "First time setup — cloning repository..."
        git clone "https://github.com/iPmartNetwork/VortexUiPro.git" "${VORTEX_PROJECT}"
    fi

    cd "${VORTEX_PROJECT}"
    log_info "Current branch: $(git branch --show-current)"
    log_info "Current commit: $(git rev-parse --short HEAD)"

    # Stash any local changes
    git stash --include-untracked 2>/dev/null || true

    # Fetch and reset to latest
    git fetch --tags --force --prune
    git checkout master 2>/dev/null || git checkout main 2>/dev/null
    git reset --hard "@{upstream}"

    log_ok "Updated to: $(git rev-parse --short HEAD)"
}

# ─── Environment ─────────────────────────────────────────────────────

load_env() {
    log_step "Loading environment: ${VORTEX_ENV}"

    cd "${VORTEX_PROJECT}"

    local env_file="${VORTEX_PROJECT}/.env.${VORTEX_ENV}"
    if [ -f "$env_file" ]; then
        cp "$env_file" .env
        log_ok "Loaded: .env.${VORTEX_ENV}"
    elif [ -f ".env" ]; then
        log_warn "No .env.${VORTEX_ENV} found, using existing .env"
    else
        log_error "No .env file found! Create .env.${VORTEX_ENV} from .env.example"
        log_info "  cp .env.example .env.${VORTEX_ENV}"
        log_info "  # Edit .env.${VORTEX_ENV} with your settings"
        log_info "  # Then re-run this script"
        exit 1
    fi

    # Source the env file for this script
    set -a; source .env 2>/dev/null || true; set +a
}

# ─── Pre-deploy Backup ───────────────────────────────────────────────

pre_deploy_backup() {
    if [ "${VORTEX_BACKUP}" != "true" ]; then
        log_info "Backup skipped (VORTEX_BACKUP=false)"
        return
    fi

    log_step "Pre-deploy database backup"

    # Check if backup container exists and is running
    if docker ps --format '{{.Names}}' | grep -q 'vortexuipro-backup'; then
        docker exec vortexuipro-backup /usr/local/bin/backup.sh --db-only 2>&1 || log_warn "Backup command failed"
        log_ok "Pre-deploy backup completed"
    else
        log_warn "Backup container not running — skipping pre-deploy backup"
        log_info "  To enable backups, start the stack first: docker compose -f ${VORTEX_COMPOSE} up -d backup"
    fi
}

# ─── Pull Images ─────────────────────────────────────────────────────

pull_images() {
    log_step "Pulling Docker images (${VORTEX_VERSION})"

    cd "${VORTEX_PROJECT}"

    # Update image tags in compose file
    local owner="${GITHUB_REPOSITORY_OWNER:-ipmartnetwork}"
    owner=$(echo "$owner" | tr '[:upper:]' '[:lower:]')

    if [ "${VORTEX_VERSION}" != "latest" ]; then
        sed -i "s|:latest|:${VORTEX_VERSION}|g" "${VORTEX_COMPOSE}"
    fi

    docker compose -f "${VORTEX_COMPOSE}" pull --quiet || log_warn "Image pull had warnings"

    log_ok "Images pulled/updated"
}

# ─── Deploy Services ─────────────────────────────────────────────────

deploy_services() {
    log_step "Starting services"

    cd "${VORTEX_PROJECT}"

    # Create required data directories
    mkdir -p deploy/data/{postgres,redis,panel,backups,logs}

    # Start all services
    docker compose -f "${VORTEX_COMPOSE}" up -d --remove-orphans --wait 2>/dev/null || \
    docker compose -f "${VORTEX_COMPOSE}" up -d --remove-orphans

    log_ok "All services started"
}

# ─── Health Check ────────────────────────────────────────────────────

wait_healthy() {
    log_step "Waiting for services to become healthy"

    local services=("vortexuipro-postgres" "vortexuipro-redis" "vortexuipro-panel" "vortexuipro-caddy")
    local all_healthy=true

    for svc in "${services[@]}"; do
        log_info "Checking: ${svc}..."

        for i in $(seq 1 ${VORTEX_HEALTH_RETRIES}); do
            local status
            status=$(docker inspect --format='{{.State.Health.Status}}' "${svc}" 2>/dev/null || echo "not_found")

            if [ "${status}" = "healthy" ]; then
                log_ok "${svc} ✓ (attempt ${i})"
                break
            elif [ "${status}" = "not_found" ]; then
                log_warn "${svc} — container not found"
                all_healthy=false
                break
            else
                if [ "${i}" -eq "${VORTEX_HEALTH_RETRIES}" ]; then
                    log_warn "${svc} — not healthy after ${VORTEX_HEALTH_RETRIES} attempts (status: ${status})"
                    all_healthy=false
                else
                    sleep "${VORTEX_HEALTH_INTERVAL}"
                fi
            fi
        done
    done

    if [ "$all_healthy" = true ]; then
        log_ok "All services healthy!"
    else
        log_warn "Some services may not be fully healthy — check with: docker ps"
    fi

    # Show status
    echo ""
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
}

# ─── Cleanup ─────────────────────────────────────────────────────────

cleanup() {
    log_step "Cleaning up"

    # Remove stopped containers
    docker container prune -f 2>/dev/null || true

    # Remove old images (older than 24h)
    docker image prune -af --filter 'until=24h' 2>/dev/null || true

    # Remove unused volumes (careful!)
    # docker volume prune -f 2>/dev/null || true

    log_ok "Cleanup complete"
}

# ─── Rollback ────────────────────────────────────────────────────────

rollback() {
    log_step "Rolling back to previous version"

    cd "${VORTEX_PROJECT}"

    # Git rollback
    log_info "Rolling back Git..."
    git reflog | head -5
    git reset --hard HEAD@{1}

    # Docker rollback — restart with previous compose
    log_info "Restarting services with previous config..."
    docker compose -f "${VORTEX_COMPOSE}" up -d --force-recreate

    log_ok "Rollback complete — monitoring health..."
    sleep 10
    docker ps
}

# ─── Status ──────────────────────────────────────────────────────────

show_status() {
    echo ""
    echo -e "${CYAN}╔══════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║       VortexUiPro — Deployment Status            ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════╝${NC}"
    echo ""

    # Git info
    cd "${VORTEX_PROJECT}" 2>/dev/null || true
    echo -e "${BLUE}Repository:${NC} $(git remote get-url origin 2>/dev/null || echo 'N/A')"
    echo -e "${BLUE}Branch:${NC}     $(git branch --show-current 2>/dev/null || echo 'N/A')"
    echo -e "${BLUE}Commit:${NC}     $(git rev-parse --short HEAD 2>/dev/null || echo 'N/A')"
    echo ""

    # Docker status
    echo -e "${BLUE}Running containers:${NC}"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "  (none running)"

    # Disk usage
    echo ""
    echo -e "${BLUE}Disk usage:${NC}"
    df -h "${VORTEX_PROJECT}" 2>/dev/null | tail -1

    # Backup status
    echo ""
    echo -e "${BLUE}Latest backups:${NC}"
    ls -lt "${VORTEX_PROJECT}/deploy/data/backups/db/" 2>/dev/null | head -3 || echo "  (no backups found)"
}

# ─── Main ─────────────────────────────────────────────────────────────

main() {
    echo -e "${CYAN}"
    echo '╔══════════════════════════════════════════════════╗'
    echo '║       VortexUiPro — Deployment Script           ║'
    echo '╚══════════════════════════════════════════════════╝'
    echo -e "${NC}"

    case "${1:-}" in
        --help|-h) usage ;;
        --rollback|-r) rollback; exit 0 ;;
        --status|-s) show_status; exit 0 ;;
        --cleanup|-c) cleanup; exit 0 ;;
        *)
            check_prereqs
            git_pull
            load_env
            pre_deploy_backup
            pull_images
            deploy_services
            wait_healthy
            cleanup

            echo ""
            echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
            echo -e "${GREEN}║       ✅ Deployment Complete!                    ║${NC}"
            echo -e "${GREEN}╠══════════════════════════════════════════════════╣${NC}"
            echo -e "${GREEN}║ Environment: ${VORTEX_ENV}${NC}"
            echo -e "${GREEN}║ Version:     ${VORTEX_VERSION}${NC}"
            echo -e "${GREEN}║ Project:     ${VORTEX_PROJECT}${NC}"
            echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
            echo ""
            show_status
            ;;
    esac
}

main "$@"
