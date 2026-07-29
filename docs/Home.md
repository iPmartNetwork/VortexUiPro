<div align="center">

# 🚀 VortexUiPro Wiki

**The Ultimate Proxy Management Panel** — *Next Generation • Enterprise Grade • Multi-Core*

---

[**📖 API Reference**](API-Reference) •
[**🚀 Quick Start**](Quick-Start) •
[**🐳 Deployment**](Deployment-Guide) •
[**⚙️ Configuration**](Configuration) •
[**🌐 Cluster Setup**](Cluster-Setup)

---

</div>

## 📚 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features Overview](#features-overview)
- [Quick Links](#quick-links)
- [Tech Stack](#tech-stack)

---

## Overview

**VortexUiPro** is a next-generation proxy management panel built on **Clean Architecture** principles. It combines the power of **Xray** and **Sing-Box** engines with an enterprise-grade management interface, supporting 7 proxy protocols, 10 languages, and multi-node clustering.

### Key Capabilities

| Capability | Description |
|:-----------|:------------|
| **40+ Backend Services** | Complete proxy lifecycle management from user creation to traffic analytics |
| **39 Frontend Pages** | Modern glassmorphism UI with dark/light themes, 10 languages |
| **7 Proxy Protocols** | VMess, VLESS, Trojan, Shadowsocks, Hysteria2, WireGuard, MTProto |
| **Multi-Node Cluster** | gRPC mesh with leader election, mTLS encryption, live topology visualization |
| **Anti-Censorship Suite** | TLS fragmentation, Domain Fronting, WARP+ integration, Smart DNS, Clean IP scanning |
| **Payment Gateways** | ZarinPal (IRR) + NOWPayments (crypto) with wallet system |
| **Enterprise Security** | RBAC (3-tier), TOTP 2FA, Audit Logs, IP whitelist/ban, Geo-blocking |

---

## Architecture

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
│                    Service Layer (40+ Services)              │
│  User • Inbound • Outbound • Subscription • Analytics        │
│  Backup • Cluster • Federation • Ticket • Plugin • Health   │
│  Terminal • Logs • WARP • TLS Tricks • WebRTC               │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────┴──────────────────────────────────────┐
│                    Core Engine Layer                         │
│  ┌──────────────────┐      ┌──────────────────┐             │
│  │   Xray gRPC API  │      │  Sing-box Config  │             │
│  │  (Stats + Route) │      │   (JSON Builder)  │             │
│  └────────┬─────────┘      └────────┬─────────┘             │
│           │                         │                       │
│  ┌────────┴─────────────────────────┴─────────┐             │
│  │        Engine Manager (Multi-Core)          │             │
│  └─────────────────────────────────────────────┘             │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────┴──────────────────────────────────────┐
│                    Data Layer                                │
│      SQLite / PostgreSQL • Prometheus Metrics • Redis        │
└─────────────────────────────────────────────────────────────┘
```

---

## Features Overview

### 🎯 Core Engine
- Dual-core: **Xray** + **Sing-Box** with hot-switch
- 40+ backend services for complete lifecycle management
- 7 proxy protocols with config generation
- Multi-node gRPC cluster

### 🛡️ Security & Access Control
| Feature | Details |
|:--------|:--------|
| **RBAC** | Super Admin → Admin → Reseller with granular permissions |
| **TOTP/2FA** | Time-based one-time passwords |
| **API Tokens** | Scoped, delegated API access with expiration |
| **Geo-Blocking** | Country-based allow/block lists |
| **IP Security** | Ban/whitelist with auto-ban |
| **Audit Logs** | Complete event logging with compliance export |
| **Rate Limiting** | 4 zones: auth (20/5min), api (300/min), sub (600/min), default (100/min) |

### 🌐 Cluster & Federation
- **Multi-Node Mesh** — gRPC with mTLS encryption
- **Leader Election** — Raft-style with priority-based voting
- **State Sync** — Users, inbounds, plans, traffic, bans
- **Federation** — Cross-panel user/plan/traffic sync
- **Live Topology** — Real-time Canvas visualization

### 📊 Analytics & Monitoring
- **Real-time Dashboard** — Live traffic, online users, system performance
- **Prometheus Metrics** — Full Grafana integration
- **WebSocket Streaming** — Sub-second data push
- **Smart Health Check** — Configurable probes + auto-recovery
- **Client Activity** — Per-user online tracking

### 💳 Payment & Billing
- **ZarinPal** — Iranian payment gateway (IRR/IRHT)
- **NOWPayments** — Cryptocurrency (BTC, ETH, USDT, +100 coins)
- **Wallet System** — Internal credit with deposit/withdraw/transfer
- **Plan Management** — Subscription plans with bandwidth limits

---

## Quick Links

| Resource | Link |
|:---------|:-----|
| **📖 Full API Reference** | [API-Reference](API-Reference) |
| **🚀 Quick Start Guide** | [Quick-Start](Quick-Start) |
| **🐳 Deployment Guide** | [Deployment-Guide](Deployment-Guide) |
| **⚙️ Configuration** | [Configuration](Configuration) |
| **🌐 Cluster Setup** | [Cluster-Setup](Cluster-Setup) |
| **📦 GitHub Repository** | [iPmartNetwork/VortexUiPro](https://github.com/iPmartNetwork/VortexUiPro) |
| **🐛 Bug Reports** | [GitHub Issues](https://github.com/iPmartNetwork/VortexUiPro/issues) |
| **💬 Discussions** | [GitHub Discussions](https://github.com/iPmartNetwork/VortexUiPro/discussions) |

---

## Tech Stack

| Layer | Technology | Version |
|:------|:-----------|:-------:|
| **Backend** | Go (Gin) | 1.23+ |
| **Frontend** | React + TypeScript | 18.x |
| **Styling** | Tailwind CSS | 3.x |
| **Database** | SQLite / PostgreSQL | — |
| **Cache** | Redis (optional) | 7.x |
| **Core Engine** | Xray, Sing-Box | Latest |
| **Proxy** | Caddy (auto HTTPS) | 2.x |
| **Container** | Docker + Docker Compose | Latest |
| **Orchestration** | Kubernetes Helm Chart | — |
| **Monitoring** | Prometheus + Grafana | Latest |
| **Metrics** | GoReleaser + Trivy + Syft | Latest |

---

<div align="center">

*Made with ❤️ by the VortexUiPro Team*

</div>
