#!/bin/bash

# ═══════════════════════════════════════════════════════════════════════
# VortexUiPro — Management Script
# Version: 0.0.1
# Usage:   vortexui {start|stop|restart|status|logs|update|backup|restore}
# ═══════════════════════════════════════════════════════════════════════

red='\033[0;31m'
green='\033[0;32m'
yellow='\033[0;33m'
blue='\033[0;34m'
cyan='\033[0;36m'
plain='\033[0m'

VORTEX_BIN="/usr/local/bin/vortexuipro"
VORTEX_DATA="/etc/vortexuipro"
VORTEX_SERVICE="vortexuipro"

# ─── Help ───────────────────────────────────────────────────────────────
show_help() {
    echo -e "${cyan}VortexUiPro Management Script${plain}"
    echo -e ""
    echo -e "  ${green}Usage:${plain} vortexui {command} [options]"
    echo -e ""
    echo -e "  ${yellow}Commands:${plain}"
    echo -e "    ${green}start${plain}      Start the panel"
    echo -e "    ${green}stop${plain}       Stop the panel"
    echo -e "    ${green}restart${plain}    Restart the panel"
    echo -e "    ${green}status${plain}     Show panel status"
    echo -e "    ${green}logs${plain}       Show panel logs (use -f to follow)"
    echo -e "    ${green}update${plain}     Update VortexUiPro to latest version"
    echo -e "    ${green}backup${plain}     Create a backup of all data"
    echo -e "    ${green}restore${plain}    Restore from a backup file"
    echo -e "    ${green}password${plain}   Reset admin password"
    echo -e "    ${green}port${plain}       Change panel port"
    echo -e "    ${green}cert${plain}       Configure SSL certificate"
    echo -e "    ${green}info${plain}       Show installation information"
    echo -e "    ${green}version${plain}    Show version"
    echo -e "    ${green}help${plain}       Show this help"
}

# ─── Start ──────────────────────────────────────────────────────────────
cmd_start() {
    echo -e "${yellow}Starting VortexUiPro...${plain}"
    systemctl start ${VORTEX_SERVICE}
    if systemctl is-active --quiet ${VORTEX_SERVICE}; then
        echo -e "${green}✓ VortexUiPro started${plain}"
    else
        echo -e "${red}✗ Failed to start${plain}"
        exit 1
    fi
}

# ─── Stop ───────────────────────────────────────────────────────────────
cmd_stop() {
    echo -e "${yellow}Stopping VortexUiPro...${plain}"
    systemctl stop ${VORTEX_SERVICE}
    echo -e "${green}✓ VortexUiPro stopped${plain}"
}

# ─── Restart ────────────────────────────────────────────────────────────
cmd_restart() {
    echo -e "${yellow}Restarting VortexUiPro...${plain}"
    systemctl restart ${VORTEX_SERVICE}
    echo -e "${green}✓ VortexUiPro restarted${plain}"
}

# ─── Status ─────────────────────────────────────────────────────────────
cmd_status() {
    systemctl status ${VORTEX_SERVICE} --no-pager
}

# ─── Logs ───────────────────────────────────────────────────────────────
cmd_logs() {
    if [[ "$1" == "-f" ]]; then
        journalctl -u ${VORTEX_SERVICE} -f
    else
        journalctl -u ${VORTEX_SERVICE} --no-pager -n 50
    fi
}

# ─── Update ─────────────────────────────────────────────────────────────
cmd_update() {
    echo -e "${yellow}Updating VortexUiPro...${plain}"

    local LATEST_VERSION=$(curl -fsSL "https://api.github.com/repos/iPmartNetwork/VortexUiPro/releases/latest" 2>/dev/null | grep '"tag_name"' | head -1 | cut -d'"' -f4 | tr -d 'v')
    if [[ -z "${LATEST_VERSION}" ]]; then
        echo -e "${yellow}Could not check latest version. Updating from master...${plain}"
        LATEST_VERSION="0.0.1"
    fi

    echo -e "${blue}Current version:${plain} $(cat ${VORTEX_DATA}/VERSION 2>/dev/null || echo 'unknown')"
    echo -e "${blue}Latest version:${plain} ${LATEST_VERSION}"

    # Download and replace binary
    local ARCH="amd64"
    case "$(uname -m)" in
        armv8* | aarch64) ARCH="arm64" ;;
        armv7*) ARCH="armv7" ;;
    esac

    local DOWNLOAD_URL="https://github.com/iPmartNetwork/VortexUiPro/releases/download/v${LATEST_VERSION}/vortexuipro-linux-${ARCH}.tar.gz"

    systemctl stop ${VORTEX_SERVICE}

    if curl -fsSL --connect-timeout 10 "${DOWNLOAD_URL}" -o /tmp/vortexuipro_update.tar.gz; then
        tar -xzf /tmp/vortexuipro_update.tar.gz -C /tmp/
        if [[ -f /tmp/vortexuipro ]]; then
            mv /tmp/vortexuipro ${VORTEX_BIN}.new
        fi
        rm -f /tmp/vortexuipro_update.tar.gz
    else
        echo -e "${yellow}Binary download failed, building from source...${plain}"
        local TMPDIR=$(mktemp -d)
        cd ${TMPDIR}
        git clone --depth 1 https://github.com/iPmartNetwork/VortexUiPro.git .
        go build -o vortexuipro -ldflags="-s -w" ./cmd/panel
        cp vortexuipro ${VORTEX_BIN}.new
        cd /
        rm -rf ${TMPDIR}
    fi

    if [[ -f ${VORTEX_BIN}.new ]]; then
        chmod +x ${VORTEX_BIN}.new
        mv ${VORTEX_BIN}.new ${VORTEX_BIN}
        echo "${LATEST_VERSION}" > ${VORTEX_DATA}/VERSION
        echo -e "${green}✓ Update complete${plain}"
    else
        echo -e "${red}✗ Update failed${plain}"
    fi

    systemctl start ${VORTEX_SERVICE}
}

# ─── Backup ─────────────────────────────────────────────────────────────
cmd_backup() {
    local BACKUP_DIR="${VORTEX_DATA}/backups"
    local TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    local BACKUP_FILE="${BACKUP_DIR}/vortexuipro_backup_${TIMESTAMP}.tar.gz"

    mkdir -p ${BACKUP_DIR}

    echo -e "${yellow}Creating backup...${plain}"

    # Stop the panel briefly to ensure DB consistency
    systemctl stop ${VORTEX_SERVICE}
    sleep 1

    tar -czf ${BACKUP_FILE} \
        -C ${VORTEX_DATA} \
        --exclude='backups/*' \
        --exclude='logs/*' \
        data/ env certs/

    systemctl start ${VORTEX_SERVICE}

    local SIZE=$(du -h ${BACKUP_FILE} | cut -f1)
    echo -e "${green}✓ Backup created:${plain} ${BACKUP_FILE}"
    echo -e "${green}  Size:${plain} ${SIZE}"
}

# ─── Restore ────────────────────────────────────────────────────────────
cmd_restore() {
    local BACKUP_FILE="$1"

    if [[ -z "${BACKUP_FILE}" ]]; then
        # List available backups
        echo -e "${yellow}Available backups:${plain}"
        ls -lh ${VORTEX_DATA}/backups/ 2>/dev/null || echo "  No backups found"
        echo ""
        echo -e "${blue}Usage:${plain} vortexui restore /path/to/backup.tar.gz"
        return
    fi

    if [[ ! -f "${BACKUP_FILE}" ]]; then
        echo -e "${red}Backup file not found: ${BACKUP_FILE}${plain}"
        exit 1
    fi

    echo -e "${yellow}Restoring from backup: ${BACKUP_FILE}${plain}"
    echo -e "${red}Warning: This will overwrite current data!${plain}"

    read -rp "Continue? (y/N): " CONFIRM
    if [[ "${CONFIRM}" != "y" && "${CONFIRM}" != "Y" ]]; then
        echo -e "${yellow}Restore cancelled${plain}"
        exit 0
    fi

    systemctl stop ${VORTEX_SERVICE}

    tar -xzf ${BACKUP_FILE} -C ${VORTEX_DATA}/

    systemctl start ${VORTEX_SERVICE}

    echo -e "${green}✓ Restore complete${plain}"
}

# ─── Change Password ────────────────────────────────────────────────────
cmd_password() {
    echo -e "${yellow}Reset Admin Password${plain}"
    echo -e "${red}This will reset the admin password in the database.${plain}"
    read -rp "Enter new password (leave empty for random): " NEW_PASS

    if [[ -z "${NEW_PASS}" ]]; then
        NEW_PASS=$(openssl rand -base64 12 | tr -dc 'a-zA-Z0-9!@#$%^&*' | head -c 16)
        echo -e "${yellow}Generated password: ${NEW_PASS}${plain}"
    fi

    # Use the panel's internal API to reset password
    # This requires the panel to be running to hash properly with Argon2id
    if systemctl is-active --quiet ${VORTEX_SERVICE}; then
        echo -e "${yellow}Password change requested. Use the panel's admin interface:${plain}"
        echo -e "  ${blue}Access panel, go to Admin > Settings, change password.${plain}"
        echo -e "${yellow}Or stop the panel and edit the database directly (SQLite):${plain}"
        echo -e "  ${blue}systemctl stop ${VORTEX_SERVICE}${plain}"
        echo -e "  ${blue}sqlite3 ${VORTEX_DATA}/data/vortex.db \"SELECT id,username FROM admins WHERE role='super_admin';\"${plain}"
        echo -e "  ${blue}# Update using proper hash from: https://argon2.online${plain}"
    else
        echo -e "${yellow}Panel is not running. Start it first: systemctl start ${VORTEX_SERVICE}${plain}"
    fi
}

# ─── Change Port ────────────────────────────────────────────────────────
cmd_port() {
    local NEW_PORT="$1"
    if [[ -z "${NEW_PORT}" ]]; then
        read -rp "Enter new port: " NEW_PORT
    fi

    if [[ ! "${NEW_PORT}" =~ ^[0-9]+$ ]] || ((NEW_PORT < 1 || NEW_PORT > 65535)); then
        echo -e "${red}Invalid port number${plain}"
        exit 1
    fi

    # Update env file
    sed -i "s/VORTEX_HTTP_ADDR=:[0-9]*/VORTEX_HTTP_ADDR=:${NEW_PORT}/" ${VORTEX_DATA}/env

    systemctl restart ${VORTEX_SERVICE}
    echo -e "${green}✓ Port changed to ${NEW_PORT}${plain}"
    echo -e "${yellow}Update your firewall rules if needed.${plain}"
}

# ─── Info ───────────────────────────────────────────────────────────────
cmd_info() {
    local SERVER_IP=$(curl -fsSL https://api.ipify.org 2>/dev/null || echo "N/A")
    local PANEL_PORT=$(grep 'VORTEX_HTTP_ADDR' ${VORTEX_DATA}/env 2>/dev/null | cut -d: -f2 || echo "8080")

    echo -e "${cyan}══════════════════════════════════════${plain}"
    echo -e "${green}   VortexUiPro Information${plain}"
    echo -e "${cyan}══════════════════════════════════════${plain}"
    echo -e ""
    echo -e "  ${yellow}Version:${plain}  $(cat ${VORTEX_DATA}/VERSION 2>/dev/null || echo 'unknown')"
    echo -e "  ${yellow}Status:${plain}   $(systemctl is-active ${VORTEX_SERVICE} 2>/dev/null || echo 'inactive')"
    echo -e "  ${yellow}Binary:${plain}   ${VORTEX_BIN}"
    echo -e "  ${yellow}Data:${plain}     ${VORTEX_DATA}"
    echo -e "  ${yellow}Port:${plain}     ${PANEL_PORT}"
    echo -e "  ${yellow}Server IP:${plain} ${SERVER_IP}"
    echo -e "  ${yellow}Access:${plain}   http://${SERVER_IP}:${PANEL_PORT}"
    echo -e ""
    echo -e "  ${yellow}DB Type:${plain}  $(grep 'VORTEX_DB_TYPE' ${VORTEX_DATA}/env 2>/dev/null | cut -d= -f2 || echo 'sqlite')"
    echo -e "  ${yellow}DB Path:${plain}  $(grep 'VORTEX_DATABASE_URL' ${VORTEX_DATA}/env 2>/dev/null | cut -d= -f2- || echo 'N/A')"
    echo -e ""
    echo -e "  ${yellow}Logs:${plain}     journalctl -u ${VORTEX_SERVICE} -f"
    echo -e "  ${yellow}Config:${plain}   ${VORTEX_DATA}/env"
    echo -e "${cyan}══════════════════════════════════════${plain}"
}

# ─── Version ────────────────────────────────────────────────────────────
cmd_version() {
    if [[ -f "${VORTEX_BIN}" ]]; then
        ${VORTEX_BIN} -v 2>&1 || echo "VortexUiPro $(cat ${VORTEX_DATA}/VERSION 2>/dev/null || echo 'unknown')"
    else
        echo -e "${red}Binary not found${plain}"
    fi
}

# ─── SSL Cert ───────────────────────────────────────────────────────────
cmd_cert() {
    echo -e "${yellow}SSL Certificate Configuration${plain}"
    echo -e "  ${blue}1.${plain} Let's Encrypt (Domain)"
    echo -e "  ${blue}2.${plain} Let's Encrypt (IP)"
    echo -e "  ${blue}3.${plain} Custom certificate"
    echo -e "  ${blue}4.${plain} Show current cert info"

    read -rp "Choose option: " CHOICE
    case "${CHOICE}" in
        1)
            read -rp "Enter domain: " DOMAIN
            ~/.acme.sh/acme.sh --issue -d ${DOMAIN} --standalone --force 2>/dev/null
            ~/.acme.sh/acme.sh --installcert -d ${DOMAIN} \
                --key-file ${VORTEX_DATA}/certs/privkey.pem \
                --fullchain-file ${VORTEX_DATA}/certs/fullchain.pem \
                --reloadcmd "systemctl restart ${VORTEX_SERVICE}"
            systemctl restart ${VORTEX_SERVICE}
            echo -e "${green}✓ SSL configured for ${DOMAIN}${plain}"
            ;;
        2)
            local IP=$(curl -fsSL https://api.ipify.org)
            ~/.acme.sh/acme.sh --issue -d ${IP} --standalone --force \
                --certificate-profile shortlived --days 6 2>/dev/null
            ~/.acme.sh/acme.sh --installcert -d ${IP} \
                --key-file ${VORTEX_DATA}/certs/privkey.pem \
                --fullchain-file ${VORTEX_DATA}/certs/fullchain.pem \
                --reloadcmd "systemctl restart ${VORTEX_SERVICE}"
            systemctl restart ${VORTEX_SERVICE}
            echo -e "${green}✓ SSL configured for ${IP}${plain}"
            ;;
        3)
            read -rp "Certificate path: " CERT
            read -rp "Key path: " KEY
            if [[ -f "${CERT}" && -f "${KEY}" ]]; then
                cp "${CERT}" ${VORTEX_DATA}/certs/fullchain.pem
                cp "${KEY}" ${VORTEX_DATA}/certs/privkey.pem
                systemctl restart ${VORTEX_SERVICE}
                echo -e "${green}✓ Custom certificate applied${plain}"
            else
                echo -e "${red}Files not found${plain}"
            fi
            ;;
        4)
            echo -e "${yellow}Current certificate:${plain}"
            openssl x509 -in ${VORTEX_DATA}/certs/fullchain.pem -text -noout 2>/dev/null || echo "  No certificate installed"
            ;;
    esac
}

# ─── Main Dispatch ──────────────────────────────────────────────────────
main() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${red}Please run as root${plain}"
        exit 1
    fi

    CMD="$1"
    shift

    case "${CMD}" in
        start)     cmd_start ;;
        stop)      cmd_stop ;;
        restart)   cmd_restart ;;
        status)    cmd_status ;;
        logs)      cmd_logs "$@" ;;
        update)    cmd_update ;;
        backup)    cmd_backup ;;
        restore)   cmd_restore "$@" ;;
        password)  cmd_password ;;
        port)      cmd_port "$@" ;;
        cert)      cmd_cert ;;
        info)      cmd_info ;;
        version)   cmd_version ;;
        help|--help|-h) show_help ;;
        *)
            echo -e "${red}Unknown command: ${CMD}${plain}"
            echo -e "${yellow}Usage: vortexui {start|stop|restart|status|logs|update|backup|restore|password|port|cert|info|version|help}${plain}"
            exit 1
            ;;
    esac
}

main "$@"
