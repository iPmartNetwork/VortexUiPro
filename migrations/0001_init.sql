-- VortexUiPro Initial Schema v0.0.1
-- Supports SQLite and PostgreSQL

-- Admins (panel operators)
CREATE TABLE IF NOT EXISTS admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email VARCHAR(128),
    role VARCHAR(32) NOT NULL DEFAULT 'admin',
    totp_secret TEXT,
    totp_enabled BOOLEAN DEFAULT FALSE,
    login_attempts INTEGER DEFAULT 0,
    locked_until INTEGER DEFAULT 0,
    api_key TEXT,
    api_token_hash TEXT,
    traffic_limit BIGINT DEFAULT 0,
    user_limit INTEGER DEFAULT 0,
    inbound_limit INTEGER DEFAULT 0,
    commission_rate INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Users (subscribers)
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id INTEGER REFERENCES admins(id),
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash TEXT,
    email VARCHAR(128),
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    traffic_up BIGINT DEFAULT 0,
    traffic_down BIGINT DEFAULT 0,
    traffic_total BIGINT DEFAULT 0,
    data_limit BIGINT DEFAULT 0,
    expiry_time INTEGER DEFAULT 0,
    device_limit INTEGER DEFAULT 0,
    speed_limit_up INTEGER DEFAULT 0,
    speed_limit_down INTEGER DEFAULT 0,
    note TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Proxy Inbounds
CREATE TABLE IF NOT EXISTS inbounds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id),
    node_id INTEGER,
    tag VARCHAR(64) NOT NULL,
    protocol VARCHAR(32) NOT NULL,
    listen_ip VARCHAR(64) DEFAULT '0.0.0.0',
    port INTEGER NOT NULL,
    status VARCHAR(16) DEFAULT 'active',
    stream_settings TEXT,
    settings TEXT,
    sniffing TEXT,
    allocate TEXT,
    remark VARCHAR(128),
    up_mbps INTEGER DEFAULT 0,
    down_mbps INTEGER DEFAULT 0,
    total_gb BIGINT DEFAULT 0,
    expiry_time INTEGER DEFAULT 0,
    enable BOOLEAN DEFAULT TRUE,
    subscription_id VARCHAR(64),
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Subscription Profiles (multi-profile per inbound - Heimdall feature)
CREATE TABLE IF NOT EXISTS subscription_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    inbound_id INTEGER NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    dest VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    remark VARCHAR(128),
    enabled BOOLEAN DEFAULT TRUE,
    network VARCHAR(32),
    security VARCHAR(32),
    tls_settings TEXT,
    reality_settings TEXT,
    sockopt TEXT,
    mux_config TEXT,
    sni VARCHAR(255),
    alpn VARCHAR(255),
    fingerprint VARCHAR(64),
    created_at INTEGER NOT NULL
);

-- Proxy Outbounds
CREATE TABLE IF NOT EXISTS outbounds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id INTEGER,
    tag VARCHAR(64) NOT NULL,
    protocol VARCHAR(32) NOT NULL,
    settings TEXT,
    stream_settings TEXT,
    remark VARCHAR(128),
    enable BOOLEAN DEFAULT TRUE,
    hidden BOOLEAN DEFAULT FALSE,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Routing Rules
CREATE TABLE IF NOT EXISTS routing_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    inbound_tags TEXT,
    outbound_tag VARCHAR(64) NOT NULL,
    domain TEXT,
    ip TEXT,
    port TEXT,
    network VARCHAR(16),
    protocol TEXT,
    geoip TEXT,
    geosite TEXT,
    source_ip TEXT,
    source_port TEXT,
    user_email TEXT,
    balancer_tag VARCHAR(64),
    rule_type VARCHAR(32),
    enabled BOOLEAN DEFAULT TRUE,
    created_at INTEGER NOT NULL
);

-- Proxy Nodes
CREATE TABLE IF NOT EXISTS nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(128) NOT NULL,
    address VARCHAR(255) NOT NULL,
    port INTEGER DEFAULT 0,
    api_port INTEGER DEFAULT 10085,
    status VARCHAR(16) DEFAULT 'offline',
    core_type VARCHAR(16) DEFAULT 'xray',
    enable BOOLEAN DEFAULT TRUE,
    country VARCHAR(64),
    location VARCHAR(128),
    cpu_load REAL DEFAULT 0,
    memory_used REAL DEFAULT 0,
    uplink BIGINT DEFAULT 0,
    downlink BIGINT DEFAULT 0,
    traffic_up BIGINT DEFAULT 0,
    traffic_down BIGINT DEFAULT 0,
    last_heartbeat INTEGER,
    remark VARCHAR(255),
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Panel Settings
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    port INTEGER DEFAULT 8080,
    listen_ip VARCHAR(64) DEFAULT '0.0.0.0',
    base_path VARCHAR(128) DEFAULT '/',
    cert_file TEXT,
    key_file TEXT,
    sub_port INTEGER DEFAULT 2087,
    sub_path VARCHAR(128) DEFAULT '/sub',
    sub_cert_file TEXT,
    sub_key_file TEXT,
    sub_domain VARCHAR(255),
    sub_enable BOOLEAN DEFAULT TRUE,
    sub_json_rules TEXT,
    sub_clash_rules TEXT,
    sub_mux TEXT,
    sub_final_mask TEXT,
    sub_enable_routing BOOLEAN DEFAULT TRUE,
    totp_enabled BOOLEAN DEFAULT FALSE,
    totp_token TEXT,
    telegram_token TEXT,
    telegram_chat_id TEXT,
    telegram_runtime VARCHAR(64),
    telegram_enabled BOOLEAN DEFAULT FALSE,
    webhook_url TEXT,
    webhook_secret TEXT,
    webhook_enabled BOOLEAN DEFAULT FALSE,
    default_core VARCHAR(16) DEFAULT 'xray',
    enable_tunnel_monitor BOOLEAN DEFAULT FALSE,
    tunnel_monitor_url TEXT,
    tunnel_monitor_proxy TEXT,
    auto_restart_core BOOLEAN DEFAULT TRUE,
    log_level VARCHAR(16) DEFAULT 'info',
    language VARCHAR(16) DEFAULT 'en',
    brand_name VARCHAR(128),
    brand_website VARCHAR(255),
    brand_logo TEXT
);

-- Clients (proxy users)
CREATE TABLE IF NOT EXISTS clients (
    id VARCHAR(64) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    inbound_id INTEGER REFERENCES inbounds(id),
    email VARCHAR(128) NOT NULL UNIQUE,
    enable BOOLEAN DEFAULT TRUE,
    flow VARCHAR(32),
    password TEXT,
    security VARCHAR(32),
    total_gb BIGINT DEFAULT 0,
    expiry_time INTEGER DEFAULT 0,
    sub_id VARCHAR(64),
    up_mbps INTEGER DEFAULT 0,
    down_mbps INTEGER DEFAULT 0,
    private_key TEXT,
    public_key TEXT,
    pre_shared_key TEXT,
    allowed_ips TEXT,
    keep_alive INTEGER DEFAULT 0,
    secret TEXT,
    ad_tag TEXT
);

-- Notification Channels
CREATE TABLE IF NOT EXISTS notification_channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type VARCHAR(32) NOT NULL,
    name VARCHAR(128) NOT NULL,
    token TEXT,
    chat_id VARCHAR(64),
    webhook_url TEXT,
    webhook_secret TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at INTEGER NOT NULL
);

-- Event Subscriptions (link channels to event types)
CREATE TABLE IF NOT EXISTS event_subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id INTEGER NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    events TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at INTEGER NOT NULL
);

-- Security Events
CREATE TABLE IF NOT EXISTS security_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type VARCHAR(32) NOT NULL,
    source_ip VARCHAR(64),
    username VARCHAR(64),
    detail TEXT,
    level VARCHAR(16) DEFAULT 'info',
    node_id INTEGER,
    created_at INTEGER NOT NULL
);

-- Clean IP Scan Results
CREATE TABLE IF NOT EXISTS clean_ip_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id INTEGER NOT NULL,
    ip VARCHAR(64) NOT NULL,
    port INTEGER NOT NULL,
    protocol VARCHAR(16),
    latency_ms INTEGER,
    is_clean BOOLEAN DEFAULT FALSE,
    checked_at INTEGER NOT NULL
);

-- Plans (service packages)
CREATE TABLE IF NOT EXISTS plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    price BIGINT NOT NULL,
    data_limit BIGINT DEFAULT 0,
    speed_limit INTEGER DEFAULT 0,
    device_limit INTEGER DEFAULT 0,
    duration INTEGER DEFAULT 0,
    protocol VARCHAR(32),
    node_group VARCHAR(64),
    enabled BOOLEAN DEFAULT TRUE,
    created_at INTEGER NOT NULL
);

-- Orders (user purchases)
CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    plan_id INTEGER NOT NULL REFERENCES plans(id),
    amount BIGINT NOT NULL,
    status VARCHAR(16) DEFAULT 'pending',
    proof_file TEXT,
    created_at INTEGER NOT NULL,
    paid_at INTEGER
);

-- Wallet Transactions
CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL,
    type VARCHAR(16) NOT NULL,
    description TEXT,
    reference_id VARCHAR(64),
    status VARCHAR(16) DEFAULT 'pending',
    created_at INTEGER NOT NULL
);

-- Support Tickets
CREATE TABLE IF NOT EXISTS tickets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    subject VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    status VARCHAR(16) DEFAULT 'open',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS ticket_replies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ticket_id INTEGER NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    message TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_admin_id ON users(admin_id);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_inbounds_node_id ON inbounds(node_id);
CREATE INDEX IF NOT EXISTS idx_inbounds_tag ON inbounds(tag);
CREATE INDEX IF NOT EXISTS idx_outbounds_tag ON outbounds(tag);
CREATE INDEX IF NOT EXISTS idx_clients_user_id ON clients(user_id);
CREATE INDEX IF NOT EXISTS idx_clients_email ON clients(email);
CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);
CREATE INDEX IF NOT EXISTS idx_nodes_last_heartbeat ON nodes(last_heartbeat);
CREATE INDEX IF NOT EXISTS idx_security_events_type ON security_events(type);
CREATE INDEX IF NOT EXISTS idx_security_events_created ON security_events(created_at);
CREATE INDEX IF NOT EXISTS idx_transactions_user ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_tickets_user ON tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
