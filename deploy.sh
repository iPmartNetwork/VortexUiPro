#!/usr/bin/env bash
# ─── VortexUiPro Deploy/Update Script ────────────────────────────────
# Used for updating an existing VortexUiPro installation.
# Usage: sudo ./deploy.sh [--skip-backend] [--skip-frontend] [--skip-migrate]
# ──────────────────────────────────────────────────────────────────────
set -euo pipefail

VERSION="0.0.1"
INSTALL_DIR="${VORTEX_REPO_DIR:-/opt/vortexuipro}"
WEB_ROOT="${VORTEX_WEB_ROOT:-/var/www/vortexuipro}"
SERVICE="vortexuipro-panel"
BRANCH="master"

SKIP_BACKEND=0; SKIP_FRONTEND=0; SKIP_MIGRATE=0

for arg in "$@"; do
    case "$arg" in
        --skip-backend)  SKIP_BACKEND=1 ;;
        --skip-frontend) SKIP_FRONTEND=1 ;;
        --skip-migrate)  SKIP_MIGRATE=1 ;;
    esac
done

[[ $EUID -ne 0 ]] && { echo "Must be run as root."; exit 1; }

cd "$INSTALL_DIR"

echo "==> VortexUiPro v${VERSION} update"
echo "==> git pull"
git pull origin "$BRANCH"

# Frontend
if [[ "$SKIP_FRONTEND" -eq 0 ]]; then
    echo "==> building frontend"
    cd "$INSTALL_DIR/web"
    npm ci --prefer-offline
    npm run build
    mkdir -p "$WEB_ROOT"
    cp -r dist/* "$WEB_ROOT/"
    cd "$INSTALL_DIR"
fi

# Backend
if [[ "$SKIP_BACKEND" -eq 0 ]]; then
    echo "==> building backend"
    CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" \
        -o /usr/local/bin/vortexuipro ./cmd/panel
fi

# Restart service
echo "==> restarting ${SERVICE}"
systemctl daemon-reload
systemctl restart "$SERVICE"

echo ""
echo "==> Done! VortexUiPro v${VERSION} updated."
