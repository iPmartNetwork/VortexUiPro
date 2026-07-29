package domain

// PanelSettings represents configurable panel settings.
type PanelSettings struct {
	ID                  int64  `json:"id" db:"id"`
	Port                int    `json:"port" db:"port"`
	ListenIP            string `json:"listen_ip,omitempty" db:"listen_ip"`
	BasePath            string `json:"base_path,omitempty" db:"base_path"`
	CertFile            string `json:"cert_file,omitempty" db:"cert_file"`
	KeyFile             string `json:"key_file,omitempty" db:"key_file"`
	SubPort             int    `json:"sub_port,omitempty" db:"sub_port"`
	SubPath             string `json:"sub_path,omitempty" db:"sub_path"`
	SubCertFile         string `json:"sub_cert_file,omitempty" db:"sub_cert_file"`
	SubKeyFile          string `json:"sub_key_file,omitempty" db:"sub_key_file"`
	SubDomain           string `json:"sub_domain,omitempty" db:"sub_domain"`
	SubEnable           bool   `json:"sub_enable" db:"sub_enable"`
	SubJSONRules        string `json:"sub_json_rules,omitempty" db:"sub_json_rules"`
	SubClashRules       string `json:"sub_clash_rules,omitempty" db:"sub_clash_rules"`
	SubMux              string `json:"sub_mux,omitempty" db:"sub_mux"`
	SubFinalMask        string `json:"sub_final_mask,omitempty" db:"sub_final_mask"`
	SubEnableRouting    bool   `json:"sub_enable_routing" db:"sub_enable_routing"`
	TOTPEnabled         bool   `json:"totp_enabled" db:"totp_enabled"`
	TOTPToken           string `json:"-" db:"totp_token"`
	TelegramToken       string `json:"-" db:"telegram_token"`
	TelegramChatID      string `json:"-" db:"telegram_chat_id"`
	TelegramRuntime     string `json:"telegram_runtime,omitempty" db:"telegram_runtime"`
	TelegramEnabled     bool   `json:"telegram_enabled" db:"telegram_enabled"`
	WebhookURL          string `json:"-" db:"webhook_url"`
	WebhookSecret       string `json:"-" db:"webhook_secret"`
	WebhookEnabled      bool   `json:"webhook_enabled" db:"webhook_enabled"`
	DefaultCore         string `json:"default_core" db:"default_core"` // xray or singbox
	EnableTunnelMonitor bool   `json:"enable_tunnel_monitor" db:"enable_tunnel_monitor"`
	TunnelMonitorURL    string `json:"tunnel_monitor_url,omitempty" db:"tunnel_monitor_url"`
	TunnelMonitorProxy  string `json:"tunnel_monitor_proxy,omitempty" db:"tunnel_monitor_proxy"`
	AutoRestartCore     bool   `json:"auto_restart_core" db:"auto_restart_core"`
	AutoBackup          bool   `json:"auto_backup" db:"auto_backup"`
	AutoBackupInterval  string `json:"auto_backup_interval,omitempty" db:"auto_backup_interval"`
	LogLevel            string `json:"log_level" db:"log_level"`
	Language            string `json:"language" db:"language"`
	BrandName           string `json:"brand_name,omitempty" db:"brand_name"`
	BrandWebsite        string `json:"brand_website,omitempty" db:"brand_website"`
	BrandLogo           string `json:"brand_logo,omitempty" db:"brand_logo"`
}

// SubscriptionSettings controls subscription delivery behavior.
type SubscriptionSettings struct {
	ID              int64  `json:"id" db:"id"`
	Enable          bool   `json:"enable" db:"enable"`
	Host            string `json:"host,omitempty" db:"host"`
	Path            string `json:"path,omitempty" db:"path"`
	Port            int    `json:"port,omitempty" db:"port"`
	CertFile        string `json:"cert_file,omitempty" db:"cert_file"`
	KeyFile         string `json:"key_file,omitempty" db:"key_file"`
	JSONRules       string `json:"json_rules,omitempty" db:"json_rules"`
	ClashRules      string `json:"clash_rules,omitempty" db:"clash_rules"`
	Mux             string `json:"mux,omitempty" db:"mux"`
	FinalMask       string `json:"final_mask,omitempty" db:"final_mask"`
	EnableRouting   bool   `json:"enable_routing" db:"enable_routing"`
	DefaultEndpoint string `json:"default_endpoint,omitempty" db:"default_endpoint"`
}
