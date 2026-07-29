# 🐳 Deployment Guide

> Complete deployment options for VortexUiPro — from single-server to Kubernetes.

---

## 📋 Table of Contents

- [System Requirements](#system-requirements)
- [Native Installation](#native-installation)
- [Docker Compose](#docker-compose)
- [Production Docker Compose](#production-docker-compose)
- [Kubernetes (Helm)](#kubernetes-helm)
- [SSL Setup](#ssl-setup)
- [Database Options](#database-options)
- [Monitoring Setup](#monitoring-setup)

---

## 📊 System Requirements

| Component | Minimum | Recommended | Production |
|:---------:|:-------:|:-----------:|:----------:|
| **CPU** | 1 core | 2 cores | 4+ cores |
| **RAM** | 512 MB | 1 GB | 4+ GB |
| **Storage** | 5 GB | 10 GB | 50+ GB |
| **OS** | Linux | Ubuntu 22.04+ | Debian 12 |
| **Xray Core** | v1.8.0 | Latest | Latest |
| **Database** | SQLite | PostgreSQL | PostgreSQL Cluster |
| **Node.js** | — | — | 22+ (build only) |

---

## 🐧 Supported Distributions

| Distribution | Support | Notes |
|:-----------:|:-------:|:------|
| **Ubuntu** 20.04+ | ✅ Full | Recommended for beginners |
| **Debian** 11+ | ✅ Full | Recommended for production |
| **CentOS** 7+ | ✅ Full | AlmaLinux/Rocky |
| **Fedora** 37+ | ✅ Full | — |
| **Arch Linux** | ✅ Full | — |
| **Alpine** 3.18+ | ✅ Docker | Docker-only |

---

## 🏠 Native Installation

### Step 1: Run the installer

```bash
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh)
```

### Step 2: Configure

```bash
# Edit config at /etc/vortexuipro/env
nano /etc/vortexuipro/env
```

### Step 3: Start

```bash
systemctl start vortexuipro
systemctl enable vortexuipro
```

---

## 🐳 Docker Compose

### Single Node

```bash
# Clone
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro

# Start single node
docker compose -f deploy/compose.yml up -d node-1

# Verify
curl http://localhost:8080/api/v1/health
```

### Multi-Node Cluster (3 Nodes)

```yaml
# deploy/compose.yml
services:
  node-1:
    build: .
    ports:
      - "8080:8080"     # HTTP API
      - "1337:1337"     # Cluster mesh
    environment:
      VORTEX_HTTP_ADDR: ":8080"
      VORTEX_CLUSTER_ENABLED: "true"
      VORTEX_CLUSTER_NODE_NAME: "node-1"
      VORTEX_CLUSTER_ADDR: ":1337"
      VORTEX_CLUSTER_PEERS: "node-2:1337,node-3:1337"
      VORTEX_CLUSTER_PRIORITY: "200"
    volumes:
      - vortex-data-1:/etc/vortex/data
      - vortex-certs:/etc/vortex/certs

  node-2:
    build: .
    ports:
      - "8081:8080"
      - "1338:1337"
    environment:
      VORTEX_HTTP_ADDR: ":8080"
      VORTEX_CLUSTER_ENABLED: "true"
      VORTEX_CLUSTER_NODE_NAME: "node-2"
      VORTEX_CLUSTER_ADDR: ":1337"
      VORTEX_CLUSTER_PEERS: "node-1:1337,node-3:1337"
      VORTEX_CLUSTER_PRIORITY: "100"
    volumes:
      - vortex-data-2:/etc/vortex/data
```

---

## 🏭 Production Docker Compose

See `deploy/production.yml` for a production-ready single-node setup:

```bash
docker compose -f deploy/production.yml up -d
```

Features:
- **Health checks** on all services
- **Log rotation** with max-size/max-file
- **Resource limits** (CPU/Memory)
- **Restart policy**: always
- **Auto backup** cron job
- **Prometheus + Grafana** monitoring

---

## ☸️ Kubernetes (Helm)

### Prerequisites

```bash
# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Add Bitnami repo
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

### Install

```bash
# Update dependencies
helm dependency update deploy/chart

# Install VortexUiPro
helm install vortexuipro deploy/chart \
  --namespace vortexuipro --create-namespace \
  --set panel.secrets.jwtSecret="$(openssl rand -hex 32)" \
  --set panel.secrets.databasePassword="$(openssl rand -hex 16)" \
  --set web.ingress.hosts[0].host="panel.yourdomain.com"
```

### Upgrade

```bash
helm upgrade vortexuipro deploy/chart \
  --namespace vortexuipro \
  --set image.tag="v0.1.0-beta"
```

### Configuration Options

| Parameter | Default | Description |
|:----------|:--------|:------------|
| `panel.replicaCount` | `2` | Number of panel replicas |
| `panel.autoscaling.enabled` | `false` | Enable HPA |
| `panel.autoscaling.minReplicas` | `2` | Minimum pods |
| `panel.autoscaling.maxReplicas` | `10` | Maximum pods |
| `panel.autoscaling.targetCPU` | `70` | CPU target % |
| `panel.persistence.size` | `8Gi` | Data volume size |
| `postgresql.enabled` | `true` | Deploy PostgreSQL |
| `postgresql.auth.database` | `vortex` | Database name |
| `redis.enabled` | `true` | Deploy Redis |
| `prometheus.enabled` | `false` | Deploy Prometheus |
| `grafana.enabled` | `false` | Deploy Grafana |

---

## 🔒 SSL Setup

### Option 1: Let's Encrypt (Domain Required)

```bash
vortexui cert
# Choose option 1: Let's Encrypt
# Enter your domain: panel.yourdomain.com
# Enter email: admin@yourdomain.com
```

### Option 2: Let's Encrypt (IP Address)

```bash
vortexui cert
# Choose option 2: Let's Encrypt IP
```

### Option 3: Custom Certificate

```bash
vortexui cert
# Choose option 3: Custom
# Enter certificate path: /path/to/cert.pem
# Enter key path: /path/to/key.pem
```

### Option 4: Caddy Auto-HTTPS (Docker)

Use the provided `deploy/web.Dockerfile` + `deploy/Caddyfile`:

```bash
docker compose -f deploy/compose.yml -f deploy/caddy-compose.yml up -d
```

---

## 🗄️ Database Options

### SQLite (Default — Good for small/medium deployments)

```bash
# Default configuration — no setup needed
VORTEX_DB_TYPE=sqlite
VORTEX_DATABASE_URL=/etc/vortex/data/vortex.db
```

### PostgreSQL (Recommended for production)

```bash
# 1. Install PostgreSQL
apt install postgresql -y

# 2. Create database
sudo -u postgres psql -c "CREATE DATABASE vortex;"
sudo -u postgres psql -c "CREATE USER vortex WITH PASSWORD 'vortex123';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE vortex TO vortex;"

# 3. Configure VortexUiPro
VORTEX_DB_TYPE=postgres
VORTEX_DATABASE_URL=postgres://vortex:vortex123@localhost:5432/vortex?sslmode=disable
```

---

## 📊 Monitoring Setup

### Prometheus + Grafana (Docker)

```bash
# Start monitoring stack
docker compose -f deploy/compose.yml -f deploy/monitoring.yml up -d

# Access:
# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9090
```

### Import Grafana Dashboard

1. Open Grafana at `http://localhost:3000`
2. Login with `admin` / `admin`
3. Go to **Dashboards → Import**
4. Upload `deploy/grafana/vortexuipro-dashboard.json`
5. Select Prometheus data source

### Available Metrics

| Metric | Type | Description |
|:-------|:----:|:------------|
| `vortex_online_users` | Gauge | Current online users |
| `vortex_total_users` | Gauge | Total registered users |
| `vortex_traffic_bytes_total` | Counter | Total traffic (bytes) |
| `vortex_traffic_rate_bytes` | Gauge | Traffic rate (bytes/sec) |
| `vortex_api_requests_total` | Counter | API request count |
| `vortex_api_request_duration` | Histogram | API latency |
| `vortex_inbounds_total` | Gauge | Total inbounds |
| `vortex_nodes_total` | Gauge | Total nodes |
| `vortex_cluster_leader` | Gauge | Current cluster leader |

---

## 🔄 Backup & Restore

### Manual Backup

```bash
vortexui backup
# Creates: /etc/vortex/backups/backup-YYYY-MM-DD-HHMMSS.tar.gz
```

### Auto Backup

```bash
# Configure automatic daily backups
curl -X POST http://localhost:8080/api/v1/backups/auto-config \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "interval_hours": 24,
    "retention_days": 7,
    "encrypt": true,
    "remote_storage_id": 1
  }'
```

### Restore

```bash
vortexui restore /etc/vortex/backups/backup-2025-07-29-120000.tar.gz
```

---

## 📝 Environment Variables

See [Configuration](Configuration) for the complete reference.

---

<div align="center">

*Last updated: 2025-07-29 • VortexUiPro v0.1.0-beta*

</div>
