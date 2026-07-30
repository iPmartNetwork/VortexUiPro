# ═══════════════════════════════════════════════════════════════════════
# VortexUiPro — Dockerfile 🐳
#
# Multi-stage build:
#   frontend-builder  — Builds the React SPA (for panel target)
#   backend-builder   — Builds the panel binary (depends on frontend)
#   agent-builder     — Builds the lightweight agent binary (no frontend!)
#   panel             — Production runtime image for the panel
#   node              — Lightweight runtime image for the node agent
#
# Usage:
#   docker build --target panel -t vortexuipro-panel .
#   docker build --target node -t vortexuipro-node .
#   docker build --target web -f deploy/web.Dockerfile -t vortexuipro-web .
#
# ═══════════════════════════════════════════════════════════════════════

# ── Stage 1: Frontend (only needed for panel) ──────────────────────
FROM node:22-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm ci --prefer-offline --no-audit
COPY web/ ./
ARG VERSION=0.0.1
ENV VITE_APP_VERSION=$VERSION
RUN npm run build

# ── Stage 2: Backend (panel binary, needs web dist) ────────────────
# Use --platform=$BUILDPLATFORM so the builder runs native (amd64)
# and cross-compiles for the target platform.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS backend-builder
WORKDIR /app
ARG TARGETARCH
ARG TARGETOS=linux
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/dist ./web/dist
ARG VERSION=0.0.1
# CGO_ENABLED=0 — avoids CGO cross-compilation issues in multi-arch builds
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w -X main.version=${VERSION}" \
    -o /usr/local/bin/vortexuipro ./cmd/panel

# ── Stage 3: Agent (lightweight node binary, NO frontend deps!) ────
FROM golang:1.25-alpine AS agent-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY internal/ ./internal/
COPY cmd/agent/ ./cmd/agent/
ARG VERSION=0.0.1
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o /usr/local/bin/vortexuipro-agent ./cmd/agent

# ── Stage 4: Panel Runtime ──────────────────────────────────────────
FROM alpine:3.20 AS panel
RUN apk add --no-cache ca-certificates tzdata bash curl

WORKDIR /etc/vortex

# Copy binary
COPY --from=backend-builder /usr/local/bin/vortexuipro /usr/local/bin/vortexuipro

# Create data directories
RUN mkdir -p /etc/vortex/data /etc/vortex/certs /etc/vortex/mtproto /etc/vortex/backups

# Environment defaults
ENV VORTEX_HTTP_ADDR=:8080
ENV VORTEX_DB_TYPE=sqlite
ENV VORTEX_DATABASE_URL=/etc/vortex/data/vortex.db
ENV VORTEX_LOG_LEVEL=info
ENV VORTEX_CORE=xray
ENV VORTEX_LOCAL_NODE=true
ENV VORTEX_AUTO_RECOVER=true

EXPOSE 8080 50051 2087

VOLUME ["/etc/vortex/data", "/etc/vortex/certs", "/etc/vortex/mtproto"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -sf http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["vortexuipro"]

LABEL org.opencontainers.image.title="VortexUiPro Panel"
LABEL org.opencontainers.image.description="Next-generation proxy management panel"
LABEL org.opencontainers.image.version="${VERSION}"

# ── Stage 5: Node Agent Runtime (lightweight, no panel dependencies) ──
FROM alpine:3.20 AS node
RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /etc/vortex

# Copy agent binary (built independently from frontend!)
COPY --from=agent-builder /usr/local/bin/vortexuipro-agent /usr/local/bin/vortexuipro-agent

# Create minimal data directory
RUN mkdir -p /etc/vortex/data

# Node-specific defaults
ENV NODE_GRPC_ADDR=:50051

EXPOSE 50051

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -sf http://localhost:50051/health || exit 1

ENTRYPOINT ["vortexuipro-agent"]

LABEL org.opencontainers.image.title="VortexUiPro Node Agent"
LABEL org.opencontainers.image.description="Lightweight node agent for distributed proxy management"
LABEL org.opencontainers.image.version="${VERSION}"
