package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vortexuipro/internal/agent"
	"vortexuipro/internal/api"
	"vortexuipro/internal/api/handlers"
	"vortexuipro/internal/cluster"
	"vortexuipro/internal/config"
	"vortexuipro/internal/core"
	"vortexuipro/internal/core/singbox"
	"vortexuipro/internal/core/xray"
	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
	"vortexuipro/internal/metrics"
	"vortexuipro/internal/monitor"
	"vortexuipro/internal/service"
)

// clusterTLSEnabled checks if cluster TLS should be enabled.
func clusterTLSEnabled(cfg *config.Config) bool {
	return cfg.ClusterEnabled && (cfg.ClusterTLSCertPath != "" || cfg.ClusterTLSCAPath != "")
}

var (
	version = "0.0.1"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("VortexUiPro v%s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	fmt.Printf(`
╔══════════════════════════════════════════════╗
║         VortexUiPro v%-30s ║
║     The Ultimate Proxy Management Panel      ║
╚══════════════════════════════════════════════╝
`, version)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Config loaded: %s", cfg)

	// Initialize database (GORM + SQLite)
	if err := database.InitDB(database.Config{
		Type:     cfg.DBType,
		DSN:      cfg.DatabaseURL,
		LogLevel: cfg.LogLevel,
	}); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Seed default admin if none exists
	if err := database.SeedDefaults(); err != nil {
		log.Printf("Warning: seed defaults: %v", err)
	}
	log.Println("Database initialized")

	// Initialize event bus
	eventBus := events.New(func(msg string, args ...any) {
		log.Printf(msg, args...)
	})
	log.Println("Event bus initialized")

	// Initialize core engine manager
	engineMgr := core.NewEngineManager()
	log.Println("Core engine manager initialized")

	// Initialize services
	xrayCfg := xray.Config{
		BinaryPath: cfg.CoreBin,
		ConfigPath: cfg.CoreConfig,
		APIPort:    cfg.CoreAPIPort,
	}
	xrayDriver := xray.New(xrayCfg)
	engineMgr.Register(xrayDriver)
	xrayService := service.NewXrayService(xrayCfg, eventBus)
	log.Println("Xray service initialized (registered in EngineManager)")

	inboundSvc := service.NewInboundService(eventBus, xrayService)
	userSvc := service.NewUserService(eventBus)
	adminSvc := service.NewAdminService(cfg.JWTSecret, eventBus)

	outboundSvc := service.NewOutboundService(eventBus)
	log.Println("Outbound service initialized")

	routingSvc := service.NewRoutingService(eventBus)
	log.Println("Routing service initialized")

	subSvc := service.NewSubscriptionService(inboundSvc, userSvc)
	log.Println("Subscription service initialized")

	// Initialize analytics service
	analyticsSvc := service.NewAnalyticsService()
	log.Println("Analytics service initialized")

	// Initialize anti-censorship service
	antiCensorSvc := service.NewAntiCensorshipService("/etc/vortex/data")
	antiCensorHandler := handlers.NewAntiCensorHandler(antiCensorSvc)
	log.Println("Anti-censorship service initialized")

	// Initialize payment handler
	panelURL := fmt.Sprintf("http://localhost%s", cfg.HTTPAddr)
	paymentHandler := handlers.NewPaymentHandler(cfg.ZarinPalMerchant, cfg.NowPaymentsAPIKey, panelURL)
	log.Println("Payment handler initialized")

	// Register sing-box engine (multi-core support)
	singboxCfg := singbox.Config{
		BinaryPath: cfg.SingboxBin,
		ConfigPath: "/etc/vortex/singbox.json",
	}
	singboxDriver := singbox.New(singboxCfg)
	engineMgr.Register(singboxDriver)
	log.Println("Sing-box engine registered")

	// Initialize Client Activity & Tracking Services
	activitySvc := service.NewClientActivityService(cfg.ActivityFlushSec)
	activitySvc.StartAutoFlush()
	trafficCollector := service.NewTrafficCollector(xrayService, activitySvc)
	trafficCollector.Start()
	log.Println("Client activity tracking started")

	// Initialize Telegram Bot & Notification Service
	telegramBot := service.NewTelegramBot(cfg.TelegramBotToken)
	telegramBot.Start()
	notifySvc := service.NewNotificationService(telegramBot)
	log.Println("Notification service initialized")

	// Wire notification service to event bus for system events
	eventCh := eventBus.Subscribe(64)
	go func() {
		for e := range eventCh {
			notifySvc.NotifySystem(string(e.Type), e.Message)
		}
	}()

	// Start Node Agent gRPC server
	nodeAgent := agent.NewNodeAgentServer(cfg.GRPCAddr)
	if err := nodeAgent.Start(); err != nil {
		log.Printf("Warning: node agent server: %v", err)
	} else {
		log.Printf("Node Agent gRPC server started on %s", cfg.GRPCAddr)
	}

	// Seed default RBAC roles if none exist
	if err := service.SeedDefaultRoles(database.DB); err != nil {
		log.Printf("Warning: seed default roles: %v", err)
	} else {
		log.Println("Default RBAC roles seeded")
	}

	// Initialize RBAC services
	adminRoleSvc := service.NewAdminRoleService()
	apiTokenSvc := service.NewApiTokenService()
	log.Println("RBAC services initialized (AdminRoles + ApiTokens)")

	// Initialize Online Tracker (real-time user monitoring)
	onlineTracker := service.NewOnlineTracker()
	log.Println("Online tracker initialized")

	// Initialize Backup Crypto + Remote Storage services
	cryptoSvc := service.NewBackupCryptoService()
	storageSvc := service.NewRemoteStorageService()

	// Initialize Backup Service (enhanced with encryption + remote storage + Telegram)
	backupSvc := service.NewBackupService("/etc/vortex/backups", cryptoSvc, storageSvc, telegramBot)
	backupSvc.StartAutoBackup(24 * time.Hour)
	backupSvc.CleanupOldBackups(30 * 24 * time.Hour) // cleanup backups older than 30 days
	log.Println("Backup service initialized (auto every 24h)")

	// Initialize Prometheus Metrics Collector
	metricsCollector := metrics.NewCollector(onlineTracker)
	log.Println("Metrics collector initialized")

	// Initialize monitoring handler
	monitorHandler := handlers.NewMonitorHandler(onlineTracker, activitySvc, xrayService)
	telegramHandler := handlers.NewTelegramSettingsHandler(telegramBot)
	backupHandler := handlers.NewBackupHandler(backupSvc)
	metricsHandler := handlers.NewMetricsHandler(metricsCollector)
	routingHandler := handlers.NewRoutingHandler(routingSvc)
	log.Println("Monitor + Telegram + Backup + Metrics + Routing handlers initialized")

	// Initialize SubProfile Service + Handler (Phase 5.2)
	subProfileSvc := service.NewSubProfileService(eventBus)
	subProfileHandler := handlers.NewSubProfileHandler(subProfileSvc)
	log.Println("SubProfile service + handler initialized (Phase 5.2)")

	// Initialize Advanced Security Service + Handler (Phase 7.1)
	advancedSecuritySvc := service.NewAdvancedSecurityService(eventBus)
	advancedSecuritySvc.Start()
	advancedSecurityHandler := handlers.NewAdvancedSecurityHandler(advancedSecuritySvc)
	log.Println("Advanced Security service + handler initialized (Phase 7.1)")

	// Initialize Email / SMTP Service + Handler (Phase 7.4)
	emailSvc := service.NewEmailService()
	emailHandler := handlers.NewEmailHandler(emailSvc)
	log.Println("Email/SMTP service + handler initialized (Phase 7.4)")

	// Initialize Clean IP Scanner (Phase 7.5)
	cleanIPSvc := service.NewCleanIPScanner()
	cleanIPSvc.Start()
	log.Println("Clean IP scanner initialized (Phase 7.5)")

	// Initialize Config Version Service (Phase 7.6)
	configVersionSvc := service.NewConfigVersionService(eventBus)
	log.Println("Config Version service initialized (Phase 7.6)")

	// Create CleanIP + ConfigVersion handlers for router
	cleanIPHandler := handlers.NewCleanIPHandler(cleanIPSvc)
	cvHandler := handlers.NewConfigVersionHandler(configVersionSvc)
	log.Println("CleanIP + ConfigVersion handlers created")

	// Initialize Federation Service + Handler (Phase 7.7)
	federationSvc := service.NewFederationService(eventBus)
	federationSvc.Start()
	federationHandler := handlers.NewFederationHandler(federationSvc)
	log.Println("Federation service + handler initialized (Phase 7.7)")

	// Initialize Security Settings Service + Handler (Phase 6.2)
	securitySvc := service.NewSecuritySettingsService()
	securityHandler := handlers.NewSecuritySettingsHandler(securitySvc, advancedSecuritySvc)
	log.Println("Security Settings service + handler initialized (Phase 6.2)")

	// Initialize Client Group Service + Handler (Phase 7.3)
	clientGroupSvc := service.NewClientGroupService(eventBus)
	clientGroupHandler := handlers.NewClientGroupHandler(clientGroupSvc)
	log.Println("Client Group service + handler initialized (Phase 7.3)")

	// ═══ Phase 8: Initialize all new services ════════════════

	// 8.1: Web Terminal (SSH Console)
	terminalSvc := service.NewTerminalService()
	terminalHandler := handlers.NewTerminalHandler(terminalSvc)
	log.Println("[Phase 8] Web Terminal service initialized")

	// 8.2: Live Log Streaming
	logStreamSvc := service.NewLogStreamService(
		"/var/log/vortexui/core.log",  // core log path
		"/var/log/vortexui/panel.log", // panel log path
	)
	logStreamSvc.Start()
	logStreamHandler := handlers.NewLogStreamHandler(logStreamSvc)
	log.Println("[Phase 8] Live Log Streaming service initialized")

	// 8.3: WARP+ Outbound
	warpSvc := service.NewWARPProxyService("/etc/vortex/data")
	warpHandler := handlers.NewWARPHandler(warpSvc)
	log.Println("[Phase 8] WARP+ Outbound service initialized")

	// 8.4: TLS Tricks Suite
	tlsTricksSvc := service.NewTLSTricksService()
	tlsTricksHandler := handlers.NewTLSTricksHandler(tlsTricksSvc)
	log.Println("[Phase 8] TLS Tricks Suite service initialized")

	// 8.5: Plugin System
	pluginSvc := service.NewPluginService("/etc/vortex/plugins")
	pluginHandler := handlers.NewPluginHandler(pluginSvc)
	log.Println("[Phase 8] Plugin System service initialized")

	// ═══ Phase 10: WebRTC Service ═══════════════════════════
	webrtcSvc := service.NewWebRTCService(eventBus)
	webrtcSvc.Start()
	webrtcHandler := handlers.NewWebRTCHandler(webrtcSvc)
	log.Println("[Phase 10] WebRTC + P2P Mesh service initialized")

	// ═══ Phase 12: Smart Health Check System ═══════════════
	healthCheckSvc := service.NewHealthCheckService(eventBus)
	if err := healthCheckSvc.Start(); err != nil {
		log.Printf("Warning: health check service start: %v", err)
	} else {
		log.Println("[Phase 12] Smart Health Check + Auto-Recovery service initialized")
	}
	healthHandler := handlers.NewHealthHandler(healthCheckSvc)
	log.Println("[Phase 12] Health Check handler created")

	// ═══ Phase 15: Domain Fronting + CDN Proxy ═════════════
	domainFrontingSvc := service.NewDomainFrontingService()
	domainFrontingHandler := handlers.NewDomainFrontingHandler(domainFrontingSvc)
	log.Println("[Phase 15] Domain Fronting + CDN Proxy service initialized")

	// ═══ Phase 15: Smart DNS System ═════════════════════════
	smartDNSSvc := service.NewSmartDNSService()
	smartDNSHandler := handlers.NewSmartDNSHandler(smartDNSSvc)
	log.Println("[Phase 15] Smart DNS system initialized")

	// ═══ Phase 15: Docker Native Mode ═══════════════════════
	dockerSvc := service.NewDockerService()
	dockerHandler := handlers.NewDockerHandler(dockerSvc)
	log.Println("[Phase 15] Docker Native Mode service initialized")

	// ═══ Phase 16: Xray gRPC Real Integration ═══════════════
	xrayAPIHandler := handlers.NewXrayAPIHandler(xrayService)
	log.Println("[Phase 16] Xray gRPC API handler initialized")

	// Initialize Cluster Manager (Multi-Node Mesh) with mTLS + gRPC + Topology streaming
	clusterCfg := cluster.Config{
		Enabled:           cfg.ClusterEnabled,
		NodeName:          cfg.ClusterNodeName,
		Addr:              cfg.ClusterAddr,
		Peers:             cfg.ClusterPeers,
		Region:            cfg.ClusterRegion,
		Priority:          cfg.ClusterPriority,
		HeartbeatInterval: 5 * time.Second,
		HeartbeatTimeout:  15 * time.Second,
		PKIDir:            "/etc/vortex/pki",
		TLSEnabled:        clusterTLSEnabled(cfg),
		GRPCEnabled:       cfg.ClusterGRPCEnabled,
		WebSocketHub:      nil, // Set after router is created, see below
	}
	clusterMgr := cluster.NewManager(clusterCfg)
	if cfg.ClusterEnabled {
		if err := clusterMgr.Start(); err != nil {
			log.Printf("Warning: cluster manager start: %v", err)
		} else {
			log.Println("Cluster manager started (multi-node mesh)")
		}
	} else {
		log.Println("Cluster mode disabled")
	}

	// Setup API router
	router := api.NewRouter(
		adminSvc, userSvc, inboundSvc, outboundSvc, xrayService, subSvc,
		engineMgr, analyticsSvc, paymentHandler, antiCensorHandler,
		adminRoleSvc, apiTokenSvc,
		monitorHandler, telegramHandler, onlineTracker,
		backupHandler, metricsHandler,
		clusterMgr, routingHandler,
		subProfileHandler, advancedSecurityHandler, emailHandler,
		cleanIPHandler, cvHandler,
		securityHandler, clientGroupHandler,
		federationHandler,
		terminalHandler, logStreamHandler, warpHandler, tlsTricksHandler, pluginHandler,
		webrtcHandler,
		healthHandler,
		domainFrontingHandler,
		smartDNSHandler,
		dockerHandler,
		xrayAPIHandler,
	)
	log.Printf("API router initialized on %s", cfg.HTTPAddr)

	// Store hub reference for graceful shutdown
	wsHub := router.Hub

	// HTTP server
	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router.Engine(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Tunnel health monitor (from Heimdall-main)
	tunnelMonitor, err := monitor.New(monitor.ConfigFromEnv(), func(ctx context.Context) error {
		log.Println("Tunnel health monitor: triggering core restart")
		return xrayService.Restart(ctx)
	})
	if err != nil {
		log.Printf("Warning: tunnel monitor disabled: %v", err)
	}

	// Start server
	go func() {
		log.Printf("VortexUiPro listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start tunnel monitor
	tunnelCtx, tunnelCancel := context.WithCancel(context.Background())
	if tunnelMonitor != nil {
		go tunnelMonitor.Run(tunnelCtx)
		log.Println("Tunnel health monitor started")
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	tunnelCancel()

	// Stop activity tracking
	activitySvc.StopAutoFlush()
	activitySvc.Flush() // flush remaining records
	log.Println("Activity tracking stopped")

	// Stop traffic collector
	trafficCollector.Stop()
	log.Println("Traffic collector stopped")

	// Unsubscribe event bus channel
	eventBus.Unsubscribe(eventCh)
	log.Println("Event bus unsubscribed")

	// Stop auto-backup
	backupSvc.StopAutoBackup()
	log.Println("Auto-backup stopped")

	// Stop online tracker
	onlineTracker.Stop()
	log.Println("Online tracker stopped")

	// Stop Telegram bot
	telegramBot.Stop()
	log.Println("Telegram bot stopped")

	// Stop metrics collector
	metricsCollector.Stop()
	log.Println("Metrics collector stopped")

	// Stop WebSocket hub
	// Stop Health Check Service
	healthCheckSvc.Stop()
	log.Println("Health check service stopped")

	wsHub.Stop()
	log.Println("WebSocket hub stopped")

	// Stop Cluster Manager
	clusterMgr.Stop()
	log.Println("Cluster manager stopped")

	// Stop Federation Service
	federationSvc.Stop()
	log.Println("Federation service stopped")

	// Stop Clean IP Scanner
	cleanIPSvc.Stop()
	log.Println("Clean IP scanner stopped")

	// Stop Node Agent
	nodeAgent.Stop()
	log.Println("Node Agent stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("VortexUiPro stopped gracefully")
}
