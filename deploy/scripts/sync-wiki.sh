#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# VortexUiPro — Wiki Sync Script 📖
#
# Syncs the docs/ folder to the GitHub Wiki repository.
# Run this after updating any documentation in docs/.
#
# Usage:
#   bash deploy/scripts/sync-wiki.sh                    # Interactive
#   bash deploy/scripts/sync-wiki.sh --dry-run          # Preview only
#   bash deploy/scripts/sync-wiki.sh --force            # Skip confirmation
#
# ═══════════════════════════════════════════════════════════════════════

set -euo pipefail

# ─── Configuration ─────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DOCS_DIR="${PROJECT_DIR}/docs"
WIKI_DIR="/tmp/vortexuipro-wiki"

GITHUB_OWNER="${GITHUB_OWNER:-iPmartNetwork}"
GITHUB_REPO="${GITHUB_REPO:-VortexUiPro}"
WIKI_REPO="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}.wiki.git"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ─── Functions ────────────────────────────────────────────────────
log_info()  { echo -e "${CYAN}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Sync docs/ to GitHub Wiki repository.

Options:
  --dry-run    Only show what would be done, don't actually sync
  --force      Skip confirmation prompt
  --help       Show this help message

Environment variables:
  GITHUB_OWNER  GitHub owner (default: iPmartNetwork)
  GITHUB_REPO   GitHub repo name (default: VortexUiPro)
  GH_TOKEN      GitHub token for authentication (recommended)

Examples:
  bash deploy/scripts/sync-wiki.sh
  bash deploy/scripts/sync-wiki.sh --dry-run
  GH_TOKEN=ghp_xxx bash deploy/scripts/sync-wiki.sh
EOF
    exit 0
}

# ─── Parse arguments ──────────────────────────────────────────────
DRY_RUN=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true; shift ;;
        --force)   FORCE=true; shift ;;
        --help)    usage ;;
        *) log_error "Unknown option: $1"; usage ;;
    esac
done

# ─── Main ─────────────────────────────────────────────────────────
echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║     VortexUiPro — Wiki Sync Script 📖           ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════╝${NC}"
echo ""

# Check if docs directory exists
if [[ ! -d "${DOCS_DIR}" ]]; then
    log_error "Docs directory not found: ${DOCS_DIR}"
    exit 1
fi

# Show what will be synced
log_info "Source: ${DOCS_DIR}"
log_info "Target: ${WIKI_REPO}"
log_info "Files to sync:"
find "${DOCS_DIR}" -name '*.md' -maxdepth 1 | while read -r f; do
    filename=$(basename "$f")
    size=$(wc -c < "$f" | tr -d ' ')
    echo "  📄 ${filename} (${size} bytes)"
done
echo ""

if [[ "${DRY_RUN}" == "true" ]]; then
    log_info "Dry-run mode — no changes will be made."
    exit 0
fi

# Confirm
if [[ "${FORCE}" != "true" ]]; then
    echo -e "${YELLOW}This will sync ${DOCS_DIR} to the GitHub Wiki repository.${NC}"
    read -rp "Continue? [y/N] " confirm
    if [[ "${confirm}" != "y" && "${confirm}" != "Y" ]]; then
        log_info "Aborted."
        exit 0
    fi
fi

# Clone wiki repo
log_info "Cloning wiki repository..."
if [[ -d "${WIKI_DIR}" ]]; then
    rm -rf "${WIKI_DIR}"
fi

if [[ -n "${GH_TOKEN:-}" ]]; then
    git clone "https://${GH_TOKEN}@github.com/${GITHUB_OWNER}/${GITHUB_REPO}.wiki.git" "${WIKI_DIR}"
else
    git clone "https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}.wiki.git" "${WIKI_DIR}"
fi

log_ok "Wiki repository cloned to ${WIKI_DIR}"

# Copy all markdown files from docs/ to wiki
log_info "Copying documentation files..."
copy_count=0
for f in "${DOCS_DIR}"/*.md; do
    filename=$(basename "$f")
    cp "$f" "${WIKI_DIR}/${filename}"
    log_ok "  Copied: ${filename}"
    copy_count=$((copy_count + 1))
done

# Check if there are changes
cd "${WIKI_DIR}"
if git diff --quiet && git diff --cached --quiet && [[ -z "$(git status --porcelain)" ]]; then
    log_warn "No changes to sync — wiki is up to date."
    rm -rf "${WIKI_DIR}"
    exit 0
fi

# Show changes
log_info "Changes to be pushed:"
git status --short

# Commit and push
log_info "Committing and pushing to wiki..."
git add -A
git commit -m "📖 Wiki sync: $(date +'%Y-%m-%d %H:%M:%S')

Auto-synced from docs/ folder.

Files synced:
$(find "${DOCS_DIR}" -name '*.md' -maxdepth 1 -exec basename {} \; | sed 's/^/- /')"

git push origin master

log_ok "✅ Wiki synced successfully!"
log_info "Visit: https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/wiki"

# Cleanup
rm -rf "${WIKI_DIR}"
echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  ✅ Wiki sync complete!                         ║${NC}"
echo -e "${GREEN}║  📖 https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/wiki  ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
