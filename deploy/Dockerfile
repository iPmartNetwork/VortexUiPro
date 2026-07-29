# -- Build Frontend --
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm ci
COPY web/ .
RUN npm run build

# -- Build Backend --
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=1 go build -o /usr/local/bin/vortexuipro ./cmd/panel

# -- Node Agent --
FROM alpine:3.20 AS node
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend-builder /usr/local/bin/vortexuipro /usr/local/bin/vortexuipro
ENV NODE_GRPC_ADDR=:50051
EXPOSE 50051
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
    CMD wget -qO- http://localhost:50051/health || exit 1
CMD ["vortexuipro", "--node"]

# -- Runtime (Panel) --
FROM alpine:3.20 AS panel
RUN apk add --no-cache ca-certificates tzdata bash curl

WORKDIR /etc/vortex

# Copy binary
COPY --from=backend-builder /usr/local/bin/vortexuipro /usr/local/bin/vortexuipro

# Create data directories
RUN mkdir -p /etc/vortex/data /etc/vortex/certs /etc/vortex/mtproto /etc/vortex/backups

# Environment variables
ENV VORTEX_HTTP_ADDR=:8080
ENV VORTEX_DB_TYPE=sqlite
ENV VORTEX_DATABASE_URL=/etc/vortex/data/vortex.db
ENV VORTEX_LOG_LEVEL=info
ENV VORTEX_CORE=xray
ENV VORTEX_LOCAL_NODE=true
ENV VORTEX_AUTO_RECOVER=true

EXPOSE 8080 50051 2087

VOLUME ["/etc/vortex/data", "/etc/vortex/certs", "/etc/vortex/mtproto"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
    CMD curl -sf http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["vortexuipro"]
