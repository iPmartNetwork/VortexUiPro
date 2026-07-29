<div align="center">
  <br/>
  
  ```ascii
  ██╗   ██╗ ██████╗ ██████╗ ████████╗███████╗██╗  ██╗██╗   ██╗██╗██████╗ ██████╗ 
  ██║   ██║██╔═══██╗██╔══██╗╚══██╔══╝██╔════╝╚██╗██╔╝██║   ██║██║██╔══██╗██╔══██╗
  ██║   ██║██║   ██║██████╔╝   ██║   █████╗   ╚███╔╝ ██║   ██║██║██████╔╝██████╔╝
  ╚██╗ ██╔╝██║   ██║██╔══██╗   ██║   ██╔══╝   ██╔██╗ ██║   ██║██║██╔═══╝ ██╔═══╝ 
   ╚████╔╝ ╚██████╔╝██║  ██║   ██║   ███████╗██╔╝ ██╗╚██████╔╝██║██║     ██║     
    ╚═══╝   ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝╚═╝     ╚═╝     
  ```

  <h1>VortexUiPro</h1>
  <p><strong>The Ultimate Proxy Management Panel</strong> — <em>Next Generation • Enterprise Grade • Multi-Core</em></p>

  <p>
    <a href="https://github.com/iPmartNetwork/VortexUiPro/releases">
      <img src="https://img.shields.io/github/v/release/iPmartNetwork/VortexUiPro?style=for-the-badge&logo=github&color=blueviolet" alt="Release"/>
    </a>
    <a href="https://go.dev/">
      <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go" alt="Go"/>
    </a>
    <a href="https://react.dev/">
      <img src="https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react" alt="React"/>
    </a>
    <a href="LICENSE">
      <img src="https://img.shields.io/github/license/iPmartNetwork/VortexUiPro?style=for-the-badge&color=green" alt="License"/>
    </a>
    <a href="https://github.com/iPmartNetwork/VortexUiPro/stargazers">
      <img src="https://img.shields.io/github/stars/iPmartNetwork/VortexUiPro?style=for-the-badge&logo=github&color=yellow" alt="Stars"/>
    </a>
  </p>

  <p>
    <a href="#-features">Features</a> •
    <a href="#-quick-start">Quick Start</a> •
    <a href="#-architecture">Architecture</a> •
    <a href="#-deployment">Deployment</a> •
    <a href="#-api-reference">API</a> •
    <a href="#-faq">FAQ</a>
  </p>

  <p>
    <strong>English</strong> |
    <a href="README.fa.md">فارسی</a>
  </p>

  <br/>

  <p align="center">
    <b>⚡ 40+ Backend Services • 39 Frontend Pages • 10 Languages • 7 Protocols • Multi-Core ⚡</b>
  </p>

  <br/>
</div>

---

## 📋 Table of Contents

- [✨ Features](#-features)
- [🚀 Quick Start](#-quick-start)
- [🏗️ Architecture](#️-architecture)
- [📸 Screenshots](#-screenshots)
- [🛠️ Deployment](#️-deployment)
- [🔧 Management](#-management)
- [📖 API Reference](#-api-reference)
- [⚙️ Configuration](#️-configuration)
- [🗺️ Roadmap](#️-roadmap)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

---

## ✨ Features

### 🎯 Core Engine

| Feature | Details |
|---------|---------|
| **Multi-Core Support** | Simultaneous **Xray** + **Sing-box** engine management |
| **40+ Backend Services** | Complete proxy lifecycle management |
| **39 Frontend Pages** | Modern glassmorphism UI with dark/light themes |
| **7 Proxy Protocols** | VMess, VLESS, Trojan, Shadowsocks, Hysteria2, WireGuard, MTProto |
| **Multi-Node Cluster** | gRPC mesh with leader election, mTLS, live topology |

### 🛡️ Anti-Censorship Suite

```
┌────────────────────────────────────────────────────┐
│                 Anti-Censorship Suite               │
├──────────┬──────────┬──────────┬───────────────────┤
│TLS Tricks│ Domain   │ WARP+    │ MTProto Proxy    │
│ Fragment │ Fronting │ Outbound │                   │
│ Padding  │ CDN Scan │ Cloudflr │ Telegram Secret  │
│ Mixing   │ Config   │ Integrat │ Auto-Generate    │
│ Anti-DPI │ Generate │         │                   │
├──────────┴──────────┴──────────┴───────────────────┤
│  Clean IP Scanner          Reality Scanner          │
│  Cloudflare IP Discovery   Fingerprint Detection   │
└────────────────────────────────────────────────────┘
```

- **TLS Tricks**: Fragment, padding, mixing, and anti-DPI techniques to bypass deep packet inspection
- **Domain Fronting**: Automated CDN discovery (Cloudflare, Fastly, Akamai, CloudFront) + proxy config generation
- **WARP+ Outbound**: Cloudflare WARP integration for obfuscated outbound connections
- **MTProto**: Telegram MTProto proxy generation with secret management
- **Clean IP Scanner**: Find clean Cloudflare IPs for uninterrupted access
- **Reality Scan**: TLS reality scanning with advanced fingerprinting

### 🔐 Security & Access Control

| Layer | Technology | Description |
|:-----:|:-----------|:------------|
| 🧑‍💼 | **RBAC** | Super Admin, Admin, Reseller with granular permissions |
| 🔑 | **API Tokens** | Scoped, delegated API access with expiration |
| 🌍 | **Geo-Blocking** | Country-based allow/block lists |
| 🔒 | **Password Policy** | Configurable complexity rules, expiration |
| 🚫 | **IP Ban/Whitelist** | IP-based security with auto-ban |
| 📱 | **TOTP/2FA** | Time-based one-time passwords |
| 📝 | **Audit Logs** | Complete security event logging with compliance export |
| 🛡️ | **Threat Detection** | Brute-force protection, suspicious activity alerts |

### 🌐 Cluster & Federation

```
┌─────────────────────────────────────────────────────┐
│                Cluster Mesh (gRPC + mTLS)             │
│                                                       │
│   ┌──────────┐     ┌──────────┐     ┌──────────┐    │
│   │  Node 1  │◄────│  Node 2  │◄────│  Node 3  │    │
│   │ (Leader) │────►│(Follower)│────►│(Follower)│    │
│   └──────────┘     └──────────┘     └──────────┘    │
│        │                 │                │           │
│   ┌────┴────┐      ┌────┴────┐      ┌────┴────┐     │
│   │ 10,000  │      │ 15,000  │      │ 12,000  │     │
│   │ Clients │      │ Clients │      │ Clients │     │
│   └─────────┘      └─────────┘      └─────────┘     │
│                                                       │
│   Features: Heartbeat • Leader Election • Topology   │
│   Federation: Users • Plans • Traffic • Bans Sync    │
└─────────────────────────────────────────────────────┘
```

### 📊 Analytics & Monitoring

- **Real-time Dashboard**: Live traffic metrics, online users, system performance
- **Prometheus Metrics**: Full Prometheus export for Grafana dashboards
- **WebSocket Streaming**: Sub-second data push for real-time updates
- **Smart Health Check**: Configurable probes with auto-recovery rules
- **Online User Tracker**: Real-time user monitoring with IP tracking
- **Network Topology**: Visual Canvas 2D topology map of cluster nodes
- **Traffic Analytics**: Per-user, per-inbound, and aggregate traffic stats

### 💳 Payment & Billing

```
  ┌──────┐    ┌────────────┐    ┌──────────┐
  │ User ├───►│   Payment  ├───►│  Order   │
  └──────┘    │ Gateways   │    │ Created  │
              ├────────────┤    └────┬─────┘
              │ ZarinPal   │         │
              │ NOWPayments│    ┌────┴─────┐
              └────────────┘    │  Plan    │
                                │ Activated│
                                └────┬─────┘
                          ┌──────────┼──────────┐
                          ▼          ▼          ▼
                    ┌─────────┐ ┌────────┐ ┌────────┐
                    │ Wallet  │ │Service │ │Transact│
                    │ Credit  │ │Access  │ │History │
                    └─────────┘ └────────┘ └────────┘
```

- **ZarinPal**: Iranian payment gateway (IRR/IRHT)
- **NOWPayments**: Cryptocurrency payments (BTC, ETH, USDT, +100 coins)
- **Plan Management**: Create/edit/delete subscription plans with bandwidth limits
- **Wallet System**: Internal wallet with deposit/withdraw/transfer
- **Order Management**: Full order lifecycle with proof images

### 📦 Backup & Restore

| Method | Encryption | Schedule | Target |
|:------:|:----------:|:--------:|:------:|
| 🔐 AES-256-GCM | ✅ Automatic | On-demand | Local |
| ☁️ S3 Compatible | ✅ | Auto (24h) | MinIO, AWS S3 |
| 🗄️ Google Drive | ✅ | Auto (24h) | Google Drive |
| 🤖 Telegram | ✅ | On-demand | Telegram Bot |

### 🎨 User Portal

- **Client Dashboard**: Self-service interface for end users
- **Subscription Links**: Auto-generated subscription URLs (Xray JSON, Clash YAML, Sing-box)
- **Traffic Charts**: Real-time visual consumption charts
- **Support Tickets**: Integrated ticketing system with replies
- **Telegram Bot**: Client self-service via Telegram (traffic check, renewal, support)
- **Multi-Protocol Links**: Share links for all 7 supported protocols

### 🐳 Infrastructure & DevOps

```
Native Install                  Docker Install
┌──────────────────┐           ┌──────────────────┐
│  systemd service  │           │  docker compose   │
│  GO binary        │           │  ├─ panel:8080    │
│  SQLite/PostgreSQL│           │  ├─ node-2:8081   │
│  Xray + Sing-box  │           │  └─ node-3:8082   │
│  Caddy reverse px │           │  Caddy + mTLS     │
└──────────────────┘           └──────────────────┘
```

- **Docker Support**: Multi-node cluster with Docker Compose
- **Web Terminal**: SSH console accessible via browser
- **Live Log Streaming**: Real-time log viewer via WebSocket
- **Plugin System**: Extensible architecture for custom functionality
- **Config Versioning**: Full config history with rollback support
- **Docker Container Management**: CRUD for containers, images, logs, stats

---

## 🚀 Quick Start

### 🔥 One-Click Install (Linux)

The fastest way to get started:

```bash
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh)
```

### 🎯 Advanced Install

```bash
# Custom port + SSL
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) \
  --port 9090 \
  --ssl-domain panel.example.com

# Docker cluster
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) --docker

# Non-interactive + skip SSL
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) \
  --port 8080 \
  --skip-ssl
```

### 🐳 Docker Compose (Manual)

```bash
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro
docker compose -f deploy/compose.yml up -d
```

After startup:
| Node | URL | Description |
|:----:|:---:|:------------|
| Node-1 | http://localhost:8080 | Leader candidate (highest priority) |
| Node-2 | http://localhost:8081 | Follower |
| Node-3 | http://localhost:8082 | Follower |
| **Login** | `admin` / `admin123` | Default credentials |

### 📦 Manual Build

```bash
# Prerequisites: Go 1.25+, Node.js 20+
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro

# Build backend
go build -o vortexuipro -ldflags="-s -w" ./cmd/panel

# Build frontend
cd web
npm ci
npm run build
cd ..

# Configure & run
export VORTEX_HTTP_ADDR=:8080
export VORTEX_DB_TYPE=sqlite
export VORTEX_DATABASE_URL=./data/vortex.db
export VORTEX_JWT_SECRET=$(openssl rand -base64 32)

./vortexuipro
```

---

## 🏗️ Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Reverse Proxy (Caddy/Nginx)               │
│            TLS Termination • Static Files • Rate Limit       │
└────────────────────────┬───────────────────────────────────┘
                         │
┌────────────────────────┴───────────────────────────────────┐
│                    Gin HTTP Router                           │
│         REST API • WebSocket • Subscription Endpoints        │
└────────────────────────┬───────────────────────────────────┘
                         │
┌────────────────────────┴───────────────────────────────────┐
│                    Service Layer                             │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐   │
│  │User  │ │Inbound│ │Outbnd│ │Subsc │ │Analyt│ │Backup│   │
│  │ Svc  │ │ Svc  │ │ Svc  │ │ Svc  │ │ Svc  │ │ Svc  │   │
│  └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘   │
│  ┌──┴───┐ ┌──┴───┐ ┌──┴───┐ ┌──┴───┐ ┌──┴───┐ ┌──┴───┐ │
│  │Ticket│ │Anti- │ │Health│ │Cluster│ │Feder │ │Plugin│ │
│  │ Svc  │ │Censor│ │Check │ │ Svc  │ │ Svc  │ │ Svc  │ │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ │
└────────────────────────┬───────────────────────────────────┘
                         │
┌────────────────────────┴───────────────────────────────────┐
│                    Core Engine Layer                         │
│  ┌──────────────────┐      ┌──────────────────┐             │
│  │   Xray gRPC API  │      │  Sing-box Config  │             │
│  │  (Stats + Route) │      │   (JSON Builder)  │             │
│  └────────┬─────────┘      └────────┬─────────┘             │
│           │                         │                       │
│  ┌────────┴─────────────────────────┴─────────┐             │
│  │        Engine Manager (Multi-Core)          │             │
│  │     Hot Reload • Config Diff • Fallback      │             │
│  └─────────────────────────────────────────────┘             │
└────────────────────────┬───────────────────────────────────┘
                         │
┌────────────────────────┴───────────────────────────────────┐
│                    Data Layer                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   SQLite /    │  │   Prometheus  │  │     Redis     │     │
│  │  PostgreSQL   │  │    Metrics    │  │   (Optional)  │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

### 📁 Project Structure

```
VortexUiPro/
├── cmd/
│   └── panel/main.go          # Entry point — service wiring
├── internal/
│   ├── api/                    # HTTP layer
│   │   ├── handlers/          # 30+ route handlers
│   │   ├── middleware/        # Auth, CORS, Rate Limiter, RBAC
│   │   └── hub/              # WebSocket hub
│   ├── cluster/               # Multi-node mesh
│   │   ├── election.go       # Raft-style leader election
│   │   ├── peer.go           # Peer discovery
│   │   └── sync.go           # State replication
│   ├── config/               # Environment-based config
│   ├── core/                 # Engine drivers
│   │   ├── xray/             # gRPC client, process mgmt, traffic
│   │   └── singbox/          # Sing-box config builder
│   ├── database/             # GORM models + migrations (55+)
│   ├── domain/               # Domain types & constants
│   ├── events/               # Event bus (pub/sub)
│   ├── metrics/              # Prometheus collector
│   ├── monitor/              # Health check engine
│   ├── rbac/                 # Role-based access control
│   └── service/              # Business logic (40+ services)
├── web/
│   ├── src/
│   │   ├── pages/            # 39 React pages
│   │   ├── components/       # Reusable UI components
│   │   ├── hooks/            # Custom React hooks
│   │   └── locales/          # 10 i18n languages
│   └── public/
├── deploy/                   # Docker, Caddy, systemd
│   ├── compose.yml           # 3-node cluster compose
│   ├── Dockerfile            # Multi-stage build
│   ├── web.Dockerfile        # Caddy + SPA
│   ├── Caddyfile             # Reverse proxy config
│   └── vortexuipro-panel.service  # systemd unit
├── install.sh               # Auto-install script
├── vortexui.sh              # Management script
├── CHANGELOG.md             # Version history
├── VERSION                  # v0.0.1
├── README.md                # This file
└── README.fa.md             # Persian translation
```

---

## 📸 Screenshots

<div align="center">
  <p align="center">
    <b>🖼️ Screenshots coming soon — stay tuned!</b>
  </p>
  <p align="center">
    <code>🎯 Login Page • 📊 Dashboard • 👥 Users • 📈 Analytics • 🛡️ Security • 🌐 Cluster</code>
  </p>
</div>

---

## 🛠️ Deployment

### 📋 System Requirements

| Component | Minimum | Recommended | Production |
|:---------:|:-------:|:-----------:|:----------:|
| **CPU** | 1 core | 2 cores | 4+ cores |
| **RAM** | 512 MB | 1 GB | 4+ GB |
| **Storage** | 5 GB | 10 GB | 50+ GB |
| **OS** | Linux | Ubuntu 22.04+ | Debian 12 |
| **Xray Core** | v1.8.0 | Latest | Latest |
| **Database** | SQLite | PostgreSQL | PostgreSQL |
| **Node.js** | — | — | 20+ (build only) |

### 🐧 Supported Distributions

| Distribution | Support | Notes |
|:-----------:|:-------:|:------|
| **Ubuntu** 20.04+ | ✅ Full | Recommended |
| **Debian** 11+ | ✅ Full | Production choice |
| **CentOS** 7+ | ✅ Full | AlmaLinux/Rocky |
| **Fedora** 37+ | ✅ Full | — |
| **Arch Linux** | ✅ Full | — |
| **Alpine** 3.18+ | ✅ Docker | Docker-only |
| **OpenSUSE** | ✅ Community | — |

### 🔧 Environment Variables

All configuration is via environment variables.

```bash
# === Server ===
VORTEX_HTTP_ADDR=:8080                     # HTTP listen address
VORTEX_GRPC_ADDR=:50051                    # gRPC listen address
VORTEX_JWT_SECRET=<random-32-chars>        # JWT signing secret
VORTEX_LOG_LEVEL=info                      # debug | info | warn | error

# === Database ===
VORTEX_DB_TYPE=sqlite                      # sqlite | postgres
VORTEX_DATABASE_URL=/etc/vortexuipro/data/vortex.db

# === Core Engine ===
VORTEX_CORE_BIN=/usr/local/bin/xray        # Xray binary path
VORTEX_CORE_CONFIG=/etc/vortexuipro/data/xray.json
VORTEX_CORE_API_PORT=10085

# === Activity Tracking ===
VORTEX_ACTIVITY_FLUSH_SEC=30

# === Telegram Bot ===
# VORTEX_TELEGRAM_BOT_TOKEN=<your-bot-token>

# === Cluster ===
# VORTEX_CLUSTER_ENABLED=true
# VORTEX_CLUSTER_NODE_NAME=node-1
# VORTEX_CLUSTER_ADDR=:1337
# VORTEX_CLUSTER_PEERS=node-2:1337,node-3:1337
# VORTEX_CLUSTER_REGION=us-east
# VORTEX_CLUSTER_PRIORITY=100

# === Plugin System ===
VORTEX_PLUGIN_DIR=/etc/vortexuipro/plugins
```

### 🐳 Docker Compose (Production)

```yaml
# deploy/compose.yml — 3-node cluster with auto-mesh
services:
  node-1:
    build: .
    ports:
      - "8080:8080"     # HTTP API
      - "1337:1337"     # Cluster mesh
    environment:
      VORTEX_CLUSTER_ENABLED: "true"
      VORTEX_CLUSTER_NODE_NAME: "node-1"
      VORTEX_CLUSTER_PEERS: "node-2:1337,node-3:1337"
      VORTEX_CLUSTER_PRIORITY: "200"
      # ... (see deploy/compose.yml for full config)
```

### 🔒 SSL Setup

```bash
# Option 1: Let's Encrypt (Domain) — 90-day auto-renew
vortexui cert     # Choose option 1 from the menu

# Option 2: Let's Encrypt (IP) — 6-day short-lived
vortexui cert     # Choose option 2 from the menu

# Option 3: Custom certificate
vortexui cert     # Choose option 3, enter paths

# Option 4: Reverse proxy (Caddy auto-HTTPS)
# Use the provided Caddyfile with deploy/web.Dockerfile
```

---

## 🔧 Management

After installation, use the `vortexui` command:

```bash
# Service Control
vortexui start                     # Start the panel
vortexui stop                      # Stop the panel
vortexui restart                   # Restart the panel
vortexui status                    # Show service status
vortexui logs [-f]                # View logs (follow with -f)

# Updates & Backup
vortexui update                    # Update to latest version
vortexui backup                    # Create full backup
vortexui restore /path/to/file    # Restore from backup

# Configuration
vortexui password                  # Reset admin password
vortexui port 9090                 # Change HTTP port
vortexui cert                      # Configure SSL certificate

# Information
vortexui info                      # Show installation details
vortexui version                   # Show version
```

---

## 📖 API Reference

### Public Endpoints

```http
GET  /api/v1/health              # Health check
GET  /metrics                    # Prometheus metrics
GET  /ws                         # WebSocket connection
GET  /sub/:clientId              # Subscription config
GET  /sub/:clientId/info         # Subscription info
GET  /sub/:clientId/link         # Subscription link
GET  /sub/:clientId/share-links  # Share links (all protocols)
POST /api/v1/login               # Authentication
POST /api/v1/auth/refresh        # Token refresh
```

### Protected Endpoints (require JWT)

```http
# Admin
GET  /api/v1/admin/users              # List users
POST /api/v1/admin/users              # Create user
PUT  /api/v1/admin/users/:id          # Update user
DEL  /api/v1/admin/users/:id          # Delete user
POST /api/v1/admin/users/:id/clients  # Add client to user

# Inbounds
GET  /api/v1/inbounds                 # List inbounds
POST /api/v1/inbounds                 # Create inbound
PUT  /api/v1/inbounds/:id             # Update inbound
DEL  /api/v1/inbounds/:id             # Delete inbound

# Monitoring
GET  /api/v1/monitor/online           # Online users
GET  /api/v1/monitor/activity         # Recent activity
POST /api/v1/traffic/reset/:id        # Reset user traffic

# Backup
GET  /api/v1/backups                  # List backups
POST /api/v1/backups                  # Create backup
POST /api/v1/backups/:id/restore      # Restore backup
POST /api/v1/backups/:id/sync         # Sync to remote storage

# Subscriptions
GET  /api/v1/sub-profiles             # List sub profiles
POST /api/v1/sub-profiles             # Create sub profile

# Security
GET  /api/v1/security/audit-logs      # View audit logs
POST /api/v1/security/compliance-check # Run compliance check

# Cluster
GET  /api/v1/cluster/status           # Cluster health
GET  /api/v1/cluster/nodes            # List cluster nodes
GET  /api/v1/cluster/topology         # Topology map
```

> 📘 **Full API documentation**: See [Wiki API Reference](https://github.com/iPmartNetwork/VortexUiPro/wiki/API)
> 
> 📘 **Production Docker Compose**: See [`deploy/production.yml`](deploy/production.yml) for single-node production setup

---

## ⚙️ Configuration

### Database

```go
// Supported: sqlite, postgres
VORTEX_DB_TYPE=postgres
VORTEX_DATABASE_URL=postgres://user:pass@localhost:5432/vortexuipro?sslmode=disable
```

### Logging Levels

| Level | Description |
|:-----:|:-----------|
| `debug` | Detailed debug information |
| `info` | General operational messages (default) |
| `warn` | Warning messages |
| `error` | Error conditions |

### Performance Tuning

```bash
# File descriptor limits (systemd)
LimitNOFILE=65535
LimitNPROC=4096

# Activity tracking flush interval
VORTEX_ACTIVITY_FLUSH_SEC=30  # Lower = more real-time, higher = less I/O

# Database connection pool (GORM)
# Configured automatically based on DB type
```

---

## 🗺️ Roadmap

### ✅ Completed (v0.0.1)

| Phase | Feature | Status |
|:-----:|:--------|:------:|
| 1-2 | Core Architecture + Inbound/Outbound/Routing | ✅ |
| 3 | Admin + RBAC + Subscription | ✅ |
| 4 | Cluster Manager + Federation | ✅ |
| 5 | Anti-Censorship + Tickets | ✅ |
| 6 | Security Settings + Email/SMTP | ✅ |
| 7 | Client Groups + Bulk Ops + Config Versions | ✅ |
| 8 | Web Terminal + Live Logs + WARP + TLS Tricks + Plugins | ✅ |
| 9 | Analytics + Payments + Plans + Wallet + Telegram Bot | ✅ |
| 10 | WebRTC + P2P Mesh | ✅ |
| 11 | Network Topology Visualizer | ✅ |
| 12 | Smart Health Check + Auto-Recovery | ✅ |
| 13 | Multi-language i18n (10 languages + RTL) | ✅ |
| 14 | Advanced Backup (AES-256 + S3 + GDrive + Telegram) | ✅ |
| 15 | Domain Fronting + Smart DNS + Docker Native | ✅ |
| 16 | Xray gRPC Real Integration | ✅ |
| 17 | Subscription System Enhancement (7 protocols + Clash + Sing-box) | ✅ |

### 🔮 Planned

| Phase | Feature | Priority |
|:-----:|:--------|:--------:|
| 18 | Comprehensive Test Coverage (unit + integration + e2e) | 🔴 High |
| 19 | Structured Logging (zerolog/zap) | 🟡 Medium |
| 20 | ACME / Let's Encrypt Auto-SSL for Sub Hosts | 🟡 Medium |
| 21 | WireGuard Mesh VPN | 🟢 Low |
| 22 | Email Notification Templates | 🟢 Low |

---

## 🤝 Contributing

We welcome contributions from the community! Here's how to get started:

### Development Setup

```bash
# Clone the repo
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro

# Install Go dependencies
go mod download

# Install frontend dependencies
cd web && npm ci && cd ..

# Start development
go run ./cmd/panel &              # Backend on :8080
cd web && npm run dev &            # Frontend on :5173
```

### Guidelines

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Code Style

- **Go**: Follow standard `gofmt` conventions
- **TypeScript**: Use the project's ESLint + Prettier config
- **Commits**: Use conventional commit format (`feat:`, `fix:`, `docs:`, etc.)

---

## 📄 License

This project is licensed under the **GNU General Public License v3.0** — see the [LICENSE](LICENSE) file for details.

```
Copyright (C) 2026 iPmartNetwork

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
```

---

## 🙏 Acknowledgments

### Core Technologies

| Technology | Purpose |
|:-----------|:--------|
| [Xray-core](https://github.com/XTLS/Xray-core) | Core proxy engine |
| [Sing-box](https://github.com/SagerNet/sing-box) | Universal proxy platform |
| [Gin](https://github.com/gin-gonic/gin) | HTTP framework |
| [GORM](https://gorm.io) | Database ORM |
| [React](https://reactjs.org/) | Frontend framework |
| [Vite](https://vitejs.dev/) | Build tool |
| [Tailwind CSS](https://tailwindcss.com/) | CSS framework |
| [Caddy](https://caddyserver.com/) | Reverse proxy |

### Projects That Inspired Us

- **[VortexUI](https://github.com/example/vortexui)** — Original proxy management panel foundation
- **[Heimdall](https://github.com/example/heimdall)** — Advanced subscription + cluster features

---

## 🌟 Support & Community

<div align="center">
  <p>
    <a href="https://github.com/iPmartNetwork/VortexUiPro/issues">
      <img src="https://img.shields.io/github/issues/iPmartNetwork/VortexUiPro?style=for-the-badge&logo=github" alt="Issues"/>
    </a>
    <a href="https://github.com/iPmartNetwork/VortexUiPro/discussions">
      <img src="https://img.shields.io/github/discussions/iPmartNetwork/VortexUiPro?style=for-the-badge&logo=github" alt="Discussions"/>
    </a>
    <a href="https://t.me/VortexUiPro">
      <img src="https://img.shields.io/badge/Telegram-Channel-2CA5E0?style=for-the-badge&logo=telegram" alt="Telegram"/>
    </a>
  </p>
</div>

- 🐛 **Report Bugs**: [GitHub Issues](https://github.com/iPmartNetwork/VortexUiPro/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/iPmartNetwork/VortexUiPro/discussions)
- 💬 **Community Chat**: [Telegram Group](https://t.me/VortexUiPro)
- 📖 **Documentation**: [Project Wiki](https://github.com/iPmartNetwork/VortexUiPro/wiki)

---

<div align="center">
  <br/>
  <p>
    <strong>Made with ❤️ by the VortexUiPro Team</strong>
  </p>
  <p>
    <sub>If you find VortexUiPro useful, consider giving it a ⭐ on GitHub!</sub>
  </p>
  <br/>
  <p>
    <a href="#-vortexuipro">Back to Top ▲</a>
  </p>
  <br/>
</div>
