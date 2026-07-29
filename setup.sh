#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# VortexUiPro v0.0.1 — Unified Smart Deploy Script
# Detects Docker vs systemd mode and executes the appropriate deployment.
#
# Usage:
#   sudo ./setup.sh [--docker|--systemd] [--skip-backend] [--skip-frontend] [--skip-migrate]
#
# Mode detection priority:
#   1. Explicit flag (--docker or --systemd)
#   2. Running VortexUiPro Docker containers detected
#   3. Existing systemd service file found
#   4. Docker available → Docker mode (default for fresh install)
# ═══════════════════════════════════════════════════════════════════════
set -euo pipefail

VERSION="0.0.1"
REPO_URL="https://github.com/iPmartNetwork/VortexUiPro.git"
INSTALL_DIR="${VORTEX_REPO_DIR:-/opt/vortexuipro}"
WEB_ROOT="${VORTEX_WEB_ROOT:-/var/www/vortexuipro}"
SERVICE="vortexuipro-panel"
BRANCH="master"

MODE=""
SKIP_BACKEND=0
SKIP_FRONTEND=0
SKIP_MIGRATE=0

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; CYAN='\033[0;36m'; NC='\033[0m'

log()   { echo -e "${GREEN}[VortexUiPro]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }
pass()  { echo -e "  ${GREEN}[✓]${NC} $1"; }
fail()  { echo -e "  ${RED}[✗]${NC} $1"; }
info()  { echo -e "  ${BLUE}[•]${NC} $1"; }

header() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════╗"
    echo "║     VortexUiPro v${VERSION} — Smart Deploy           ║"
    echo "║   Next-Gen Proxy Management Panel               ║"
    echo "╚══════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# ─── Argument Parsing ────────────────────────────────────────────────
parse_args() {
    for arg in "$@"; do
        case "$arg" in
            --docker)        MODE="docker" ;;
            --systemd)       MODE="systemd" ;;
            --skip-backend)  SKIP_BACKEND=1 ;;
            --skip-frontend) SKIP_FRONTEND=1 ;;
            --skip-migrate)  SKIP_MIGRATE=1 ;;
            --help|-h)
                echo "Usage: sudo ./setup.sh [--docker|--systemd] [options]"
                echo ""
                echo "Modes:"
                echo "  --docker     Force Docker Compose deployment"
                echo "  --systemd    Force native systemd deployment"
                echo "  (omit)       Auto-detect based on environment"
                echo ""
                echo "Flags:"
                echo "  --skip-backend   Skip Go backend build"
                echo "  --skip-frontend  Skip frontend build"
                echo "  --skip-migrate   Skip database migrations"
                exit 0
                ;;
        esac
    done
}

# ─── Mode Detection ──────────────────────────────────────────────────
detect_mode() {
    [[ -n "$MODE" ]] && { log "Mode: $MODE (explicit)"; return; }

    if command -v docker &>/dev/null && docker ps --format '{{.Names}}' 2>/dev/null | grep -qi "vortex"; then
        MODE="docker"; log "Mode: docker (containers detected)"; return
    fi

    if systemctl list-unit-files "${SERVICE}.service" &>/dev/null 2>&1; then
        MODE="systemd"; log "Mode: systemd (service file found)"; return
    fi

    if command -v docker &>/dev/null && docker compose version &>/dev/null 2>&1; then
        MODE="docker"; log "Mode: docker (Docker available)"; return
    fi

    error "Cannot determine deployment mode. Use --docker or --systemd explicitly."
}

# ─── Doctor Check ────────────────────────────────────────────────────
doctor_check() {
    local phase="$1" failures=0
    echo ""; log "Running ${phase} health checks..."

    local free_kb; free_kb=$(df -k / | tail -1 | awk '{print $4}')
    if [[ "$free_kb" -lt 1048576 ]]; then
        fail "Disk space: $((free_kb/1024))MB free (need ≥1GB)"; ((failures++))
    else
        pass "Disk space: $((free_kb/1024))MB free"
    fi

    if command -v ss &>/dev/null; then
        local port_8080; port_8080=$(ss -tlnp 2>/dev/null | grep ":8080 " || true)
        [[ -n "$port_8080" ]] && warn "Port 8080: in use" || pass "Port 8080: available"
    fi

    if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
        pass "Docker daemon: running"
    elif [[ "$MODE" == "docker" ]]; then
        fail "Docker daemon: not running"; ((failures++))
    fi

    echo ""
    [[ "$failures" -gt 0 ]] && warn "${phase}: ${failures} issue(s)" || log "${phase}: all passed"
    return 0
}

# ─── Docker Mode ─────────────────────────────────────────────────────
check_root() { [[ $EUID -ne 0 ]] && error "Must be run as root: sudo ./setup.sh"; }

check_docker_requirements() {
    log "Checking system requirements..."
    local ARCH; ARCH=$(uname -m)
    case $ARCH in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        *)       error "Unsupported arch: $ARCH" ;;
    esac
    log "Architecture: $ARCH"

    ! command -v docker &>/dev/null && warn "Installing Docker..." && curl -fsSL https://get.docker.com | sh
    ! docker compose version &>/dev/null && error "Docker Compose v2 required"
    if ! command -v git &>/dev/null; then
        if command -v apt-get &>/dev/null; then
            apt-get update -qq && apt-get install -y -qq git
        elif command -v yum &>/dev/null; then
            yum install -y git
        elif command -v apk &>/dev/null; then
            apk add git
        fi
    fi
}

clone_or_pull() {
    if [[ -d "$INSTALL_DIR/.git" ]]; then
        log "Existing installation found. Updating..."
        cd "$INSTALL_DIR"; git fetch origin; git checkout "$BRANCH"; git pull origin "$BRANCH"
    else
        log "Fresh install — cloning repository..."
        git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$INSTALL_DIR"
        cd "$INSTALL_DIR"
    fi
    log "Repository ready at $INSTALL_DIR"
}

generate_secrets() {
    if [[ ! -f "$INSTALL_DIR/.env" ]]; then
        log "Generating configuration..."
        cp "$INSTALL_DIR/.env.example" "$INSTALL_DIR/.env"
        local jwt panel db_pass
        jwt=$(openssl rand -hex 32)
        panel=$(openssl rand -hex 32)
        db_pass=$(openssl rand -base64 16 | tr -d '=/+')
        sed -i "s|change-me-to-a-random-64-char-hex-string|${jwt}|" "$INSTALL_DIR/.env"
        log "Secrets generated"
    else
        log "Existing .env found"
    fi
}

generate_certs() {
    if [[ ! -d "$INSTALL_DIR/deploy/certs" ]]; then
        log "Generating mTLS certificates..."
        mkdir -p "$INSTALL_DIR/deploy/certs"
        openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:P-256 \
            -days 3650 -nodes -subj "/CN=VortexUiPro CA" \
            -keyout "$INSTALL_DIR/deploy/certs/ca.key" -out "$INSTALL_DIR/deploy/certs/ca.crt" 2>/dev/null
        openssl req -newkey ec -pkeyopt ec_paramgen_curve:P-256 -nodes \
            -subj "/CN=panel" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" \
            -keyout "$INSTALL_DIR/deploy/certs/panel.key" -out "$INSTALL_DIR/deploy/certs/panel.csr" 2>/dev/null
        openssl x509 -req -in "$INSTALL_DIR/deploy/certs/panel.csr" -CA "$INSTALL_DIR/deploy/certs/ca.crt" \
            -CAkey "$INSTALL_DIR/deploy/certs/ca.key" -CAcreateserial -days 3650 \
            -copy_extensions copyall -out "$INSTALL_DIR/deploy/certs/panel.crt" 2>/dev/null
        cp "$INSTALL_DIR/deploy/certs/panel.key" "$INSTALL_DIR/deploy/certs/node.key"
        cp "$INSTALL_DIR/deploy/certs/panel.crt" "$INSTALL_DIR/deploy/certs/node.crt"
        rm -f "$INSTALL_DIR/deploy/certs/panel.csr" "$INSTALL_DIR/deploy/certs/ca.srl"
        log "Certificates generated"
    fi
}

start_docker_stack() {
    log "Starting VortexUiPro stack..."
    cd "$INSTALL_DIR"
    docker compose -f deploy/production.yml up --build -d

    log "Waiting for services..."
    for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do
        curl -sf http://localhost:8080/api/v1/health &>/dev/null && break
        sleep 2
    done
}

deploy_docker() {
    check_docker_requirements
    clone_or_pull
    generate_secrets
    generate_certs
    start_docker_stack
}

# ─── Systemd Mode ────────────────────────────────────────────────────
deploy_systemd() {
    cd "$INSTALL_DIR"
    log "Systemd deployment starting..."

    # Build frontend
    if [[ "$SKIP_FRONTEND" -eq 0 ]]; then
        log "Building frontend..."
        cd "$INSTALL_DIR/web"
        npm ci --prefer-offline
        npm run build
        mkdir -p "$WEB_ROOT"
        cp -r dist/* "$WEB_ROOT/"
        cd "$INSTALL_DIR"
    fi

    # Build backend
    if [[ "$SKIP_BACKEND" -eq 0 ]]; then
        log "Building backend..."
        CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" \
            -o /usr/local/bin/vortexuipro ./cmd/panel
    fi

    # Install systemd service
    log "Installing systemd service..."
    cp "$INSTALL_DIR/deploy/vortexuipro-panel.service" /etc/systemd/system/
    systemctl daemon-reload
    systemctl enable "$SERVICE"
    systemctl restart "$SERVICE"
}

print_success() {
    local ip; ip=$(curl -sf https://api.ipify.org || hostname -I | awk '{print $1}')
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   VortexUiPro v${VERSION} installed!                ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${BLUE}Panel:${NC}   http://${ip}:8080"
    echo -e "  ${BLUE}Install:${NC} ${INSTALL_DIR}"
    [[ "$MODE" == "systemd" ]] && echo -e "  ${BLUE}Service:${NC} systemctl status ${SERVICE}"
    echo ""
}

# ─── Main ────────────────────────────────────────────────────────────
parse_args "$@"
header
check_root
detect_mode
doctor_check "Pre-flight"

case "$MODE" in
    docker)
        deploy_docker
        doctor_check "Post-deploy"
        print_success
        ;;
    systemd)
        deploy_systemd
        doctor_check "Post-deploy"
        print_success
        ;;
esac
