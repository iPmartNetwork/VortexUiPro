# 🚀 VortexUiPro — Production Deployment Guide

Deploy the **Ultimate Proxy Management Panel** in production using Docker Compose with PostgreSQL, Redis, Caddy (auto HTTPS), automated backups, Prometheus + Grafana monitoring, and log rotation.

---

## 📋 Prerequisites

| Requirement | Minimum | Recommended |
|-------------|---------|------------|
| CPU | 2 cores | 4+ cores |
| RAM | 2 GB | 4+ GB |
| Disk | 20 GB SSD | 50+ GB NVMe |
| Docker | 24+ | Latest |
| Docker Compose | 2.20+ | Latest |
| Domain | Optional | Recommended for HTTPS |

**Supported OS**: Ubuntu 22.04+, Debian 12+, CentOS 9+

---

## 🚀 Quick Start (1 Minute)

```bash
# 1. Clone the repository
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro

# 2. Configure environment
cp .env.example .env
nano .env
# ⚠️ IMPORTANT: Set these minimum variables:
#   POSTGRES_PASSWORD=your-strong-password
#   VORTEX_JWT_SECRET=your-64-char-hex-secret
#   SITE_ADDRESS=panel.yourdomain.com  ← for auto HTTPS

# 3. Create data directories
mkdir -p deploy/data/{postgres,redis,panel,backups,logs}

# 4. Start everything
docker compose -f deploy/production.yml up -d

# 5. Check logs
docker compose -f deploy/production.yml logs -f

# 6. Open browser
echo "Panel: http://localhost:8080  (admin / your-password)"
```

---

## 🔐 Security Checklist

- [ ] **Change default passwords**: admin password, JWT secret, PostgreSQL password
- [ ] **Set SITE_ADDRESS**: enables automatic HTTPS via Let's Encrypt
- [ ] **Set ACME_EMAIL**: receive certificate expiry notifications
- [ ] **Use strong secrets**: `openssl rand -hex 32` for JWT_SECRET
- [ ] **Restrict port access**: only expose ports 80/443 to the internet
- [ ] **Enable fail2ban** for SSH protection
- [ ] **Regular backups**: backup service runs automatically every 6 hours

---

## 📦 Service Overview

| Service | Port | Description |
|---------|------|-------------|
| **Caddy** | 80/443 | Reverse proxy + auto HTTPS (Let's Encrypt) |
| **Panel** | 8080 | Go backend + built SPA (internal) |
| **PostgreSQL** | 5432 | Primary database (internal) |
| **Redis** | 6379 | Cache & session store (internal) |
| **Backup** | — | Automated PostgreSQL dump + S3 sync |
| **Logrotate** | — | Daily log rotation with zstd compression |
| **Prometheus** | 9090 | Metrics collection (internal) |
| **Grafana** | 3000 | Dashboards & visualization (internal) |

---

## ⚙️ Environment Variables

### Required (set these in `.env`)

| Variable | Description | Example |
|----------|-------------|---------|
| `POSTGRES_PASSWORD` | Database password | `my-strong-password` |
| `VORTEX_JWT_SECRET` | JWT signing secret (64+ hex chars) | `$(openssl rand -hex 32)` |

### Recommended

| Variable | Description | Default |
|----------|-------------|---------|
| `SITE_ADDRESS` | Public domain/IP for HTTPS | `:80` |
| `ACME_EMAIL` | Let's Encrypt notification email | — |
| `VORTEX_LOG_LEVEL` | Log verbosity | `info` |
| `VORTEX_ADMIN_USERNAME` | Panel admin username | `admin` |
| `VORTEX_ADMIN_PASSWORD` | Panel admin password | `admin123` |

### Optional Integrations

| Variable | Description | Service |
|----------|-------------|---------|
| `TELEGRAM_BOT_TOKEN` | Telegram bot for notifications | Backup |
| `TELEGRAM_CHAT_ID` | Admin Telegram chat ID | Backup |
| `S3_ENDPOINT` | S3-compatible endpoint for remote backups | Backup |
| `S3_BUCKET` | S3 bucket name | Backup |
| `VORTEX_SMTP_HOST` | SMTP server for email | Panel |
| `VORTEX_TELEGRAM_BOT_TOKEN` | Panel Telegram bot | Panel |
| `VORTEX_ZARINPAL_MERCHANT` | ZarinPal merchant ID | Payment |

---

## 📊 Monitoring

### Prometheus

The panel exposes metrics at `http://panel:8080/metrics` in Prometheus format. Prometheus auto-scrapes every 15 seconds.

**Pre-configured alerts** (in `deploy/prometheus/alerts.yml`):
- 🔴 PanelDown — Panel unreachable for >1m
- 🟡 HighCPUUsage — CPU > 80% for >5m
- 🟡 HighMemoryUsage — Memory > 85% for >5m
- 🔴 DiskSpaceLow — Disk < 10% free
- 🟡 HighErrorRate — Error rate > 5%
- 🔴 NoBackups — No backup in >48h

### Grafana

Access Grafana at `http://your-server:3000` (admin / admin).

A pre-configured dashboard is auto-loaded from `deploy/grafana/vortexuipro-dashboard.json` with:
- Real-time traffic (inbound/outbound)
- Active users & clients
- System resources (CPU, RAM, Disk)
- Online users timeline
- Heartbeat & cluster health

---

## 💾 Backup & Restore

### Automatic Backups

The backup service runs every **6 hours** and:
1. Dumps PostgreSQL database (zstd compressed)
2. Archives configuration files
3. Retains: 7 daily + 4 weekly + 12 monthly backups
4. Syncs to S3 (if configured)
5. Sends Telegram notification (if configured)

### Manual Backup

```bash
# Enter the backup container
docker exec -it vortexuipro-backup sh

# Run full backup
backup.sh --db-only
backup.sh --config-only

# List available backups
backup.sh --list

# Restore from backup
backup.sh --restore /backups/db/vortex-20250101-120000.sql.zst
```

### Restore from Backup

```bash
# 1. Stop the panel
docker compose -f deploy/production.yml stop panel

# 2. Restore database
docker exec vortexuipro-backup backup.sh --restore /backups/db/vortex-YYYYMMDD-HHMMSS.sql.zst

# 3. Restart the panel
docker compose -f deploy/production.yml start panel
```

---

## 🔧 Useful Commands

```bash
# View logs
docker compose -f deploy/production.yml logs -f panel
docker compose -f deploy/production.yml logs -f backup

# Scale (single-node cluster)
docker compose -f deploy/production.yml up -d --scale panel=1

# Update panel
docker compose -f deploy/production.yml pull
docker compose -f deploy/production.yml up -d

# Full restart
docker compose -f deploy/production.yml down
docker compose -f deploy/production.yml up -d

# Database maintenance
docker exec -it vortexuipro-postgres psql -U vortex -d vortex
```

---

## 🔒 Security Hardening

### Firewall (UFW)
```bash
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 22/tcp
ufw enable
```

### Fail2ban
```bash
apt install fail2ban
cat > /etc/fail2ban/jail.local << 'EOF'
[vortexuipro]
enabled = true
port = http,https
filter = caddy
logpath = /var/log/vortexuipro/caddy/access.log
maxretry = 10
bantime = 3600
EOF
systemctl restart fail2ban
```

### Docker Security
- **Read-only root filesystem**: Panel runs with `read_only: true`
- **No new privileges**: `security_opt: no-new-privileges:true`
- **Tmpfs**: Sensitive runtime data in RAM (`/tmp`, `/run`)
- **Internal ports**: PostgreSQL, Redis, Prometheus, Grafana bound to `127.0.0.1`

---

## ❓ Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Panel won't start | PostgreSQL not ready | Wait 20s for DB health check |
| HTTPS not working | No SITE_ADDRESS set | Set `SITE_ADDRESS=panel.yourdomain.com` |
| Backup fails | Missing credentials | Check `POSTGRES_PASSWORD` in `.env` |
| High memory usage | Redis/Xray caching | Reduce `maxmemory` in Redis config |
| WebSocket disconnected | Firewall/proxy blocking | Ensure 80/443 are open | 

---

## 📚 Additional Resources

- [Main README](../README.md) — Full feature documentation
- [Architecture Overview](../README.md#-architecture) — System architecture
- [API Reference](../README.md#-api-reference) — API documentation
- [Docker Compose File](production.yml) — Full compose configuration
- [Prometheus Alerts](prometheus/alerts.yml) — Alert rules
- [Grafana Dashboard](grafana/vortexuipro-dashboard.json) — Pre-configured dashboard
