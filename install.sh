#!/bin/bash

set -e

# ═══════════════════════════════════════════════════════════════════════
# VortexUiPro — Install Script
# Version: 0.0.1
# Repo:    https://github.com/iPmartNetwork/VortexUiPro
#
# Usage:
#   bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh)
#
# Or with options:
#   bash <(curl -fsSL ...) -- --port 8080 --username admin --password mypass
# ═══════════════════════════════════════════════════════════════════════

red='\033[0;31m'
green='\033[0;32m'
blue='\033[0;34m'
yellow='\033[0;33m'
cyan='\033[0;36m'
plain='\033[0m'

# ─── Defaults ──────────────────────────────────────────────────────────
VORTEX_REPO="iPmartNetwork/VortexUiPro"
VORTEX_VERSION="0.1.1"
VORTEX_FOLDER="/usr/local/vortexuipro"
VORTEX_DATA="/etc/vortexuipro"
VORTEX_BIN="/usr/local/bin/vortexuipro"
VORTEX_SERVICE="/etc/systemd/system/vortexuipro.service"
VORTEX_USER="vortexuipro"
VORTEX_GROUP="vortexuipro"

# ─── Banner ─────────────────────────────────────────────────────────────
echo -e ""
echo -e "${cyan}╔══════════════════════════════════════════════════════════╗${plain}"
echo -e "${cyan}║${plain}                                                      ${cyan}║${plain}"
echo -e "${cyan}║${plain}  ${green}██╗   ██╗ ██████╗ ██████╗ ████████╗███████╗██╗  ██╗${plain}  ${cyan}║${plain}"
echo -e "${cyan}║${plain}  ${green}██║   ██║██╔═══██╗██╔══██╗╚══██╔══╝██╔════╝╚██╗██╔╝${plain}  ${cyan}║${plain}"
echo -e "${cyan}║${plain}  ${green}██║   ██║██║   ██║██████╔╝   ██║   █████╗   ╚███╔╝ ${plain}  ${cyan}║${plain}"
echo -e "${cyan}║${plain}  ${green}╚██╗ ██╔╝██║   ██║██╔══██╗   ██║   ██╔══╝   ██╔██╗ ${plain}  ${cyan}║${plain}"
echo -e "${cyan}║${plain}  ${green} ╚████╔╝ ╚██████╔╝██║  ██║   ██║   ███████╗██╔╝ ██╗${plain}  ${cyan}║${plain}"
echo -e "${cyan}║${plain}  ${green}  ╚═══╝   ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝${plain}  ${cyan}║${plain}"
echo -e "${cyan}║${plain}                                                      ${cyan}║${plain}"
echo -e "${cyan}║${plain}              ${yellow}Ultimate Proxy Management Panel${plain}        ${cyan}║${plain}"
echo -e "${cyan}║${plain}              ${blue}v${VORTEX_VERSION} — The Next Generation${plain}           ${cyan}║${plain}"
echo -e "${cyan}╚══════════════════════════════════════════════════════════╝${plain}"
echo -e ""

# ─── Root Check ─────────────────────────────────────────────────────────
[[ $EUID -ne 0 ]] && echo -e "${red}Error: Please run as root (use sudo).${plain}" && exit 1

# ─── OS Detection ───────────────────────────────────────────────────────
if [[ -f /etc/os-release ]]; then
    source /etc/os-release
    OS=$ID
    OS_VERSION=$VERSION_ID
else
    echo -e "${red}Failed to detect OS.${plain}"
    exit 1
fi

echo -e "${green}✓ OS:${plain} $OS $OS_VERSION"
echo -e "${green}✓ Arch:${plain} $(uname -m)"

# ─── Architecture Detection ─────────────────────────────────────────────
detect_arch() {
    case "$(uname -m)" in
        x86_64 | x64 | amd64) echo 'amd64' ;;
        i*86 | x86) echo '386' ;;
        armv8* | armv8 | arm64 | aarch64) echo 'arm64' ;;
        armv7* | armv7 | arm) echo 'armv7' ;;
        armv6* | armv6) echo 'armv6' ;;
        *) echo -e "${red}Unsupported architecture: $(uname -m)${plain}" && exit 1 ;;
    esac
}

ARCH=$(detect_arch)
echo -e "${green}✓ Target:${plain} vortexuipro-linux-${ARCH}"

# ─── Package Installation ───────────────────────────────────────────────
install_deps() {
    echo -e "\n${yellow}[1/7] Installing system dependencies...${plain}"

    # Install git for source builds (fallback), plus essential tools
    case "${OS}" in
        ubuntu | debian | armbian | linuxmint | parrot | kali)
            apt-get update -qq
            apt-get install -y -qq curl wget tar gzip git socat ca-certificates openssl systemd
            ;;
        fedora | amzn | rhel | almalinux | rocky | centos | ol)
            if command -v dnf &>/dev/null; then
                dnf install -y -q curl wget tar gzip git socat ca-certificates openssl systemd
            else
                yum install -y curl wget tar gzip git socat ca-certificates openssl systemd
            fi
            ;;
        arch | manjaro | endeavouros | arcolinux)
            pacman -Sy --noconfirm curl wget tar gzip git socat ca-certificates openssl systemd
            ;;
        opensuse* | suse)
            zypper -q install -y curl wget tar gzip git socat ca-certificates openssl systemd
            ;;
        alpine)
            apk add --no-cache curl wget tar gzip git socat ca-certificates openssl
            ;;
        *)
            echo -e "${yellow}Warning: Unsupported OS. Trying apt-get...${plain}"
            apt-get update -qq && apt-get install -y -qq curl wget tar gzip git socat ca-certificates openssl
            ;;
    esac

    echo -e "${green}✓ System dependencies installed${plain}"
}

# ─── Create User ────────────────────────────────────────────────────────
create_user() {
    echo -e "\n${yellow}[2/7] Creating system user...${plain}"

    if ! id -u ${VORTEX_USER} &>/dev/null; then
        useradd -r -s /sbin/nologin -d ${VORTEX_DATA} -M ${VORTEX_USER}
        echo -e "${green}✓ User '${VORTEX_USER}' created${plain}"
    else
        echo -e "${green}✓ User '${VORTEX_USER}' already exists${plain}"
    fi
}

# ─── Directory Structure ────────────────────────────────────────────────
create_dirs() {
    echo -e "\n${yellow}[3/7] Creating directory structure...${plain}"

    mkdir -p ${VORTEX_DATA}/{data,certs,backups,mtproto,plugins,logs}
    mkdir -p ${VORTEX_FOLDER}
    mkdir -p ${VORTEX_DATA}/pki

    echo -e "${green}✓ Directories created${plain}"
    echo -e "  ${blue}Data:${plain}     ${VORTEX_DATA}"
    echo -e "  ${blue}Binary:${plain}   ${VORTEX_FOLDER}"
    echo -e "  ${blue}Certs:${plain}    ${VORTEX_DATA}/certs"
    echo -e "  ${blue}Backups:${plain}  ${VORTEX_DATA}/backups"
}

# ─── Download Binary ────────────────────────────────────────────────────
download_binary() {
    echo -e "\n${yellow}[4/7] Downloading VortexUiPro binary...${plain}"

    local DOWNLOAD_URL="https://github.com/${VORTEX_REPO}/releases/download/v${VORTEX_VERSION}/vortexuipro-linux-${ARCH}.tar.gz"

    echo -e "  ${blue}URL:${plain} ${DOWNLOAD_URL}"

    # Try release, fallback to build from source
    if curl -fsSL --connect-timeout 10 "${DOWNLOAD_URL}" -o /tmp/vortexuipro.tar.gz; then
        tar -xzf /tmp/vortexuipro.tar.gz -C /tmp/
        if [[ -f /tmp/vortexuipro ]]; then
            mv /tmp/vortexuipro ${VORTEX_BIN}
        elif [[ -f /tmp/vortexuipro-linux-${ARCH} ]]; then
            mv /tmp/vortexuipro-linux-${ARCH} ${VORTEX_BIN}
        else
            echo -e "${yellow}Binary not found in archive, building from source...${plain}"
            build_from_source
        fi
        rm -f /tmp/vortexuipro.tar.gz
    else
        echo -e "${yellow}Release binary not found, building from source...${plain}"
        build_from_source
    fi

    chmod +x ${VORTEX_BIN}
    echo -e "${green}✓ Binary installed: ${VORTEX_BIN}${plain}"
}

# ─── Build from Source ──────────────────────────────────────────────────
build_from_source() {
    echo -e "${yellow}Building from source (this may take a while)...${plain}"

    # Check Go version — need 1.21+
    local GO_OK=0
    if command -v go &>/dev/null; then
        local GO_VER
        GO_VER=$(go version | awk '{print $3}' | sed 's/go//')
        local GO_MAJOR GO_MINOR
        GO_MAJOR=$(echo "$GO_VER" | cut -d. -f1)
        GO_MINOR=$(echo "$GO_VER" | cut -d. -f2)
        if [[ "$GO_MAJOR" -gt 1 ]] || [[ "$GO_MAJOR" -eq 1 && "$GO_MINOR" -ge 21 ]]; then
            GO_OK=1
        else
            echo -e "${yellow}Go ${GO_VER} is too old (need 1.21+), installing newer Go...${plain}"
        fi
    fi

    if [[ "$GO_OK" -eq 0 ]]; then
        local GO_INSTALL_VERSION="1.23.0"
        local GO_ARCH
        case "$(uname -m)" in
            x86_64) GO_ARCH="amd64" ;;
            aarch64|arm64) GO_ARCH="arm64" ;;
            *) GO_ARCH="amd64" ;;
        esac
        echo -e "${yellow}Installing Go ${GO_INSTALL_VERSION}...${plain}"
        curl -fsSL "https://go.dev/dl/go${GO_INSTALL_VERSION}.linux-${GO_ARCH}.tar.gz" -o /tmp/go.tar.gz
        rm -rf /usr/local/go
        tar -C /usr/local -xzf /tmp/go.tar.gz
        rm -f /tmp/go.tar.gz
        export PATH=$PATH:/usr/local/go/bin
        echo -e "${green}✓ Go $(go version | awk '{print $3}') installed${plain}"
    fi

    local TMPDIR=$(mktemp -d)
    cd ${TMPDIR}

    echo -e "${yellow}Cloning repository...${plain}"
    git clone --depth 1 https://github.com/${VORTEX_REPO}.git .
    
    echo -e "${yellow}Building backend...${plain}"
    go build -o vortexuipro -ldflags="-s -w" ./cmd/panel

    cp vortexuipro ${VORTEX_BIN}
    cd /
    rm -rf ${TMPDIR}
    
    echo -e "${green}✓ Build complete${plain}"
}

# ─── Create Config ──────────────────────────────────────────────────────
create_config() {
    echo -e "\n${yellow}[5/7] Creating configuration...${plain}"

    local JWT_SECRET=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 32)
    local PANEL_PORT="${VORTEX_PORT:-8080}"
    local GRPC_PORT="${VORTEX_GRPC_PORT:-50051}"

    cat > ${VORTEX_DATA}/env << EOF
# VortexUiPro Configuration
# Generated by install.sh on $(date)

# Server
VORTEX_HTTP_ADDR=:${PANEL_PORT}
VORTEX_GRPC_ADDR=:${GRPC_PORT}
VORTEX_JWT_SECRET=${JWT_SECRET}
VORTEX_LOG_LEVEL=info

# Database (SQLite default)
VORTEX_DB_TYPE=sqlite
VORTEX_DATABASE_URL=${VORTEX_DATA}/data/vortex.db

# Core Engine
VORTEX_CORE_BIN=/usr/local/bin/xray
VORTEX_CORE_CONFIG=${VORTEX_DATA}/data/xray.json
VORTEX_CORE_API_PORT=10085

# Activity Tracking
VORTEX_ACTIVITY_FLUSH_SEC=30

# Telegram Bot (optional — set your token)
# VORTEX_TELEGRAM_BOT_TOKEN=

# Cluster (optional — uncomment to enable)
# VORTEX_CLUSTER_ENABLED=true
# VORTEX_CLUSTER_NODE_NAME=node-1
# VORTEX_CLUSTER_ADDR=:1337
# VORTEX_CLUSTER_PEERS=
# VORTEX_CLUSTER_REGION=default
# VORTEX_CLUSTER_PRIORITY=100

# Plugin Directory
VORTEX_PLUGIN_DIR=${VORTEX_DATA}/plugins
EOF

    chmod 600 ${VORTEX_DATA}/env
    chown ${VORTEX_USER}:${VORTEX_GROUP} ${VORTEX_DATA}/env

    echo -e "${green}✓ Configuration created at ${VORTEX_DATA}/env${plain}"
    echo -e "  ${blue}Port:${plain}     ${PANEL_PORT}"
    echo -e "  ${blue}gRPC:${plain}    ${GRPC_PORT}"
}

# ─── Create Systemd Service ─────────────────────────────────────────────
create_service() {
    echo -e "\n${yellow}[6/7] Creating systemd service...${plain}"

    cat > ${VORTEX_SERVICE} << EOF
[Unit]
Description=VortexUiPro — Ultimate Proxy Management Panel
Documentation=https://github.com/${VORTEX_REPO}
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=${VORTEX_DATA}/env
ExecStart=${VORTEX_BIN}
ExecReload=/bin/kill -HUP \$MAINPID
WorkingDirectory=${VORTEX_DATA}

User=${VORTEX_USER}
Group=${VORTEX_GROUP}

Restart=on-failure
RestartSec=5
StartLimitIntervalSec=60
StartLimitBurst=5

NoNewPrivileges=true
ProtectSystem=full
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/etc/vortexuipro
CapabilityBoundingSet=CAP_NET_BIND_SERVICE CAP_NET_ADMIN CAP_SYS_PTRACE
AmbientCapabilities=CAP_NET_BIND_SERVICE

LimitNOFILE=65535
LimitNPROC=4096

TimeoutStartSec=30
TimeoutStopSec=30

StandardOutput=journal
StandardError=journal
SyslogIdentifier=vortexuipro

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    echo -e "${green}✓ Service created: vortexuipro.service${plain}"
}

# ─── Start Service ──────────────────────────────────────────────────────
start_service() {
    echo -e "\n${yellow}[7/7] Starting VortexUiPro...${plain}"

    chown -R ${VORTEX_USER}:${VORTEX_GROUP} ${VORTEX_DATA}
    chown root:root ${VORTEX_BIN}
    chmod 755 ${VORTEX_BIN}

    systemctl enable vortexuipro
    systemctl start vortexuipro

    sleep 2

    if systemctl is-active --quiet vortexuipro; then
        echo -e "${green}✓ VortexUiPro service started successfully!${plain}"
    else
        echo -e "${red}✗ Service failed to start. Check logs: journalctl -u vortexuipro -f${plain}"
        exit 1
    fi
}

# ─── SSL Setup (optional) ───────────────────────────────────────────────
setup_ssl() {
    echo -e "\n${yellow}[Optional] SSL Certificate Setup${plain}"
    echo -e "  ${blue}1.${plain} Let's Encrypt (Domain) — 90-day, auto-renew"
    echo -e "  ${blue}2.${plain} Let's Encrypt (IP)      — 6-day, auto-renew"
    echo -e "  ${blue}3.${plain} Custom SSL Certificate"
    echo -e "  ${blue}4.${plain} Skip (use HTTP)"

    local SSL_CHOICE
    if [[ -n "${VORTEX_SKIP_SSL}" ]] && [[ "${VORTEX_SKIP_SSL}" == "true" ]]; then
        SSL_CHOICE="4"
    elif [[ -n "${VORTEX_SSL_MODE}" ]]; then
        case "${VORTEX_SSL_MODE}" in
            domain) SSL_CHOICE="1" ;;
            ip) SSL_CHOICE="2" ;;
            custom) SSL_CHOICE="3" ;;
            *) SSL_CHOICE="4" ;;
        esac
    else
        read -rp "Choose SSL option (default: 4): " SSL_CHOICE
        SSL_CHOICE="${SSL_CHOICE:-4}"
    fi

    case "${SSL_CHOICE}" in
        1)
            echo -e "${green}Setting up Let's Encrypt for domain...${plain}"
            if ! command -v ~/.acme.sh/acme.sh &>/dev/null; then
                curl -fsSL https://get.acme.sh | sh
            fi

            local DOMAIN=""
            if [[ -n "${VORTEX_SSL_DOMAIN}" ]]; then
                DOMAIN="${VORTEX_SSL_DOMAIN}"
            else
                read -rp "Enter your domain: " DOMAIN
            fi

            ~/.acme.sh/acme.sh --issue -d ${DOMAIN} --standalone --force
            ~/.acme.sh/acme.sh --installcert -d ${DOMAIN} \
                --key-file ${VORTEX_DATA}/certs/privkey.pem \
                --fullchain-file ${VORTEX_DATA}/certs/fullchain.pem \
                --reloadcmd "systemctl restart vortexuipro"

            echo -e "${green}✓ SSL certificate installed for ${DOMAIN}${plain}"
            echo -e "  ${blue}Access:${plain} https://${DOMAIN}:${VORTEX_PORT:-8080}"
            ;;
        2)
            echo -e "${green}Setting up Let's Encrypt for IP...${plain}"
            local SERVER_IP=$(curl -fsSL https://api.ipify.org 2>/dev/null || curl -fsSL https://ipv4.icanhazip.com 2>/dev/null)
            
            if ! command -v ~/.acme.sh/acme.sh &>/dev/null; then
                curl -fsSL https://get.acme.sh | sh
            fi

            ~/.acme.sh/acme.sh --issue -d ${SERVER_IP} --standalone --force \
                --certificate-profile shortlived --days 6
            ~/.acme.sh/acme.sh --installcert -d ${SERVER_IP} \
                --key-file ${VORTEX_DATA}/certs/privkey.pem \
                --fullchain-file ${VORTEX_DATA}/certs/fullchain.pem \
                --reloadcmd "systemctl restart vortexuipro"

            echo -e "${green}✓ SSL certificate installed for ${SERVER_IP}${plain}"
            echo -e "  ${blue}Access:${plain} https://${SERVER_IP}:${VORTEX_PORT:-8080}"
            ;;
        3)
            echo -e "${green}Custom SSL certificate...${plain}"
            local CERT_PATH KEY_PATH
            read -rp "Full certificate path: " CERT_PATH
            read -rp "Private key path: " KEY_PATH

            if [[ -f "${CERT_PATH}" && -f "${KEY_PATH}" ]]; then
                cp "${CERT_PATH}" ${VORTEX_DATA}/certs/fullchain.pem
                cp "${KEY_PATH}" ${VORTEX_DATA}/certs/privkey.pem
                echo -e "${green}✓ Custom certificate installed${plain}"
            else
                echo -e "${red}Certificate files not found${plain}"
            fi
            ;;
        *)
            echo -e "${yellow}SSL skipped. Panel will use HTTP.${plain}"
            echo -e "${yellow}For production, use a reverse proxy.${plain}"
            ;;
    esac
}

# ─── Docker Installation ────────────────────────────────────────────────
install_docker() {
    echo -e "\n${yellow}[Docker Mode] Setting up with Docker...${plain}"

    if ! command -v docker &>/dev/null; then
        echo -e "${yellow}Installing Docker...${plain}"
        curl -fsSL https://get.docker.com | sh
    fi

    local TMPDIR=$(mktemp -d)
    cd ${TMPDIR}

    echo -e "${yellow}Downloading Docker Compose configuration...${plain}"
    curl -fsSL "https://raw.githubusercontent.com/${VORTEX_REPO}/master/deploy/compose.yml" -o docker-compose.yml
    curl -fsSL "https://raw.githubusercontent.com/${VORTEX_REPO}/master/deploy/Caddyfile" -o Caddyfile
    curl -fsSL "https://raw.githubusercontent.com/${VORTEX_REPO}/master/deploy/Dockerfile" -o Dockerfile
    curl -fsSL "https://raw.githubusercontent.com/${VORTEX_REPO}/master/deploy/web.Dockerfile" -o web.Dockerfile

    mkdir -p data/node-1 data/node-2 data/node-3

    echo -e "${green}✓ Files downloaded${plain}"
    echo -e ""
    echo -e "${yellow}To start:${plain}"
    echo -e "  cd ${TMPDIR}"
    echo -e "  docker compose up -d"
    echo -e ""
    echo -e "${yellow}Nodes:${plain}"
    echo -e "  Node-1: http://localhost:8080"
    echo -e "  Node-2: http://localhost:8081"
    echo -e "  Node-3: http://localhost:8082"
    echo -e "  Default login: admin / admin123"
}

# ─── Show Result ────────────────────────────────────────────────────────
show_result() {
    local SERVER_IP=$(curl -fsSL https://api.ipify.org 2>/dev/null || echo "YOUR_SERVER_IP")
    local PANEL_PORT="${VORTEX_PORT:-8080}"

    echo -e ""
    echo -e "${cyan}╔══════════════════════════════════════════════════════════╗${plain}"
    echo -e "${cyan}║${plain}        ${green}VortexUiPro Installation Complete!${plain}          ${cyan}║${plain}"
    echo -e "${cyan}╚══════════════════════════════════════════════════════════╝${plain}"
    echo -e ""
    echo -e "  ${green}✓ Version:${plain}    ${VORTEX_VERSION}"
    echo -e "  ${green}✓ Binary:${plain}      ${VORTEX_BIN}"
    echo -e "  ${green}✓ Config:${plain}      ${VORTEX_DATA}/env"
    echo -e "  ${green}✓ Service:${plain}     vortexuipro"
    echo -e ""
    echo -e "  ${yellow}Access URL:${plain}"
    echo -e "  ${blue}  http://${SERVER_IP}:${PANEL_PORT}${plain}"
    echo -e ""
    echo -e "  ${yellow}Default Login:${plain}"
    echo -e "  ${blue}  Username: admin${plain}"
    echo -e "  ${blue}  Password: admin123${plain}"
    echo -e ""
    echo -e "  ${yellow}Management:${plain}"
    echo -e "  ${blue}  Status:  systemctl status vortexuipro${plain}"
    echo -e "  ${blue}  Logs:    journalctl -u vortexuipro -f${plain}"
    echo -e "  ${blue}  Restart: systemctl restart vortexuipro${plain}"
    echo -e ""
    echo -e "  ${yellow}⚠ IMPORTANT:${plain}"
    echo -e "  ${yellow}  • Change the default password immediately!${plain}"
    echo -e "  ${yellow}  • Configure a reverse proxy for HTTPS in production${plain}"
    echo -e "  ${yellow}  • Install Xray core: https://github.com/XTLS/Xray-install${plain}"
    echo -e ""
    echo -e "${cyan}────────────────────────────────────────────────────────${plain}"
    echo -e "${green}Thank you for choosing VortexUiPro!${plain}"
    echo -e "${cyan}────────────────────────────────────────────────────────${plain}"
    echo -e ""
}

# ─── Create Symlink for vortexui Command ───────────────────────────────
create_symlink() {
    local SCRIPT_PATH="$(pwd)/vortexui.sh"
    if [[ -f "${VORTEX_FOLDER}/vortexui.sh" ]]; then
        ln -sf "${VORTEX_FOLDER}/vortexui.sh" /usr/local/bin/vortexui
        chmod +x /usr/local/bin/vortexui
        echo -e "${green}✓ Command 'vortexui' available (symlink)${plain}"
    fi
}

# ─── Uninstall ──────────────────────────────────────────────────────────
uninstall() {
    echo -e "\n${red}Uninstalling VortexUiPro...${plain}"

    systemctl stop vortexuipro 2>/dev/null || true
    systemctl disable vortexuipro 2>/dev/null || true
    rm -f ${VORTEX_SERVICE}
    rm -f ${VORTEX_BIN}
    rm -f /usr/local/bin/vortexui
    rm -rf ${VORTEX_FOLDER}
    systemctl daemon-reload

    local REMOVE_DATA="${VORTEX_REMOVE_DATA:-no}"
    if [[ "${REMOVE_DATA}" == "yes" ]]; then
        rm -rf ${VORTEX_DATA}
        userdel ${VORTEX_USER} 2>/dev/null || true
        echo -e "${red}All data removed.${plain}"
    else
        echo -e "${yellow}Data preserved at ${VORTEX_DATA}${plain}"
        echo -e "${yellow}Remove manually: rm -rf ${VORTEX_DATA}${plain}"
    fi

    echo -e "${green}✓ VortexUiPro uninstalled${plain}"
}

# ─── Main ───────────────────────────────────────────────────────────────
main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --port) VORTEX_PORT="$2"; shift 2 ;;
            --grpc-port) VORTEX_GRPC_PORT="$2"; shift 2 ;;
            # --username and --password are parsed but unused;
            # default credentials (admin/admin123) are set at first start.
            --ssl-domain) VORTEX_SSL_DOMAIN="$2"; VORTEX_SSL_MODE="domain"; shift 2 ;;
            --ssl-ip) VORTEX_SSL_MODE="ip"; shift ;;
            --skip-ssl) VORTEX_SKIP_SSL="true"; shift ;;
            --docker) VORTEX_DOCKER="true"; shift ;;
            --uninstall) uninstall; exit 0 ;;
            --help)
                echo "Usage: bash install.sh [OPTIONS]"
                echo "  --port PORT         Panel HTTP port (default: 8080)"
                echo "  --grpc-port PORT    gRPC port (default: 50051)"
                echo "  # Default admin credentials: admin / admin123"
                echo "  --ssl-domain DOM    Enable SSL with Let's Encrypt domain"
                echo "  --ssl-ip            Enable SSL with Let's Encrypt IP"
                echo "  --skip-ssl          Skip SSL setup"
                echo "  --docker            Install using Docker"
                echo "  --uninstall         Remove VortexUiPro"
                echo "  --help              Show this help"
                exit 0
                ;;
            *) shift ;;
        esac
    done

    echo -e "${blue}Starting installation in 3 seconds...${plain}"
    sleep 1

    if [[ "${VORTEX_DOCKER}" == "true" ]]; then
        install_docker
        show_result
        exit 0
    fi

    install_deps
    create_user
    create_dirs
    # Copy vortexui.sh to VORTEX_FOLDER so the symlink works
    cp "$0" "${VORTEX_FOLDER}/install.sh" 2>/dev/null || true
    local SCRIPT_DIR
    SCRIPT_DIR="$(cd "$(dirname "$0")" &>/dev/null && pwd)"
    if [[ -f "${SCRIPT_DIR}/vortexui.sh" ]]; then
        cp "${SCRIPT_DIR}/vortexui.sh" "${VORTEX_FOLDER}/vortexui.sh"
        chmod +x "${VORTEX_FOLDER}/vortexui.sh"
    fi
    download_binary
    create_config
    create_service
    start_service
    create_symlink
    setup_ssl
    show_result
}

main "$@"
