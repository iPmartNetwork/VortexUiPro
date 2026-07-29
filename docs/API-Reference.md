# 📖 API Reference

> **Base URL**: `http://<panel-address>:8080`
> **WebSocket**: `ws://<panel-address>:8080/ws`
> **Metrics**: `http://<panel-address>:8080/metrics`

---

## 📋 Table of Contents

- [Authentication](#authentication)
- [Public Endpoints](#public-endpoints)
- [Protected Endpoints](#protected-endpoints)
  - [Auth & Profile](#-auth--profile)
  - [Users](#-users)
  - [Inbounds](#-inbounds)
  - [Outbounds](#-outbounds)
  - [Nodes](#-nodes)
  - [Admin Management](#-admin-management)
  - [Roles & Permissions](#-roles--permissions)
  - [API Tokens](#-api-tokens)
  - [Tickets](#-tickets)
  - [Settings](#-settings)
  - [Subscription](#-subscription)
  - [Subscription Profiles](#-subscription-profiles)
  - [Routing](#-routing)
  - [Cluster](#-cluster)
  - [Federation](#-federation)
  - [Backup & Restore](#-backup--restore)
  - [Analytics](#-analytics)
  - [Monitoring](#-monitoring)
  - [Security](#-security)
  - [Health Check](#-health-check)
  - [Anti-Censorship](#-anti-censorship)
  - [Domain Fronting](#-domain-fronting)
  - [Smart DNS](#-smart-dns)
  - [Xray API](#-xray-api)
  - [WARP+](#-warp)
  - [TLS Tricks](#-tls-tricks)
  - [WebRTC](#-webrtc)
  - [Terminal](#-terminal)
  - [Live Logs](#-live-logs)
  - [Plugins](#-plugins)
  - [Docker](#-docker)
  - [Email / SMTP](#-email--smtp)
  - [Clean IP Scanner](#-clean-ip-scanner)
  - [Config Versions](#-config-versions)
  - [Client Groups](#-client-groups)
  - [Portal (Client Self-Service)](#-portal-client-self-service)
  - [Telegram Bot](#-telegram-bot)
  - [System](#-system)
  - [WebSocket Endpoints](#-websocket-endpoints)
- [Error Codes](#error-codes)
- [Data Models](#data-models)

---

## 🔐 Authentication

### Login

Authenticate and receive JWT tokens.

```http
POST /api/v1/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### Refresh Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600
}
```

### TOTP Setup

```http
POST /api/v1/auth/totp/setup
Authorization: Bearer <token>
```

**Response**:
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code": "data:image/png;base64,..."
}
```

### TOTP Validate

```http
POST /api/v1/auth/totp/validate
Authorization: Bearer <token>
Content-Type: application/json

{
  "code": "123456"
}
```

### Change Password

```http
POST /api/v1/auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "current_password": "oldpass123",
  "new_password": "newpass456"
}
```

---

## 🌐 Public Endpoints

### Health Check

```http
GET /api/v1/health
```

**Response**:
```json
{
  "status": "ok",
  "version": "0.1.0-beta",
  "uptime": "2h15m30s",
  "core": "xray",
  "db": "sqlite"
}
```

### Prometheus Metrics

```http
GET /metrics
```

Returns Prometheus-formatted metrics (for scraping by Prometheus/Grafana).

### WebSocket

```http
GET /ws
```

Upgrades to WebSocket connection for real-time updates (traffic, online users, metrics).

---

## 🛡️ Protected Endpoints

> All protected endpoints require `Authorization: Bearer <token>` header.

---

### 👤 Auth & Profile

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/me` | All | Get current user profile |

**Response**:
```json
{
  "id": 1,
  "username": "admin",
  "role": "super_admin",
  "totp_enabled": false,
  "created_at": "2025-07-29T00:00:00Z"
}
```

---

### 👥 Users

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/admin/users` | super_admin, admin | List all users |
| `GET` | `/api/v1/admin/users/:id` | super_admin, admin | Get user details |
| `POST` | `/api/v1/admin/users` | super_admin, admin | Create user |
| `PUT` | `/api/v1/admin/users/:id` | super_admin, admin | Update user |
| `DELETE` | `/api/v1/admin/users/:id` | super_admin, admin | Delete user |
| `GET` | `/api/v1/admin/users/:id/clients` | super_admin, admin | List user's clients |
| `POST` | `/api/v1/admin/users/:id/clients` | super_admin, admin | Add client to user |
| `DELETE` | `/api/v1/admin/clients/:clientId` | super_admin, admin | Delete client |
| `POST` | `/api/v1/admin/users/:id/reset-traffic` | super_admin, admin | Reset user traffic |

**Create User**:
```http
POST /api/v1/admin/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "newuser",
  "password": "securepass123",
  "email": "user@example.com",
  "role": "user",
  "enabled": true,
  "traffic_limit_gb": 100,
  "expire_days": 30,
  "note": "VIP customer"
}
```

**Response**:
```json
{
  "id": 42,
  "username": "newuser",
  "email": "user@example.com",
  "role": "user",
  "enabled": true,
  "traffic_limit_gb": 100,
  "traffic_used_gb": 0,
  "expire_at": "2025-08-28T00:00:00Z",
  "created_at": "2025-07-29T00:00:00Z"
}
```

---

### 📡 Inbounds

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/inbounds` | super_admin, admin, reseller | List all inbounds |
| `GET` | `/api/v1/inbounds/:id` | super_admin, admin, reseller | Get inbound details |
| `POST` | `/api/v1/inbounds` | super_admin, admin | Create inbound |
| `PUT` | `/api/v1/inbounds/:id` | super_admin, admin | Update inbound |
| `DELETE` | `/api/v1/inbounds/:id` | super_admin, admin | Delete inbound |
| `GET` | `/api/v1/inbounds/xray-config` | super_admin, admin, reseller | Get Xray config JSON |

**Create Inbound**:
```http
POST /api/v1/inbounds
Authorization: Bearer <token>
Content-Type: application/json

{
  "protocol": "vmess",
  "tag": "inbound-vmess-main",
  "port": 443,
  "listen": "0.0.0.0",
  "stream_settings": {
    "network": "tcp",
    "security": "tls",
    "tls_settings": {
      "server_name": "example.com",
      "cert_file": "/etc/vortex/certs/cert.pem",
      "key_file": "/etc/vortex/certs/key.pem"
    }
  },
  "sniffing": {
    "enabled": true,
    "dest_override": ["http", "tls"]
  },
  "remark": "Main VMess Inbound",
  "enable": true
}
```

---

### 🔌 Outbounds

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/outbounds` | super_admin, admin | List all outbounds |
| `GET` | `/api/v1/outbounds/:id` | super_admin, admin | Get outbound details |
| `POST` | `/api/v1/outbounds` | super_admin, admin | Create outbound |
| `PUT` | `/api/v1/outbounds/:id` | super_admin, admin | Update outbound |
| `DELETE` | `/api/v1/outbounds/:id` | super_admin, admin | Delete outbound |
| `PUT` | `/api/v1/outbounds/:id/visibility` | super_admin, admin | Toggle outbound visibility |

---

### 🖥️ Nodes

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/nodes` | super_admin, admin | List all nodes |
| `GET` | `/api/v1/nodes/:id` | super_admin, admin | Get node details |
| `POST` | `/api/v1/nodes` | super_admin, admin | Create node |
| `PUT` | `/api/v1/nodes/:id` | super_admin, admin | Update node |
| `DELETE` | `/api/v1/nodes/:id` | super_admin, admin | Delete node |

---

### 👑 Admin Management

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/admins` | super_admin, admin | List all admins |
| `GET` | `/api/v1/admins/:id` | super_admin, admin | Get admin details |
| `POST` | `/api/v1/admins` | super_admin, admin | Create admin |
| `PUT` | `/api/v1/admins/:id` | super_admin, admin | Update admin |
| `DELETE` | `/api/v1/admins/:id` | super_admin, admin | Delete admin |
| `PUT` | `/api/v1/admins/:id/status` | super_admin, admin | Enable/disable admin |

---

### 🎭 Roles & Permissions

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/roles` | super_admin, admin | List all roles |
| `GET` | `/api/v1/roles/:id` | super_admin, admin | Get role details |
| `POST` | `/api/v1/roles` | super_admin, admin | Create role |
| `PUT` | `/api/v1/roles/:id` | super_admin, admin | Update role |
| `POST` | `/api/v1/roles/:id/duplicate` | super_admin, admin | Duplicate role |
| `DELETE` | `/api/v1/roles/:id` | super_admin, admin | Delete role |

---

### 🔑 API Tokens

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/api-tokens` | super_admin | List API tokens |
| `GET` | `/api/v1/api-tokens/subjects` | super_admin | List delegated subjects |
| `POST` | `/api/v1/api-tokens` | super_admin | Create API token |
| `DELETE` | `/api/v1/api-tokens/:id` | super_admin | Delete API token |
| `PUT` | `/api/v1/api-tokens/:id/status` | super_admin | Enable/disable token |

---

### 🎫 Tickets

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/tickets` | super_admin, admin | List all tickets |
| `GET` | `/api/v1/tickets/stats` | super_admin, admin | Ticket statistics |
| `GET` | `/api/v1/tickets/:id` | super_admin, admin | Get ticket details |
| `POST` | `/api/v1/tickets` | super_admin, admin | Create ticket |
| `POST` | `/api/v1/tickets/:id/reply` | super_admin, admin | Reply to ticket |
| `POST` | `/api/v1/tickets/:id/close` | super_admin, admin | Close ticket |
| `DELETE` | `/api/v1/tickets/:id` | super_admin, admin | Delete ticket |

---

### ⚙️ Settings

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/settings` | super_admin | List all settings |
| `GET` | `/api/v1/settings/:key` | super_admin | Get setting by key |
| `PUT` | `/api/v1/settings` | super_admin | Update settings (batch) |

---

### 📡 Subscription

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/sub/:clientId` | Public | Get subscription config (Xray JSON) |
| `GET` | `/sub/:clientId/info` | Public | Get subscription info (traffic, expiry) |
| `GET` | `/sub/:clientId/link` | Public | Get subscription link |
| `GET` | `/sub/:clientId/share-links` | Public | Get share links (all formats) |
| `GET` | `/sub-group/:subId` | Public | Get grouped subscription links |

**Subscription Info Response**:
```json
{
  "username": "user123",
  "traffic_used_gb": 15.5,
  "traffic_limit_gb": 100,
  "expire_at": "2025-08-28T00:00:00Z",
  "days_left": 30,
  "status": "active",
  "protocols": ["vmess", "vless", "trojan"]
}
```

---

### 📋 Subscription Profiles

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/sub-profiles` | super_admin, admin | List subscription profiles |
| `POST` | `/api/v1/sub-profiles` | super_admin, admin | Create sub profile |
| `DELETE` | `/api/v1/sub-profiles/:id` | super_admin, admin | Delete sub profile |
| `GET` | `/api/v1/sub-hosts` | super_admin, admin | List subscription hosts |
| `POST` | `/api/v1/sub-hosts` | super_admin, admin | Create sub host |
| `DELETE` | `/api/v1/sub-hosts/:id` | super_admin, admin | Delete sub host |
| `GET` | `/api/v1/sub-formats` | super_admin, admin | List available formats |
| `GET` | `/api/v1/sub-vars` | super_admin, admin | List remark variables |

---

### 🔀 Routing

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/routing/rules` | super_admin, admin | List routing rules |
| `GET` | `/api/v1/routing/rules/:id` | super_admin, admin | Get routing rule |
| `POST` | `/api/v1/routing/rules` | super_admin, admin | Create routing rule |
| `PUT` | `/api/v1/routing/rules/:id` | super_admin, admin | Update routing rule |
| `DELETE` | `/api/v1/routing/rules/:id` | super_admin, admin | Delete routing rule |
| `PUT` | `/api/v1/routing/rules/:id/toggle` | super_admin, admin | Toggle rule enable/disable |
| `GET` | `/api/v1/routing/packs` | super_admin, admin | List routing packs |
| `POST` | `/api/v1/routing/packs` | super_admin, admin | Create routing pack |
| `DELETE` | `/api/v1/routing/packs/:id` | super_admin, admin | Delete routing pack |
| `GET` | `/api/v1/routing/generate` | super_admin, admin | Generate full routing config |

---

### 🌐 Cluster

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/cluster/status` | super_admin, admin | Cluster health status |
| `GET` | `/api/v1/cluster/nodes` | super_admin, admin | List cluster nodes |
| `GET` | `/api/v1/cluster/nodes/:id` | super_admin, admin | Get node details |
| `POST` | `/api/v1/cluster/nodes` | super_admin, admin | Add node to cluster |
| `PUT` | `/api/v1/cluster/nodes/:id` | super_admin, admin | Update node |
| `DELETE` | `/api/v1/cluster/nodes/:id` | super_admin, admin | Remove node |
| `GET` | `/api/v1/cluster/election` | super_admin, admin | Election statistics |
| `GET` | `/api/v1/cluster/sync-events` | super_admin, admin | Sync events log |
| `POST` | `/api/v1/cluster/election/force` | super_admin, admin | Force new election |
| `GET` | `/api/v1/cluster/topology` | super_admin, admin | Cluster topology map |
| `GET` | `/api/v1/cluster/pki` | super_admin, admin | PKI/mTLS certificate status |

---

### 🔗 Federation

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/federation/providers` | super_admin, admin | List federation providers |
| `POST` | `/api/v1/federation/providers` | super_admin, admin | Create provider |
| `PUT` | `/api/v1/federation/providers/:id` | super_admin, admin | Update provider |
| `DELETE` | `/api/v1/federation/providers/:id` | super_admin, admin | Delete provider |
| `POST` | `/api/v1/federation/providers/:id/test` | super_admin, admin | Test provider connection |
| `POST` | `/api/v1/federation/sync` | super_admin, admin | Trigger full sync |
| `POST` | `/api/v1/federation/sync/:id` | super_admin, admin | Trigger sync with provider |

**Incoming Federation (Key-based auth)**:

| Method | Endpoint | Description |
|:------:|:---------|:------------|
| `GET/POST` | `/api/v1/federation/users` | Sync users |
| `GET/POST` | `/api/v1/federation/plans` | Sync plans |
| `POST` | `/api/v1/federation/traffic` | Sync traffic data |

---

### 💾 Backup & Restore

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/backups` | super_admin, admin | List backups |
| `POST` | `/api/v1/backups` | super_admin, admin | Create backup |
| `GET` | `/api/v1/backups/:id/download` | super_admin, admin | Download backup |
| `DELETE` | `/api/v1/backups/:id` | super_admin, admin | Delete backup |
| `POST` | `/api/v1/backups/:id/restore` | super_admin, admin | Restore from backup |
| `POST` | `/api/v1/backups/auto-config` | super_admin, admin | Configure auto-backup |
| `POST` | `/api/v1/backups/:id/sync` | super_admin, admin | Sync to remote storage |
| `POST` | `/api/v1/backups/:id/telegram` | super_admin, admin | Send backup via Telegram |

**Encryption Keys** (super_admin only):

| Method | Endpoint | Description |
|:------:|:---------|:------------|
| `GET` | `/api/v1/backups/encryption/keys` | List encryption keys |
| `POST` | `/api/v1/backups/encryption/keys` | Generate encryption key |
| `DELETE` | `/api/v1/backups/encryption/keys/:id` | Delete encryption key |

**Remote Storage** (super_admin only):

| Method | Endpoint | Description |
|:------:|:---------|:------------|
| `GET` | `/api/v1/backups/remote-storage` | List remote storage configs |
| `POST` | `/api/v1/backups/remote-storage` | Save remote storage config |
| `DELETE` | `/api/v1/backups/remote-storage/:id` | Delete remote storage config |

---

### 📊 Analytics

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/analytics/stats` | super_admin, admin | General statistics |
| `GET` | `/api/v1/analytics/traffic` | super_admin, admin | Traffic analytics |
| `GET` | `/api/v1/analytics/user-growth` | super_admin, admin | User growth chart |
| `GET` | `/api/v1/analytics/revenue` | super_admin, admin | Revenue analytics |
| `GET` | `/api/v1/analytics/online` | super_admin, admin | Online user analytics |
| `GET` | `/api/v1/metrics` | super_admin, admin | System metrics snapshot |
| `GET` | `/api/v1/metrics/history` | super_admin, admin | Metrics history |

---

### 📈 Monitoring

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/monitor/online` | super_admin, admin | List online users |
| `GET` | `/api/v1/monitor/online/count` | super_admin, admin | Online user count |
| `GET` | `/api/v1/monitor/activity` | super_admin, admin | Recent user activity |
| `POST` | `/api/v1/traffic/reset/:id` | super_admin, admin | Reset user traffic |
| `GET` | `/api/v1/traffic/sync` | super_admin, admin | Sync traffic data |
| `GET` | `/api/v1/resellers/stats` | super_admin, admin | Reseller statistics |

---

### 🔒 Security

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/security/audit-logs` | super_admin, admin | List audit logs |
| `GET` | `/api/v1/security/threat-summary` | super_admin, admin | Threat summary |
| `POST` | `/api/v1/security/compliance-check` | super_admin, admin | Run compliance check |

**Security Settings**:

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/settings/security/geo-block` | super_admin, admin | Get geo-block config |
| `PUT` | `/api/v1/settings/security/geo-block` | super_admin, admin | Set geo-block rules |
| `GET` | `/api/v1/settings/security/password-policy` | super_admin, admin | Get password policy |
| `PUT` | `/api/v1/settings/security/password-policy` | super_admin, admin | Save password policy |
| `GET` | `/api/v1/settings/security/banned-ips` | super_admin, admin | List banned IPs |
| `POST` | `/api/v1/settings/security/banned-ips` | super_admin, admin | Add banned IP |
| `DELETE` | `/api/v1/settings/security/banned-ips` | super_admin, admin | Remove banned IP |
| `GET` | `/api/v1/settings/security/whitelisted-ips` | super_admin, admin | List whitelisted IPs |
| `POST` | `/api/v1/settings/security/whitelisted-ips` | super_admin, admin | Add whitelisted IP |
| `DELETE` | `/api/v1/settings/security/whitelisted-ips` | super_admin, admin | Remove whitelisted IP |
| `GET` | `/api/v1/settings/security/threat-config` | super_admin, admin | Get threat detection config |

---

### 💓 Health Check

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/health/configs` | super_admin, admin | List check configs |
| `POST` | `/api/v1/health/configs` | super_admin, admin | Create check config |
| `PUT` | `/api/v1/health/configs/:id` | super_admin, admin | Update check config |
| `DELETE` | `/api/v1/health/configs/:id` | super_admin, admin | Delete check config |
| `GET` | `/api/v1/health/statuses` | super_admin, admin | Get check statuses |
| `GET` | `/api/v1/health/configs/:id/history` | super_admin, admin | Get check history |
| `POST` | `/api/v1/health/manual-check` | super_admin, admin | Run manual check |
| `GET` | `/api/v1/health/recovery-rules` | super_admin, admin | List recovery rules |
| `POST` | `/api/v1/health/recovery-rules` | super_admin, admin | Create recovery rule |
| `PUT` | `/api/v1/health/recovery-rules/:id` | super_admin, admin | Update recovery rule |
| `DELETE` | `/api/v1/health/recovery-rules/:id` | super_admin, admin | Delete recovery rule |
| `GET` | `/api/v1/health/recovery-history` | super_admin, admin | Get recovery history |

---

### 🛡️ Anti-Censorship

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/anticensor/tricks` | super_admin, admin | List anti-censorship tricks |
| `GET` | `/api/v1/anticensor/fingerprints` | super_admin, admin | List TLS fingerprints |
| `GET` | `/api/v1/anticensor/scan` | super_admin, admin | Scan target censorship status |
| `GET` | `/api/v1/anticensor/decoy` | super_admin, admin | Generate decoy config |
| `GET` | `/api/v1/anticensor/cert` | super_admin, admin | Generate self-signed cert |
| `POST` | `/api/v1/anticensor/cert/save` | super_admin, admin | Save certificate |
| `GET` | `/api/v1/anticensor/fragment` | super_admin, admin | Get TLS fragment config |
| `GET` | `/api/v1/anticensor/padding` | super_admin, admin | Get TLS padding config |
| `GET` | `/api/v1/anticensor/mix` | super_admin, admin | Generate mixed config |
| `GET` | `/api/v1/anticensor/anti-dpi` | super_admin, admin | Generate anti-DPI config |
| `GET` | `/api/v1/anticensor/mtproto` | super_admin, admin | Generate MTProto config |
| `GET` | `/api/v1/anticensor/warp` | super_admin, admin | Generate WARP config |

---

### 🌍 Domain Fronting

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/domain-fronting/providers` | super_admin, admin | List CDN providers |
| `GET` | `/api/v1/domain-fronting/scan` | super_admin, admin | Scan domain fronting |
| `POST` | `/api/v1/domain-fronting/scan-all` | super_admin, admin | Scan all providers |
| `GET` | `/api/v1/domain-fronting/domains` | super_admin, admin | List frontable domains |
| `GET` | `/api/v1/domain-fronting/generate-config` | super_admin, admin | Generate fronting config |
| `DELETE` | `/api/v1/domain-fronting/domains/:id` | super_admin, admin | Delete fronting domain |

---

### 🔬 Smart DNS

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/dns/resolve` | super_admin, admin | Resolve DNS query |
| `GET` | `/api/v1/dns/configs` | super_admin, admin | List DNS configs |
| `POST` | `/api/v1/dns/configs` | super_admin, admin | Save DNS config |
| `DELETE` | `/api/v1/dns/configs/:id` | super_admin, admin | Delete DNS config |
| `GET` | `/api/v1/dns/rules` | super_admin, admin | List DNS rules |
| `POST` | `/api/v1/dns/rules` | super_admin, admin | Save DNS rule |
| `DELETE` | `/api/v1/dns/rules/:id` | super_admin, admin | Delete DNS rule |
| `POST` | `/api/v1/dns/load-ad-block` | super_admin, admin | Load ad-block list |
| `POST` | `/api/v1/dns/clear-cache` | super_admin, admin | Clear DNS cache |

---

### ⚡ Xray API

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/xray/process` | super_admin, admin | Get Xray process info |
| `GET` | `/api/v1/xray/logs` | super_admin, admin | Get Xray logs |
| `POST` | `/api/v1/xray/validate` | super_admin, admin | Validate Xray config |
| `GET` | `/api/v1/xray/online-users` | super_admin, admin | Get online users (from Xray) |
| `GET` | `/api/v1/xray/traffic` | super_admin, admin | Get traffic stats (from Xray) |
| `POST` | `/api/v1/xray/test-route` | super_admin, admin | Test routing rule |
| `GET` | `/api/v1/xray/balancers/:tag` | super_admin, admin | Get balancer info |
| `PUT` | `/api/v1/xray/balancers/target` | super_admin, admin | Set balancer target |

---

### 🌩️ WARP+

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/warp/config` | super_admin, admin | Get WARP config |
| `PUT` | `/api/v1/warp/config` | super_admin, admin | Update WARP config |
| `POST` | `/api/v1/warp/connect` | super_admin, admin | Connect WARP+ |
| `POST` | `/api/v1/warp/disconnect` | super_admin, admin | Disconnect WARP |
| `GET` | `/api/v1/warp/status` | super_admin, admin | Get WARP status |
| `GET` | `/api/v1/warp/xray-outbound` | super_admin, admin | Get WARP Xray outbound config |

---

### 🎭 TLS Tricks

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/tls-tricks/profiles` | super_admin, admin | List TLS trick profiles |
| `GET` | `/api/v1/tls-tricks/profiles/:id` | super_admin, admin | Get TLS trick profile |
| `POST` | `/api/v1/tls-tricks/profiles` | super_admin, admin | Save TLS trick profile |
| `PUT` | `/api/v1/tls-tricks/profiles/:id` | super_admin, admin | Enable/disable profile |
| `DELETE` | `/api/v1/tls-tricks/profiles/:id` | super_admin, admin | Delete profile |
| `GET` | `/api/v1/tls-tricks/generate` | super_admin, admin | Generate TLS trick config |

---

### 🔗 WebRTC

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/webrtc/ice-config` | super_admin, admin | Get ICE/STUN configuration |
| `GET` | `/api/v1/webrtc/turn-servers` | super_admin, admin | List TURN servers |
| `POST` | `/api/v1/webrtc/turn-servers` | super_admin, admin | Create TURN server |
| `DELETE` | `/api/v1/webrtc/turn-servers/:id` | super_admin, admin | Delete TURN server |
| `POST` | `/api/v1/webrtc/turn-servers/test` | super_admin, admin | Test TURN server |
| `GET` | `/api/v1/webrtc/mesh-config` | super_admin, admin | Get P2P mesh config |
| `PUT` | `/api/v1/webrtc/mesh-config` | super_admin, admin | Update mesh config |
| `GET` | `/api/v1/webrtc/peers` | super_admin, admin | List P2P peers |
| `GET` | `/api/v1/webrtc/peers/:id` | super_admin, admin | Get peer details |
| `DELETE` | `/api/v1/webrtc/peers/:id` | super_admin, admin | Disconnect peer |
| `GET` | `/api/v1/webrtc/peers/stats` | super_admin, admin | Get peer stats |
| `POST` | `/api/v1/webrtc/signal` | super_admin, admin | Send signaling message |
| `GET` | `/api/v1/webrtc/discover` | super_admin, admin | Discover nearby peers |
| `GET` | `/api/v1/webrtc/nat-type` | super_admin, admin | Detect NAT type |

---

### 💻 Terminal

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/terminal/sessions` | super_admin, admin | List terminal sessions |
| `DELETE` | `/api/v1/terminal/sessions/:id` | super_admin, admin | Close terminal session |
| `GET` | `/api/v1/terminal/ws` | super_admin, admin | WebSocket terminal (no auth) |

---

### 📋 Live Logs

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/logs/recent` | super_admin, admin | Get recent logs |
| `GET` | `/api/v1/logs/ws` | super_admin, admin | WebSocket log stream (no auth) |

---

### 🔌 Plugins

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/plugins` | super_admin | List loaded plugins |
| `GET` | `/api/v1/plugins/:id` | super_admin | Get plugin details |
| `POST` | `/api/v1/plugins/load` | super_admin | Load plugin |
| `DELETE` | `/api/v1/plugins/:id` | super_admin | Unload plugin |
| `PUT` | `/api/v1/plugins/:id` | super_admin | Enable/disable plugin |

---

### 🐳 Docker

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/docker/status` | super_admin | Docker daemon status |
| `GET` | `/api/v1/docker/containers` | super_admin | List containers |
| `POST` | `/api/v1/docker/containers` | super_admin | Create container |
| `POST` | `/api/v1/docker/containers/:id/start` | super_admin | Start container |
| `POST` | `/api/v1/docker/containers/:id/stop` | super_admin | Stop container |
| `POST` | `/api/v1/docker/containers/:id/restart` | super_admin | Restart container |
| `DELETE` | `/api/v1/docker/containers/:id` | super_admin | Remove container |
| `GET` | `/api/v1/docker/containers/:id/logs` | super_admin | Get container logs |
| `GET` | `/api/v1/docker/containers/:id/stats` | super_admin | Get container stats |
| `GET` | `/api/v1/docker/images` | super_admin | List images |
| `POST` | `/api/v1/docker/images/pull` | super_admin | Pull image |
| `DELETE` | `/api/v1/docker/images/:id` | super_admin | Remove image |
| `POST` | `/api/v1/docker/images/prune` | super_admin | Prune unused images |

---

### 📧 Email / SMTP

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/email/config` | super_admin | Get SMTP config |
| `PUT` | `/api/v1/email/config` | super_admin | Save SMTP config |
| `POST` | `/api/v1/email/test` | super_admin | Send test email |

---

### 🧹 Clean IP Scanner

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/clean-ip/results` | super_admin, admin | Get scan results |
| `POST` | `/api/v1/clean-ip/scan` | super_admin, admin | Run new scan |

---

### 📝 Config Versions

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/config-versions` | super_admin, admin | List config versions |
| `POST` | `/api/v1/config-versions/:id/rollback` | super_admin, admin | Rollback to version |

---

### 👥 Client Groups

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/client-groups` | super_admin, admin | List all groups |
| `POST` | `/api/v1/client-groups` | super_admin, admin | Create group |
| `GET` | `/api/v1/client-groups/with-clients` | super_admin, admin | Groups with client count |
| `GET` | `/api/v1/client-groups/:id` | super_admin, admin | Get group details |
| `PUT` | `/api/v1/client-groups/:id` | super_admin, admin | Update group |
| `DELETE` | `/api/v1/client-groups/:id` | super_admin, admin | Delete group |
| `POST` | `/api/v1/client-groups/:id/clients` | super_admin, admin | Add client to group |
| `DELETE` | `/api/v1/client-groups/:id/clients` | super_admin, admin | Remove client from group |
| `GET` | `/api/v1/client-groups/:id/clients` | super_admin, admin | List group's clients |
| `POST` | `/api/v1/client-groups/:id/clients/bulk` | super_admin, admin | Bulk add clients |
| `DELETE` | `/api/v1/client-groups/:id/clients/bulk` | super_admin, admin | Bulk remove clients |

---

### 🎯 Portal (Client Self-Service)

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/portal/clients` | super_admin, admin | List own clients |
| `GET` | `/api/v1/portal/clients/:id` | super_admin, admin | Get client detail |
| `GET` | `/api/v1/portal/traffic` | super_admin, admin | Get own traffic data |
| `GET` | `/api/v1/portal/tickets` | super_admin, admin | List own tickets |
| `POST` | `/api/v1/portal/tickets` | super_admin, admin | Create support ticket |
| `GET` | `/api/v1/payments` | Public | Payment management |
| `POST` | `/api/v1/payments/zarinpal/request` | All | Request ZarinPal payment |
| `POST` | `/api/v1/payments/nowpayments/create` | All | Create NOWPayments invoice |
| `GET` | `/api/v1/plans` | All | List subscription plans |
| `GET` | `/api/v1/orders` | All | List own orders |

---

### 🤖 Telegram Bot

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `POST` | `/api/v1/telegram/client/link` | super_admin, admin | Link Telegram to client |
| `POST` | `/api/v1/telegram/test` | super_admin, admin | Send test notification |
| `POST` | `/api/v1/telegram/notify` | super_admin, admin | Send client usage notification |

---

### ⚙️ System

| Method | Endpoint | Role | Description |
|:------:|:---------|:----:|:------------|
| `GET` | `/api/v1/system/status` | super_admin | System status |
| `GET` | `/api/v1/system/performance` | super_admin | Performance metrics |
| `GET` | `/api/v1/system/core-status` | super_admin | Core engine status |
| `GET` | `/api/v1/system/config` | super_admin | Running config |
| `GET` | `/api/v1/system/logs` | super_admin | System logs |
| `POST` | `/api/v1/system/reset-traffic` | super_admin | Reset all traffic |

---

### 🔌 WebSocket Endpoints

| Endpoint | Description |
|:---------|:------------|
| `/ws` | Main WebSocket — real-time traffic, online users, metrics |
| `/api/v1/terminal/ws` | WebSSH terminal (interactive shell) |
| `/api/v1/logs/ws` | Live log streaming with color coding |
| `/api/v1/webrtc/signal/ws` | WebRTC signaling channel (P2P mesh) |

---

## ❌ Error Codes

| Status | Code | Description |
|:------:|:-----|:------------|
| `400` | `bad_request` | Invalid request body or parameters |
| `401` | `unauthorized` | Missing or invalid authentication |
| `403` | `forbidden` | Insufficient permissions (wrong role) |
| `404` | `not_found` | Resource not found |
| `409` | `conflict` | Resource already exists |
| `422` | `validation_error` | Input validation failed |
| `429` | `rate_limit_exceeded` | Too many requests (see `X-RateLimit-Zone` header) |
| `500` | `internal_error` | Server error |
| `503` | `service_unavailable` | Core engine not running |

**Error Response Format**:
```json
{
  "error": "validation_error",
  "message": "Invalid port number: must be between 1 and 65535",
  "details": {
    "field": "port",
    "value": 99999
  }
}
```

### Rate Limiting Headers

All responses include rate limit headers:

| Header | Description |
|:-------|:------------|
| `X-RateLimit-Limit` | Maximum requests in the window |
| `X-RateLimit-Remaining` | Remaining requests in the window |
| `X-RateLimit-Zone` | Rate limit zone (auth/api/subscription/default) |

---

## 📦 Data Models

### User
```json
{
  "id": 1,
  "username": "string",
  "email": "string",
  "role": "user | reseller",
  "enabled": true,
  "traffic_limit_gb": 100,
  "traffic_used_gb": 15.5,
  "expire_at": "2025-08-28T00:00:00Z",
  "note": "string",
  "telegram_id": 123456789,
  "group_id": 1,
  "created_at": "2025-07-29T00:00:00Z"
}
```

### Inbound
```json
{
  "id": 1,
  "protocol": "vmess | vless | trojan | shadowsocks | hysteria2 | wireguard | mtproto",
  "tag": "string",
  "port": 443,
  "listen": "0.0.0.0",
  "stream_settings": {
    "network": "tcp | kcp | ws | httpupgrade | splithttp | grpc",
    "security": "none | tls | reality | xtls",
    "tls_settings": { "server_name": "string", "cert_file": "string", "key_file": "string" },
    "reality_settings": { "dest": "string", "server_names": ["string"], "private_key": "string", "short_ids": ["string"] },
    "ws_settings": { "path": "/ws", "headers": {} },
    "grpc_settings": { "service_name": "string" }
  },
  "sniffing": { "enabled": true, "dest_override": ["http", "tls"] },
  "remark": "string",
  "enable": true
}
```

### Client
```json
{
  "id": 1,
  "user_id": 1,
  "email": "client@example.com",
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "traffic_used_gb": 5.2,
  "traffic_limit_gb": 50,
  "expire_at": "2025-08-28T00:00:00Z",
  "enable": true,
  "group_id": 1,
  "inbound_ids": [1, 2, 3],
  "sub_url": "https://panel.example.com/sub/550e8400...",
  "created_at": "2025-07-29T00:00:00Z"
}
```

### Node
```json
{
  "id": 1,
  "name": "node-1",
  "address": "192.168.1.10",
  "port": 50051,
  "api_port": 8080,
  "status": "online | offline | unhealthy",
  "traffic_used_gb": 50.3,
  "traffic_limit_gb": 1000,
  "created_at": "2025-07-29T00:00:00Z"
}
```

### Routing Rule
```json
{
  "id": 1,
  "type": "direct | proxy | block",
  "domains": ["example.com", "*.google.com"],
  "ips": ["1.1.1.0/24"],
  "port": "80,443",
  "outbound_tag": "proxy",
  "remark": "Rule description",
  "enabled": true,
  "order": 1
}
```

### Backup
```json
{
  "id": 1,
  "filename": "backup-2025-07-29-120000.tar.gz",
  "size_bytes": 1048576,
  "encrypted": true,
  "encryption_key_id": 1,
  "remote_storage_id": 1,
  "telegram_sent": false,
  "created_at": "2025-07-29T12:00:00Z"
}
```

### Health Check Config
```json
{
  "id": 1,
  "name": "Xray Core Health",
  "target": "http://localhost:8080/api/v1/health",
  "probe_type": "http | tcp | ping | grpc",
  "interval_seconds": 30,
  "timeout_seconds": 5,
  "threshold": 3,
  "enabled": true
}
```

### Cluster Node (API)
```json
{
  "id": "node-1",
  "address": "192.168.1.10:1337",
  "role": "leader | follower | candidate",
  "status": "online | offline | unhealthy",
  "region": "us-east",
  "priority": 200,
  "term": 3,
  "last_heartbeat": "2025-07-29T12:00:05Z",
  "connected_peers": 2,
  "load": {
    "cpu_percent": 45.2,
    "memory_percent": 62.1,
    "disk_percent": 33.8
  },
  "version": "0.1.0-beta"
}
```

---

<div align="center">

*Last updated: 2025-07-29 • VortexUiPro v0.1.0-beta*

*For additional help, join our [Telegram](https://t.me/VortexUiPro) or open a [GitHub Issue](https://github.com/iPmartNetwork/VortexUiPro/issues).*

</div>
