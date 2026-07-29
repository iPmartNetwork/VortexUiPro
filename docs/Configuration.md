# ⚙️ Configuration Reference

> Complete environment variable reference for VortexUiPro.

---

## 📋 Table of Contents

- [Server Configuration](#server-configuration)
- [Database Configuration](#database-configuration)
- [Core Engine Configuration](#core-engine-configuration)
- [Security Configuration](#security-configuration)
- [Cluster Configuration](#cluster-configuration)
- [Subscription Configuration](#subscription-configuration)
- [Monitoring Configuration](#monitoring-configuration)
- [Telegram Configuration](#telegram-configuration)
- [Backup Configuration](#backup-configuration)
- [Plugin Configuration](#plugin-configuration)
- [Performance Tuning](#performance-tuning)
- [Example Configurations](#example-configurations)

---

## 🖥️ Server Configuration

| Variable | Default | Description |
|:---------|:--------|:------------|
| `VORTEX_HTTP_ADDR` | `:8080` | HTTP listen address |
| `VORTEX_GRPC_ADDR` | `:50051` | gRPC listen address |
| `VORTEX_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `VORTEX_LOG_FORMAT` | `text` | Log format: `text`, `json` |
| `VORTEX_JWT_SECRET` | *(auto)* | JWT signing secret (32+ chars) |
| `VORTEX_JWT_EXPIRY` | `3600` | JWT access token expiry (seconds) |
| `VORTEX_REFRESH_EXPIRY` | `604800` | Refresh token expiry (seconds, default 7 days) |
| `VORTEX_WEB_ROOT` | `./web/dist` | Frontend static files path |
| `VORTEX_CORS_ORIGIN` | `*` | Allowed CORS origin (set to your domain in production) |
| `VORTEX_RATE_LIMIT` | `100` | Default rate limit (requests/minute) |
| `VORTEX_AUTH_RATE_LIMIT` | `20` | Auth rate limit (requests/5minutes) |

---

## 💾 Database Configuration

| Variable | Default | Description |
|:---------|:--------|:------------|
| `VORTEX_DB_TYPE` | `sqlite` | Database type: `sqlite`, `postgres` |
| `VORTEX_DATABASE_URL` | `./data/vortex.db` | Database connection string |
| `VORTEX_DB_MAX_OPEN` | `25` | Max open connections |
| `VORTEX_DB_MAX_IDLE` | `5` | Max idle connections |
| `VORTEX_DB_CONN_MAX_LIFETIME` | `300` | Connection max lifetime (seconds) |
| `VORTEX_DB_AUTO_MIGRATE` | `true` | Auto-run migrations on startup |

**SQLite Examples**:
```bash
VORTEX_DB_TYPE=sqlite
VORTEX_DATABASE_URL=/etc/vortex/data/vortex.db
```

**PostgreSQL Examples**:
```bash
VORTEX_DB_TYPE=postgres
VORTEX_DATABASE_URL=postgres://user:password@localhost:5432/vortex?sslmode=disable
VORTEX_DATABASE_URL=postgres://user:password@postgres-host:5432/vortex?sslmode=require
```

---

## ⚡ Core Engine Configuration

| Variable | Default | Description |
|:---------|:--------|:------------|
| `VORTEX_CORE` | `xray` | Default core engine: `xray`, `singbox` |
| `VORTEX_CORE_BIN` | `/usr/local/bin/xray` | Xray binary path |
| `VORTEX_SINGBOX_BIN` | `/usr/local/bin/sing-box` | Sing-Box binary path |
| `VORTEX_CORE_CONFIG` | `./data/xray.json` | Core config file path |
| `VORTEX_CORE_API_PORT` | `10085` | Xray gRPC API port |
| `VORTEX_LOCAL_NODE` | `true` | Run Xray locally |
| `VORTEX_XRAY_API_ADDR` | `127.0.0.1:10085` | Xray API address |
| `VORTEX_AUTO_RECOVER` | `true` | Auto-restart core on crash |

---

## 🔐 Security Configuration

| Variable | Default | Description |
|:---------|:--------|:------------|
| `VORTEX_ENABLE_2FA` | `true` | Enable TOTP 2FA |
| `VORTEX_PASSWORD_MIN_LENGTH` | `8` | Minimum password length |
| `VORTEX_PASSWORD_REQUIRE_UPPER` | `true` | Require uppercase letters |
| `VORTEX_PASSWORD_REQUIRE_SPECIAL` | `true` | Require special characters |
| `VORTEX_PASSWORD_EXPIRY_DAYS` | `90` | Password expiry (days, 0 = never) |
| `VORTEX_SESSION_MAX_AGE` | `86400` | Session max age (seconds) |
| `VORTEX_MAX_LOGIN_ATTEMPTS` | `5` | Max failed login attempts |
| `VORTEX_LOGIN_BLOCK_MINUTES` | `15` | Block duration after failed logins |
| `VORTEX_AUDIT_LOG_RETENTION` | `90` | Audit log retention (days) |

---

## 🌐 Cluster Configuration

| Variable | Default | Description |
|:---------|:--------|:------------|
| `VORTEX_CLUSTER_ENABLED` | `false` | Enable cluster mode |
| `VORTEX_CLUSTER_NODE_NAME` | `node-1` | Unique node name |
| `VORTEX_CLUSTER_ADDR` | `:1337` | Cluster mesh listen address |
| `VORTEX_CLUSTER_PEERS` | — | Comma-separated peer addresses |
| `VORTEX_CLUSTER_SECRET` | — | Shared cluster secret |
| `VORTEX_CLUSTER_TLS_ENABLED` | `true` | Enable mTLS for cluster |
| `VORTEX_CLUSTER_TLS_CERT` | — | TLS certificate path |
| `VORTEX_CLUSTER_TLS_KEY` | — | TLS key path |
| `VORTEX_CLUSTER_CA_CERT` | — | CA certificate path |
| `VORTEX_CLUSTER_REGION` | `default` | Node region |
| `VORTEX_CLUSTER_PRIORITY` | `100` | Leader election priority (higher = more likely) |
| `VORTEX_CLUSTER_HEARTBEAT` | `5` | Heartbeat interval (seconds) |
| `VORTEX_CLUSTER_SYNC_INTERVAL` | `30` | State sync interval (seconds) |

**Example — 3-Node Cluster**:
```bash
# Node 1 (Leader priority)
VORTEX_CLUSTER_ENABLED=true
VORTEX_CLUSTER_NODE_NAME=node-1
VORTEX_CLUSTER_ADDR=:1337
VORTEX_CLUSTER_PEERS=node-2:1337,node-3:1337
VORTEX_CLUSTER_PRIORITY=200

# Node 2 (Follower)
VORTEX_CLUSTER_ENABLED=true
VORTEX_CLUSTER_NODE_NAME=node-2
VORTEX_CLUSTER_ADDR=:1337
VORTEX_CLUSTER_PEERS=node-1:1337,node-3:1337
VORTEX_CLUSTER_PRIORITY=100
```

---

## 📡 Subscription Configuration

| Variable | Default | Description |
|:---------|:--------|:------------|
| `VORTEX_SUB_HOST` | — | Subscription host (domain) |
| `VORTEX_SUB_PORT` | `443` | Subscription port |
| `VORTEX_SUB_PATH` | `/sub` | Subscription base path |
| `VORTEX_SUB_ENABLE` | `true` | Enable subscription service |
| `VORTEX_SUB_ENABLE_ROUTING` | `true` | Enable routing in subscription |
| `VORTEX_SUB_SINGBOX_ENABLE` | `true` | Enable Sing-Box subscription format |
| `VORTEX_SUB_CLASH_ENABLE` | `true` | Enable Clash subscription format |
| `VORTEX_SUB_OUTLINE_ENABLE` | `true` | Enable Outline subscription format |
| `VORTEX_SUB_LINKS_ENABLE` | `true` | Enable share links |
| `VORTEX_SUB_QR_ENABLE` | `true` | Enable QR code generation |

---

## 📊 Monitoring Configuration

| Variable | Default | Description |
|:---------|:--------|:------------|
| `VORTEX_METRICS_ENABLE` | `true` | Enable Prometheus metrics |
| `VORTEX_METRICS_PATH` | `/metrics` | Metrics endpoint path |
| `VORTEX_ACTIVITY_FLUSH_SEC` | `30` | Activity flush interval (seconds) |
| `VORTEX_ONLINE_TIMEOUT_MIN` | `5` | Online user timeout (minutes) |
| `VORTEX_HEALTH_CHECK_INTERVAL` | `30` | Health check interval (seconds) |
| `VORTEX_PROMETHEUS_SCRAPE` | `true` | Enable Prometheus scraping |

---

## 🤖 Telegram Configuration

| Variable | Default | Description |
|:---------|:--------|:------------|
| `VORTEX_TELEGRAM_BOT_TOKEN` | — | Telegram bot token (from @BotFather) |
| `VORTEX_TELEGRAM_ADMIN_CHAT_ID` | — | Admin notification chat ID |
| `VORTEX_TELEGRAM_NOTIFY_TRAFFIC` | `true` | Send traffic notifications |
| `VORTEX_TELEGRAM_NOTIFY_EXPIRY` | `true` | Send expiry notifications |
| `VORTEX_TELEGRAM_NOTIFY_LOGIN` | `false` | Send login notifications |

---

## 💾 Backup Configuration

| Variable | Default | Description |
|:---------|:--------|:------------|
| `VORTEX_BACKUP_DIR` | `./data/backups` | Backup storage directory |
| `VORTEX_BACKUP_AUTO_ENABLE` | `false` | Enable automatic backups |
| `VORTEX_BACKUP_INTERVAL_HOURS` | `24` | Backup interval (hours) |
| `VORTEX_BACKUP_RETENTION_DAYS` | `7` | Backup retention (days) |
| `VORTEX_BACKUP_ENCRYPT` | `true` | Encrypt backups with AES-256 |

---

## 🔌 Plugin Configuration

| Variable | Default | Description |
|:---------|:--------|:------------|
| `VORTEX_PLUGIN_DIR` | `./plugins` | Plugin directory |
| `VORTEX_PLUGIN_AUTO_LOAD` | `false` | Auto-load plugins on startup |
| `VORTEX_PLUGIN_ALLOW_EXTERNAL` | `false` | Allow loading external plugins |

---

## ⚡ Performance Tuning

```bash
# File descriptor limits (systemd)
LimitNOFILE=65535
LimitNPROC=4096

# Activity tracking (lower = more real-time, higher = less I/O)
VORTEX_ACTIVITY_FLUSH_SEC=30

# Database connection pool
VORTEX_DB_MAX_OPEN=25
VORTEX_DB_MAX_IDLE=5
VORTEX_DB_CONN_MAX_LIFETIME=300

# Core engine
VORTEX_AUTO_RECOVER=true

# Rate limiting
VORTEX_RATE_LIMIT=100
VORTEX_AUTH_RATE_LIMIT=20
```

---

## 📝 Example Configurations

### Development

```bash
VORTEX_HTTP_ADDR=:8080
VORTEX_LOG_LEVEL=debug
VORTEX_DB_TYPE=sqlite
VORTEX_DATABASE_URL=./data/vortex.db
VORTEX_JWT_SECRET=dev-secret-key-change-in-production
VORTEX_CORS_ORIGIN=*
```

### Production (Single Server)

```bash
VORTEX_HTTP_ADDR=:443
VORTEX_LOG_LEVEL=info
VORTEX_DB_TYPE=postgres
VORTEX_DATABASE_URL=postgres://vortex:vortex123@localhost:5432/vortex?sslmode=disable
VORTEX_JWT_SECRET=$(openssl rand -base64 32)
VORTEX_CORS_ORIGIN=https://panel.yourdomain.com
VORTEX_ENABLE_2FA=true
VORTEX_SUB_HOST=panel.yourdomain.com
```

### Production (Cluster)

```bash
VORTEX_HTTP_ADDR=:8080
VORTEX_LOG_LEVEL=info
VORTEX_DB_TYPE=postgres
VORTEX_DATABASE_URL=postgres://vortex:vortex123@postgres:5432/vortex?sslmode=disable
VORTEX_JWT_SECRET=$(openssl rand -base64 32)
VORTEX_CORS_ORIGIN=https://panel.yourdomain.com
VORTEX_CLUSTER_ENABLED=true
VORTEX_CLUSTER_NODE_NAME=node-1
VORTEX_CLUSTER_ADDR=:1337
VORTEX_CLUSTER_PEERS=node-2:1337,node-3:1337
VORTEX_CLUSTER_SECRET=cluster-shared-secret
VORTEX_CLUSTER_TLS_ENABLED=true
VORTEX_CLUSTER_PRIORITY=200
VORTEX_METRICS_ENABLE=true
VORTEX_BACKUP_AUTO_ENABLE=true
```

---

<div align="center">

*Last updated: 2025-07-29 • VortexUiPro v0.1.0-beta*

</div>
