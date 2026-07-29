# ─── VortexUiPro Web Server ──────────────────────────────────────────
# Multi-stage build: builds the Vite SPA and serves it with Caddy.
# Caddy also reverse-proxies API/WS endpoints to the panel backend.
#
# Build:
#   docker build -f deploy/web.Dockerfile -t vortexuipro/web .
#
# Run:
#   docker run -p 80:80 -p 443:443 \
#     -e SITE_ADDRESS=panel.example.com \
#     -e ACME_EMAIL=admin@example.com \
#     -e PANEL_UPSTREAM=panel:8080 \
#     vortexuipro/web
#
# ──────────────────────────────────────────────────────────────────────

# ── Stage 1: Build SPA ──────────────────────────────────────────────
FROM node:22-alpine AS build

WORKDIR /web

# Install dependencies (cache layer)
COPY web/package.json web/package-lock.json* ./
RUN npm ci --prefer-offline --no-audit

# Copy source and build
COPY web/ ./
COPY VERSION /VERSION

ARG VERSION=0.0.1
ENV VITE_APP_VERSION=$VERSION

RUN npm run build

# ── Stage 2: Serve with Caddy ──────────────────────────────────────
FROM caddy:2-alpine

# Security: run as non-root
RUN addgroup -S vortex && adduser -S vortex -G vortex

# Configuration
COPY deploy/Caddyfile /etc/caddy/Caddyfile

# Built SPA
COPY --from=build --chown=vortex:vortex /web/dist /usr/share/caddy

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
    CMD wget -qO- http://localhost/readyz || exit 1

USER vortex

EXPOSE 80 443

LABEL org.opencontainers.image.title="VortexUiPro Web"
LABEL org.opencontainers.image.description="Web server for VortexUiPro panel with automatic HTTPS"
LABEL org.opencontainers.image.version=${VERSION}
