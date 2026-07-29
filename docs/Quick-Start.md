# 🚀 Quick Start Guide

> Get VortexUiPro up and running in minutes.

---

## 📋 Prerequisites

| Requirement | Minimum | Recommended |
|:-----------|:--------|:------------|
| **OS** | Linux (Ubuntu 20.04+) | Debian 12 |
| **CPU** | 1 core | 2 cores |
| **RAM** | 512 MB | 1 GB |
| **Storage** | 5 GB | 10 GB |

---

## 🔥 One-Click Install (Recommended)

```bash
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh)
```

The installer will:
1. Detect your OS and architecture
2. Install dependencies (Xray/Sing-Box, SQLite/PostgreSQL)
3. Download and configure VortexUiPro
4. Set up systemd service
5. Start the panel

After installation, access the panel at:

```
URL:      http://<server-ip>:8080
Username: admin
Password: admin123
```

---

## 🎯 Advanced Install Options

```bash
# Custom port
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) --port 9090

# Custom domain with auto-SSL
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) \
  --ssl-domain panel.yourdomain.com

# Docker installation
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) --docker

# PostgreSQL instead of SQLite
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) \
  --db postgres \
  --db-url "postgres://user:pass@localhost:5432/vortexuipro"

# Non-interactive
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) \
  --port 8080 \
  --skip-ssl
```

---

## 🐳 Docker Compose (Manual)

```bash
# Clone repository
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro

# Start the cluster (3 nodes)
docker compose -f deploy/compose.yml up -d

# Verify all nodes are running
docker compose -f deploy/compose.yml ps

# Check logs
docker compose -f deploy/compose.yml logs -f
```

---

## 📦 Manual Build

```bash
# Prerequisites: Go 1.23+, Node.js 22+
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro

# Build backend
go build -o vortexuipro -ldflags="-s -w" ./cmd/panel

# Build frontend
cd web
npm ci
npm run build
cd ..

# Configure
export VORTEX_HTTP_ADDR=:8080
export VORTEX_DB_TYPE=sqlite
export VORTEX_DATABASE_URL=./data/vortex.db
export VORTEX_JWT_SECRET=$(openssl rand -base64 32)

# Run
./vortexuipro
```

---

## 🔧 Management Commands

After installation, use the `vortexui` command:

```bash
# Service control
vortexui start          # Start the panel
vortexui stop           # Stop the panel
vortexui restart        # Restart the panel
vortexui status         # Show service status
vortexui logs -f        # View logs (follow mode)

# Updates & maintenance
vortexui update         # Update to latest version
vortexui backup         # Create backup
vortexui restore        # Restore from backup
vortexui password       # Reset admin password
vortexui port 9090      # Change HTTP port
vortexui cert           # Configure SSL

# Information
vortexui info           # Installation details
vortexui version        # Show version
```

---

## ✅ Post-Install Checklist

- [ ] Change default admin password (`vortexui password`)
- [ ] Configure SSL certificate (`vortexui cert`)
- [ ] Set up JWT secret (auto-generated on first run)
- [ ] Configure Telegram bot (optional, for notifications)
- [ ] Set up backup schedule (`vortexui backup`)
- [ ] Add your first inbound (proxy port)
- [ ] Create users and clients
- [ ] Configure subscription settings

---

## 🔍 Troubleshooting

### Panel won't start
```bash
# Check logs
journalctl -u vortexuipro -n 50 --no-pager

# Verify config
vortexui status

# Check port conflicts
ss -tlnp | grep 8080
```

### Database issues
```bash
# Backup current database
cp /etc/vortex/data/vortex.db /root/vortex.db.bak

# Reset database (⚠️ deletes all data)
rm /etc/vortex/data/vortex.db
systemctl restart vortexuipro
```

### Core engine not running
```bash
# Check Xray status
systemctl status xray

# Restart core
vortexui restart

# Reinstall core
vortexui update --force
```

---

## 📚 Next Steps

- [📖 Full API Reference](API-Reference)
- [⚙️ Configuration Guide](Configuration)
- [🐳 Deployment Guide](Deployment-Guide)
- [🌐 Cluster Setup](Cluster-Setup)
