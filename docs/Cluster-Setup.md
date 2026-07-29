# 🌐 Cluster Setup Guide

> Set up a multi-node VortexUiPro cluster with gRPC mesh, leader election, and mTLS.

---

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Quick Start — 3-Node Docker Cluster](#quick-start--3-node-docker-cluster)
- [Manual Cluster Setup](#manual-cluster-setup)
- [Cluster Configuration](#cluster-configuration)
- [mTLS Setup](#mtls-setup)
- [Verification & Monitoring](#verification--monitoring)
- [Troubleshooting](#troubleshooting)

---

## 🏗️ Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                        Cluster Mesh (gRPC + mTLS)                 │
│                                                                   │
│   ┌─────────────┐    Heartbeat    ┌─────────────┐               │
│   │   Node 1    │◄──────────────►│   Node 2    │               │
│   │  (Leader)   │    Sync        │  (Follower) │               │
│   │  Priority:  │◄──────────────►│  Priority:  │               │
│   │    200      │    Sync        │    100      │               │
│   └──────┬──────┘                └──────┬──────┘               │
│          │                              │                        │
│   ┌──────┴──────┐                ┌──────┴──────┐               │
│   │   Node 3    │                │   Node 4    │               │
│   │  (Follower) │◄──────────────►│  (Follower) │               │
│   │  Priority:  │    Sync        │  Priority:  │               │
│   │     50      │                │     25      │               │
│   └─────────────┘                └─────────────┘               │
│                                                                   │
│   Leader Election: Raft-style with priority-based voting          │
│   State Sync: Users, Inbounds, Plans, Traffic, Bans              │
│   Security: mTLS with auto-generated or custom certificates      │
│   Heartbeat: 5s interval, 3 missed = node marked offline         │
└──────────────────────────────────────────────────────────────────┘
```

### Key Concepts

| Concept | Description |
|:--------|:------------|
| **Leader** | The node with highest priority, handles write operations |
| **Follower** | Read-only nodes, replicate data from leader |
| **Election** | When leader goes offline, highest-priority follower becomes leader |
| **Heartbeat** | 5-second health checks between nodes |
| **Sync** | State replication (users, inbounds, traffic) every 30 seconds |
| **mTLS** | Mutual TLS encryption for all cluster communication |

---

## ✅ Prerequisites

- 2+ servers/nodes with VortexUiPro installed
- Open ports between nodes:
  - `1337/tcp` — Cluster mesh (gRPC)
  - `8080/tcp` — HTTP API (optional, for multi-node management)
- Shared clock synchronization (NTP recommended)
- mTLS certificates (auto-generated or custom)

---

## 🚀 Quick Start — 3-Node Docker Cluster

```bash
# Clone the repository
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro

# Start all 3 nodes
docker compose -f deploy/compose.yml up -d

# Verify
curl http://localhost:8080/api/v1/cluster/status
```

### Docker Compose Configuration

```yaml
# deploy/compose.yml
version: "3.8"

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
      VORTEX_CLUSTER_REGION: "us-east"
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
      VORTEX_CLUSTER_REGION: "us-east"
      VORTEX_CLUSTER_PRIORITY: "100"
    volumes:
      - vortex-data-2:/etc/vortex/data

  node-3:
    build: .
    ports:
      - "8082:8080"
      - "1339:1337"
    environment:
      VORTEX_HTTP_ADDR: ":8080"
      VORTEX_CLUSTER_ENABLED: "true"
      VORTEX_CLUSTER_NODE_NAME: "node-3"
      VORTEX_CLUSTER_ADDR: ":1337"
      VORTEX_CLUSTER_PEERS: "node-1:1337,node-2:1337"
      VORTEX_CLUSTER_REGION: "us-west"
      VORTEX_CLUSTER_PRIORITY: "50"
    volumes:
      - vortex-data-3:/etc/vortex/data

volumes:
  vortex-data-1:
  vortex-data-2:
  vortex-data-3:
  vortex-certs:
```

---

## 🔧 Manual Cluster Setup

### Step 1: Configure mTLS

```bash
# On the leader node, generate cluster certificates
vortexui cluster init-pki

# Copy CA certificate to all nodes
scp /etc/vortex/certs/ca.pem node-2:/etc/vortex/certs/
scp /etc/vortex/certs/ca.pem node-3:/etc/vortex/certs/

# On each follower node, generate their certificate
vortexui cluster join --leader node-1:1337
```

### Step 2: Configure Each Node

**Node 1 (Leader Candidate)**:
```bash
# /etc/vortexuipro/env
VORTEX_CLUSTER_ENABLED=true
VORTEX_CLUSTER_NODE_NAME=node-1
VORTEX_CLUSTER_ADDR=:1337
VORTEX_CLUSTER_PEERS=node-2:1337,node-3:1337
VORTEX_CLUSTER_SECRET=your-cluster-secret
VORTEX_CLUSTER_TLS_ENABLED=true
VORTEX_CLUSTER_REGION=us-east
VORTEX_CLUSTER_PRIORITY=200
```

**Node 2 (Follower)**:
```bash
# /etc/vortexuipro/env
VORTEX_CLUSTER_ENABLED=true
VORTEX_CLUSTER_NODE_NAME=node-2
VORTEX_CLUSTER_ADDR=:1337
VORTEX_CLUSTER_PEERS=node-1:1337,node-3:1337
VORTEX_CLUSTER_SECRET=your-cluster-secret
VORTEX_CLUSTER_TLS_ENABLED=true
VORTEX_CLUSTER_REGION=us-east
VORTEX_CLUSTER_PRIORITY=100
```

### Step 3: Start Nodes

```bash
# Start each node
systemctl start vortexuipro
systemctl enable vortexuipro

# Verify cluster status
curl http://localhost:8080/api/v1/cluster/status
```

---

## 🔐 mTLS Setup

### Auto-Generated (Default)

VortexUiPro automatically generates mTLS certificates on first cluster startup:

```bash
# Check PKI status
curl http://localhost:8080/api/v1/cluster/pki
```

**Response**:
```json
{
  "ca_cert": "/etc/vortex/certs/ca.pem",
  "node_cert": "/etc/vortex/certs/node.pem",
  "node_key": "/etc/vortex/certs/node-key.pem",
  "expires_at": "2026-07-29T00:00:00Z",
  "auto_generated": true
}
```

### Custom Certificates

```bash
# 1. Generate CA
openssl genrsa -out ca-key.pem 4096
openssl req -new -x509 -days 3650 -key ca-key.pem -out ca.pem -subj "/CN=VortexUiPro Cluster CA"

# 2. Generate node certificate
openssl genrsa -out node-key.pem 2048
openssl req -new -key node-key.pem -out node.csr -subj "/CN=node-1"
openssl x509 -req -days 3650 -in node.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out node.pem

# 3. Copy to all nodes
cp ca.pem /etc/vortex/certs/
cp node.pem /etc/vortex/certs/node-1.pem
cp node-key.pem /etc/vortex/certs/node-1-key.pem

# 4. Configure
VORTEX_CLUSTER_TLS_CERT=/etc/vortex/certs/node-1.pem
VORTEX_CLUSTER_TLS_KEY=/etc/vortex/certs/node-1-key.pem
VORTEX_CLUSTER_CA_CERT=/etc/vortex/certs/ca.pem
```

---

## 📊 Verification & Monitoring

### Cluster Status

```bash
# Overall cluster health
curl http://localhost:8080/api/v1/cluster/status

# List all nodes
curl http://localhost:8080/api/v1/cluster/nodes

# Election statistics
curl http://localhost:8080/api/v1/cluster/election

# Sync events log
curl http://localhost:8080/api/v1/cluster/sync-events

# Topology visualization data
curl http://localhost:8080/api/v1/cluster/topology
```

**Cluster Status Response**:
```json
{
  "cluster_name": "vortex-main",
  "node_count": 3,
  "leader": "node-1",
  "leader_address": "192.168.1.10:1337",
  "term": 3,
  "healthy_nodes": 3,
  "total_users": 1500,
  "total_traffic_gb": 245.3,
  "uptime": "2d15h30m"
}
```

### Web UI

Access `http://<panel-address>:8080` and navigate to **Cluster** page for a visual topology map.

---

## ⚡ Leader Election

### How It Works

1. **Startup**: All nodes start as followers
2. **Election Trigger**: When no heartbeat from leader for 15 seconds
3. **Voting**: Nodes vote for the candidate with highest priority
4. **Leader**: The winner starts accepting writes
5. **Re-election**: If leader goes offline, process repeats

### Force Election

```bash
# Force a new election (useful for maintenance)
curl -X POST http://localhost:8080/api/v1/cluster/election/force \
  -H "Authorization: Bearer <token>"
```

---

## 🩺 Troubleshooting

### Node not connecting

```bash
# 1. Check if mesh port is open
nc -zv node-2 1337

# 2. Check cluster logs
journalctl -u vortexuipro | grep -i cluster | tail -50

# 3. Verify mTLS certificates
openssl x509 -in /etc/vortex/certs/node.pem -text -noout
```

### Split brain

If two nodes both think they're leader:

```bash
# 1. Force re-election on all nodes
curl -X POST http://localhost:8080/api/v1/cluster/election/force

# 2. If that doesn't work, restart all nodes
systemctl restart vortexuipro

# 3. Last resort: reset cluster state on all but the primary node
vortexui cluster reset
```

### Sync issues

```bash
# Check sync events
curl http://localhost:8080/api/v1/cluster/sync-events

# Trigger manual sync
curl -X POST http://localhost:8080/api/v1/cluster/sync/manual \
  -H "Authorization: Bearer <token>"
```

### Performance

```bash
# Adjust heartbeat interval (lower = faster detection, higher = less traffic)
VORTEX_CLUSTER_HEARTBEAT=3

# Adjust sync interval (lower = more real-time, higher = less I/O)
VORTEX_CLUSTER_SYNC_INTERVAL=15
```

---

## 📋 Federation vs Cluster

| Feature | Cluster | Federation |
|:--------|:--------|:-----------|
| **Sync Scope** | Full state | Selected data (users, plans, traffic) |
| **Latency** | <10ms (LAN) | <500ms (WAN) |
| **Authentication** | mTLS + secret | API key |
| **Use Case** | Multi-node within datacenter | Cross-datacenter / partner panels |
| **Consistency** | Strong | Eventual |
| **Write Operations** | Leader only | Both sides |

---

<div align="center">

*Last updated: 2025-07-29 • VortexUiPro v0.1.0-beta*

</div>
