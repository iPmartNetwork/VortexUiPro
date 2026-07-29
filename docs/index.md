# 🚀 VortexUiPro

**The Ultimate Proxy Management Panel** — *Next Generation • Enterprise Grade • Multi-Core*

<div class="grid cards" markdown>

-   :material-rocket-launch:{ .lg .middle } **Quick Start**

    ---

    Get VortexUiPro running in minutes with our one-click installer.

    [:octicons-arrow-right-24: Get Started](quick-start.md)

-   :material-api:{ .lg .middle } **API Reference**

    ---

    200+ endpoints documented. Complete request/response examples for every API.

    [:octicons-arrow-right-24: API Docs](api-reference.md)

-   :material-docker:{ .lg .middle } **Deployment**

    ---

    Deploy anywhere — native, Docker Compose, or Kubernetes with Helm.

    [:octicons-arrow-right-24: Deployment Guide](deployment-guide.md)

-   :material-cog:{ .lg .middle } **Configuration**

    ---

    40+ environment variables across 10 categories. Full reference guide.

    [:octicons-arrow-right-24: Configuration](configuration.md)

-   :material-server-network:{ .lg .middle } **Cluster Setup**

    ---

    Multi-node mesh with mTLS, leader election, and live topology.

    [:octicons-arrow-right-24: Cluster Guide](cluster-setup.md)

-   :material-github:{ .lg .middle } **Open Source**

    ---

    AGPL-3.0 licensed. Contributions welcome! Star us on GitHub.

    [:octicons-arrow-right-24: GitHub](https://github.com/iPmartNetwork/VortexUiPro)

</div>

---

## ✨ Key Features

### 🎯 Core Engine

| Feature | Details |
|---------|---------|
| **Multi-Core Support** | Simultaneous **Xray** + **Sing-box** engine management |
| **40+ Backend Services** | Complete proxy lifecycle management |
| **39 Frontend Pages** | Modern glassmorphism UI with dark/light themes |
| **7 Proxy Protocols** | VMess, VLESS, Trojan, Shadowsocks, Hysteria2, WireGuard, MTProto |
| **Multi-Node Cluster** | gRPC mesh with leader election, mTLS, live topology |

### 🛡️ Anti-Censorship Suite

- **TLS Tricks**: Fragment, padding, mixing, and anti-DPI techniques
- **Domain Fronting**: Automated CDN discovery (Cloudflare, Fastly, Akamai, CloudFront)
- **WARP+ Outbound**: Cloudflare WARP integration for obfuscated connections
- **MTProto**: Telegram MTProto proxy generation with secret management
- **Clean IP Scanner**: Find clean Cloudflare IPs for uninterrupted access
- **Reality Scan**: TLS reality scanning with advanced fingerprinting

### 🔐 Security & Access Control

| Layer | Technology | Description |
|:-----:|:-----------|:------------|
| 🧑‍💼 | **RBAC** | Super Admin, Admin, Reseller with granular permissions |
| 🔑 | **API Tokens** | Scoped, delegated API access with expiration |
| 🌍 | **Geo-Blocking** | Country-based allow/block lists |
| 📱 | **TOTP/2FA** | Time-based one-time passwords |
| 📝 | **Audit Logs** | Complete security event logging |
| 🛡️ | **Rate Limiting** | 4 zones with per-zone cleanup |

### 📊 Analytics & Monitoring

- **Real-time Dashboard**: Live traffic metrics, online users, system performance
- **Prometheus Metrics**: Full Prometheus export for Grafana dashboards
- **Smart Health Check**: Configurable probes with auto-recovery rules
- **Network Topology**: Visual Canvas 2D topology map of cluster nodes
- **WebSocket Streaming**: Sub-second data push for real-time updates

### 💳 Payment & Billing

- **ZarinPal**: Iranian payment gateway (IRR/IRHT)
- **NOWPayments**: Cryptocurrency payments (BTC, ETH, USDT, +100 coins)
- **Plan Management**: Create/edit/delete subscription plans
- **Wallet System**: Internal wallet with deposit/withdraw/transfer

---

## 🏗️ Architecture

``` mermaid
graph TB
    subgraph "Reverse Proxy"
        Caddy["Caddy / Nginx<br/>TLS · Static · Rate Limit"]
    end

    subgraph "HTTP Layer"
        Gin["Gin Router<br/>REST API · WebSocket · Subscription"]
    end

    subgraph "Services"
        User["User Svc"]
        Inbound["Inbound Svc"]
        Outbound["Outbound Svc"]
        Sub["Subscription Svc"]
        Analytics["Analytics Svc"]
        Backup["Backup Svc"]
        Cluster["Cluster Svc"]
        Health["Health Check"]
    end

    subgraph "Core Engine"
        Xray["Xray gRPC API"]
        Singbox["Sing-box Config"]
        EngineMgr["Engine Manager<br/>Hot Reload · Fallback"]
    end

    subgraph "Data Layer"
        DB["SQLite / PostgreSQL"]
        Prom["Prometheus Metrics"]
    end

    Caddy --> Gin
    Gin --> User
    Gin --> Inbound
    Gin --> Outbound
    Gin --> Sub
    Gin --> Analytics
    Gin --> Backup
    Gin --> Cluster
    Gin --> Health
    Inbound --> Xray
    Outbound --> Xray
    Sub --> Singbox
    User --> DB
    Inbound --> DB
    Analytics --> Prom
    Xray --> EngineMgr
    Singbox --> EngineMgr
```

---

## 📦 Quick Stats

| Metric | Value |
|:-------|:------|
| **Backend Services** | 40+ |
| **Frontend Pages** | 39 |
| **Supported Languages** | 10 (EN, FA, ES, RU, ZH, AR, DE, FR, PT, TR) |
| **Proxy Protocols** | 7 |
| **API Endpoints** | 200+ |
| **Database** | SQLite / PostgreSQL |
| **License** | AGPL-3.0 |
| **Latest Version** | v0.1.0-beta |

---

## 🚀 Quick Start

```bash
# One-click install (Linux)
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh)

# Or with Docker
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro
docker compose -f deploy/compose.yml up -d
```

---

<div align="center">
    <a href="https://github.com/iPmartNetwork/VortexUiPro" class="md-button md-button--primary">
        :material-github: View on GitHub
    </a>
    <a href="quick-start.md" class="md-button">
        :material-rocket-launch: Get Started
    </a>
</div>
