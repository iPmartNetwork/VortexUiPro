package api

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/api/handlers"
	"vortexuipro/internal/api/hub"
	"vortexuipro/internal/api/middleware"
	"vortexuipro/internal/cluster"
	"vortexuipro/internal/core"
	"vortexuipro/internal/service"
	"vortexuipro/web"
)

type Router struct {
	engine *gin.Engine
	Hub    *hub.Hub
}

func NewRouter(
	adminSvc *service.AdminService,
	userSvc *service.UserService,
	inboundSvc *service.InboundService,
	outboundSvc *service.OutboundService,
	xraySvc *service.XrayService,
	subSvc *service.SubscriptionService,
	engineMgr *core.EngineManager,
	analyticsSvc *service.AnalyticsService,
	paymentHandler *handlers.PaymentHandler,
	antiCensorHandler *handlers.AntiCensorHandler,
	adminRoleSvc *service.AdminRoleService,
	apiTokenSvc *service.ApiTokenService,
	monitorHandler *handlers.MonitorHandler,
	telegramHandler *handlers.TelegramSettingsHandler,
	onlineTracker *service.OnlineTracker,
	backupHandler *handlers.BackupHandler,
	metricsHandler *handlers.MetricsHandler,
	clusterMgr *cluster.Manager,
	routingHandler *handlers.RoutingHandler,
	subProfileHandler *handlers.SubProfileHandler,
	advancedSecurityHandler *handlers.AdvancedSecurityHandler,
	emailHandler *handlers.EmailHandler,
	cleanIPHandler *handlers.CleanIPHandler,
	cvHandler *handlers.ConfigVersionHandler,
	securityHandler *handlers.SecuritySettingsHandler,
	clientGroupHandler *handlers.ClientGroupHandler,
	federationHandler *handlers.FederationHandler,
	terminalHandler *handlers.TerminalHandler,
	logStreamHandler *handlers.LogStreamHandler,
	warpHandler *handlers.WARPHandler,
	tlsTricksHandler *handlers.TLSTricksHandler,
	pluginHandler *handlers.PluginHandler,
	webrtcHandler *handlers.WebRTCHandler,
	healthHandler *handlers.HealthHandler,
	domainFrontingHandler *handlers.DomainFrontingHandler,
	smartDNSHandler *handlers.SmartDNSHandler,
	dockerHandler *handlers.DockerHandler,
	xrayAPIHandler *handlers.XrayAPIHandler,
) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(gin.Recovery())
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.RequestIDMiddleware())
	engine.Use(middleware.SecurityHeadersMiddleware())
	rl := middleware.NewRateLimiter(
		middleware.RateLimitConfig{Name: "api", Limit: 300, Window: time.Minute},
		middleware.RateLimitConfig{Name: "auth", Limit: 20, Window: 5 * time.Minute},
		middleware.RateLimitConfig{Name: "subscription", Limit: 600, Window: time.Minute},
		middleware.RateLimitConfig{Name: "default", Limit: 100, Window: time.Minute},
	)
	engine.Use(rl.Middleware())

	wsHub := hub.NewHub()
	go wsHub.Run()
	wsHub.BroadcastNotification("info", "System", "VortexUiPro started")

	// Start broadcasting online count via WebSocket
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if onlineTracker != nil {
				count := onlineTracker.GetOnlineCount()
				wsHub.BroadcastTraffic(0, 0, count)
			}
		}
	}()

	// Start broadcasting real-time metrics via WebSocket (every 15s)
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if metricsHandler != nil {
				wsHub.BroadcastMetrics(metricsHandler.CollectorSnapshot())
			}
		}
	}()

	authHandler := handlers.NewAuthHandler(adminSvc)
	userHandler := handlers.NewUserHandler(userSvc)
	inboundHandler := handlers.NewInboundHandler(inboundSvc, xraySvc)
	systemHandler := handlers.NewSystemHandler(engineMgr)
	subHandler := handlers.NewSubscriptionHandler(subSvc, userSvc)
	ticketHandler := handlers.NewTicketHandler()
	nodeHandler := handlers.NewNodeHandler()
	settingHandler := handlers.NewSettingHandler()
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsSvc)
	adminRoleHandler := handlers.NewAdminRoleHandler(adminRoleSvc)
	adminMgmtHandler := handlers.NewAdminManagementHandler(adminSvc)
	apiTokenHandler := handlers.NewApiTokenHandler(apiTokenSvc)

	public := engine.Group("/api/v1")
	{
		public.POST("/login", authHandler.Login)
		public.POST("/auth/refresh", authHandler.Refresh)
		public.GET("/health", systemHandler.Health)
	}

	// Public Prometheus metrics endpoint (no auth required)
	engine.GET("/metrics", metricsHandler.PrometheusMetrics)

	engine.GET("/ws", wsHub.HandleWebSocket)

	sub := engine.Group("/sub")
	{
		sub.GET("/:clientId", subHandler.GetConfig)
		sub.GET("/:clientId/info", subHandler.GetInfo)
		sub.GET("/:clientId/link", subHandler.GetLink)
		sub.GET("/:clientId/share-links", subHandler.GetShareLinks)
	}

	// Subscription endpoints for grouped/subscription-based access
	subGroup := engine.Group("/sub-group")
	{
		subGroup.GET("/:subId", subHandler.GetSubLinks)
	}

	payments := engine.Group("/api/v1/payments")
	{
		payments.GET("/zarinpal/verify", paymentHandler.ZarinpalVerify)
		payments.POST("/nowpayments/ipn", paymentHandler.NOWPaymentsIPN)
	}

	protected := engine.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(adminSvc))
	{
		protected.GET("/me", authHandler.Me)
		protected.POST("/auth/totp/setup", authHandler.SetupTOTP)
		protected.POST("/auth/totp/validate", authHandler.ValidateTOTP)
		protected.POST("/auth/change-password", authHandler.ChangePassword)

		protected.GET("/analytics/stats", analyticsHandler.Stats)
		protected.GET("/analytics/traffic", analyticsHandler.Traffic)
		protected.GET("/analytics/user-growth", analyticsHandler.UserGrowth)
		protected.GET("/analytics/revenue", analyticsHandler.Revenue)
		protected.GET("/analytics/online", analyticsHandler.Online)

		// ─── Routing Rules & Packs ──────────────────────────────
		protected.GET("/routing/rules", routingHandler.ListRules)
		protected.GET("/routing/rules/:id", routingHandler.GetRule)
		protected.POST("/routing/rules", routingHandler.CreateRule)
		protected.PUT("/routing/rules/:id", routingHandler.UpdateRule)
		protected.DELETE("/routing/rules/:id", routingHandler.DeleteRule)
		protected.PUT("/routing/rules/:id/toggle", routingHandler.ToggleRule)
		protected.GET("/routing/packs", routingHandler.ListPacks)
		protected.POST("/routing/packs", routingHandler.CreatePack)
		protected.DELETE("/routing/packs/:id", routingHandler.DeletePack)
		protected.GET("/routing/generate", routingHandler.GenerateConfig)

		// ─── Outbound Management ────────────────────────────────
		outboundHandler := handlers.NewOutboundHandler(outboundSvc)
		protected.GET("/outbounds", outboundHandler.List)
		protected.GET("/outbounds/:id", outboundHandler.Get)
		protected.POST("/outbounds", outboundHandler.Create)
		protected.PUT("/outbounds/:id", outboundHandler.Update)
		protected.DELETE("/outbounds/:id", outboundHandler.Delete)
		protected.PUT("/outbounds/:id/visibility", outboundHandler.ToggleHide)

		protected.GET("/plans", paymentHandler.ListPlans)
		protected.POST("/plans", paymentHandler.CreatePlan)
		protected.DELETE("/plans/:id", paymentHandler.DeletePlan)
		protected.POST("/orders", paymentHandler.CreateOrder)
		protected.GET("/orders", paymentHandler.ListOrders)
		protected.POST("/payments/zarinpal/request", paymentHandler.ZarinpalRequest)
		protected.POST("/payments/nowpayments/create", paymentHandler.NOWPaymentsCreate)

		protected.GET("/anticensor/tricks", antiCensorHandler.ListTricks)
		protected.GET("/anticensor/fingerprints", antiCensorHandler.ListFingerprints)
		protected.GET("/anticensor/scan", antiCensorHandler.ScanTarget)
		protected.GET("/anticensor/decoy", antiCensorHandler.GenerateDecoyConfig)
		protected.GET("/anticensor/cert", antiCensorHandler.GenerateSelfSignedCert)
		protected.POST("/anticensor/cert/save", antiCensorHandler.SaveCert)
		protected.GET("/anticensor/fragment", antiCensorHandler.GetFragmentConfig)
		protected.GET("/anticensor/padding", antiCensorHandler.GetPaddingConfig)
		protected.GET("/anticensor/mix", antiCensorHandler.GenerateMixConfig)
		protected.GET("/anticensor/anti-dpi", antiCensorHandler.GenerateAntiDPIConfig)
		protected.GET("/anticensor/mtproto", antiCensorHandler.GenerateMTProtoConfig)
		protected.GET("/anticensor/warp", antiCensorHandler.GenerateWarpConfig)

		protected.GET("/portal/clients", userHandler.ListOwnClients)
		protected.GET("/portal/clients/:id", userHandler.GetClientDetail)
		protected.GET("/portal/traffic", userHandler.GetOwnTraffic)
		protected.GET("/portal/tickets", ticketHandler.List)
		protected.POST("/portal/tickets", ticketHandler.Create)

		// ─── Online Users (replaces old analytics/online) ─────
		protected.GET("/monitor/online", monitorHandler.GetOnlineUsers)
		protected.GET("/monitor/online/count", monitorHandler.GetOnlineCount)
		protected.GET("/monitor/activity", monitorHandler.GetRecentActivity)

		// ─── Traffic Reset + Sync ─────────────────────────────
		protected.POST("/traffic/reset/:id", monitorHandler.ResetUserTraffic)
		protected.GET("/traffic/sync", monitorHandler.SyncUserTraffic)

		// ─── Reseller Management ─────────────────────────────
		reseller := protected.Group("/resellers")
		reseller.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			reseller.GET("/stats", monitorHandler.ResellerStats)
		}

		// ─── Telegram Bot for Clients ────────────────────────
		telegram := protected.Group("/telegram")
		{
			telegram.POST("/client/link", telegramHandler.SetClientTelegram)
			telegram.POST("/test", telegramHandler.SendTestNotification)
			telegram.POST("/notify", telegramHandler.NotifyClientUsage)
		}

		// ─── Backup & Restore ────────────────────────────────
		backups := protected.Group("/backups")
		backups.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			backups.GET("", backupHandler.List)
			backups.POST("", backupHandler.Create)
			backups.GET("/:id/download", backupHandler.Download)
			backups.DELETE("/:id", backupHandler.Delete)
			backups.POST("/:id/restore", backupHandler.Restore)
			backups.POST("/auto-config", backupHandler.AutoBackupConfig)
			// Phase 14: Advanced Backup
			backups.POST("/:id/sync", backupHandler.SyncBackupToRemote)
			backups.POST("/:id/telegram", backupHandler.SendBackupToTelegram)
		}

		// Phase 14: Encryption Keys
		encryption := protected.Group("/backups/encryption")
		encryption.Use(middleware.RoleMiddleware("super_admin"))
		{
			encryption.GET("/keys", backupHandler.ListEncryptionKeys)
			encryption.POST("/keys", backupHandler.GenerateEncryptionKey)
			encryption.DELETE("/keys/:id", backupHandler.DeleteEncryptionKey)
		}

		// Phase 14: Remote Storage Configs
		remoteStorage := protected.Group("/backups/remote-storage")
		remoteStorage.Use(middleware.RoleMiddleware("super_admin"))
		{
			remoteStorage.GET("", backupHandler.ListRemoteStorageConfigs)
			remoteStorage.POST("", backupHandler.SaveRemoteStorageConfig)
			remoteStorage.DELETE("/:id", backupHandler.DeleteRemoteStorageConfig)
		}

		// ─── Metrics & Prometheus ────────────────────────────
		protected.GET("/metrics", metricsHandler.GetMetrics)
		protected.GET("/metrics/history", metricsHandler.GetHistory)

		// ─── Subscription Profiles (Multi-Profile, Heimdall feature) ──
		protected.GET("/sub-profiles", subProfileHandler.ListProfiles)
		protected.POST("/sub-profiles", subProfileHandler.CreateProfile)
		protected.DELETE("/sub-profiles/:id", subProfileHandler.DeleteProfile)
		protected.GET("/sub-hosts", subProfileHandler.ListHosts)
		protected.POST("/sub-hosts", subProfileHandler.CreateHost)
		protected.DELETE("/sub-hosts/:id", subProfileHandler.DeleteHost)
		protected.GET("/sub-formats", subProfileHandler.ListFormats)
		protected.GET("/sub-vars", subProfileHandler.ListRemarkVars)

		// ─── Advanced Security ────────────────────────────────
		sec := protected.Group("/security")
		sec.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			sec.GET("/audit-logs", advancedSecurityHandler.ListAuditLogs)
			sec.GET("/threat-summary", advancedSecurityHandler.GetThreatSummary)
			sec.POST("/compliance-check", advancedSecurityHandler.RunComplianceCheck)
		}

		// ─── Email / SMTP (Restricted: super_admin only) ──────
		email := protected.Group("/email")
		email.Use(middleware.RoleMiddleware("super_admin"))
		{
			email.GET("/config", emailHandler.GetConfig)
			email.PUT("/config", emailHandler.SaveConfig)
			email.POST("/test", emailHandler.SendTest)
		}

		// ─── Clean IP Scanner ─────────────────────────────────
		cleanIP := protected.Group("/clean-ip")
		cleanIP.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			cleanIP.GET("/results", cleanIPHandler.GetResults)
			cleanIP.POST("/scan", cleanIPHandler.ScanNow)
		}

		// ─── Config Versions ──────────────────────────────────
		cv := protected.Group("/config-versions")
		cv.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			cv.GET("", cvHandler.ListVersions)
			cv.POST("/:id/rollback", cvHandler.Rollback)
		}

		// ─── Security Settings (Geo-Block, Password Policy, IP) ──
		secSettings := protected.Group("/settings/security")
		secSettings.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			secSettings.GET("/geo-block", securityHandler.GetGeoBlock)
			secSettings.PUT("/geo-block", securityHandler.SetGeoBlock)
			secSettings.GET("/password-policy", securityHandler.GetPasswordPolicy)
			secSettings.PUT("/password-policy", securityHandler.SavePasswordPolicy)
			secSettings.GET("/banned-ips", securityHandler.GetBannedIPs)
			secSettings.POST("/banned-ips", securityHandler.AddBannedIP)
			secSettings.DELETE("/banned-ips", securityHandler.RemoveBannedIP)
			secSettings.GET("/whitelisted-ips", securityHandler.GetWhitelistedIPs)
			secSettings.POST("/whitelisted-ips", securityHandler.AddWhitelistedIP)
			secSettings.DELETE("/whitelisted-ips", securityHandler.RemoveWhitelistedIP)
			secSettings.GET("/threat-config", securityHandler.GetThreatConfig)
		}

		// ─── Client Groups + Bulk Operations ────────────────
		groups := protected.Group("/client-groups")
		{
			groups.GET("", clientGroupHandler.ListGroups)
			groups.POST("", clientGroupHandler.CreateGroup)
			groups.GET("/with-clients", clientGroupHandler.GetClientsWithGroups)
			groups.GET("/:id", clientGroupHandler.GetGroup)
			groups.PUT("/:id", clientGroupHandler.UpdateGroup)
			groups.DELETE("/:id", clientGroupHandler.DeleteGroup)
			groups.POST("/:id/clients", clientGroupHandler.AddClientToGroup)
			groups.DELETE("/:id/clients", clientGroupHandler.RemoveClientFromGroup)
			groups.GET("/:id/clients", clientGroupHandler.GetGroupClients)
			groups.POST("/:id/clients/bulk", clientGroupHandler.BulkAddClients)
			groups.DELETE("/:id/clients/bulk", clientGroupHandler.BulkRemoveClients)
		}

		// ─── Federation (Cross-Panel Sync) ───────────────────
		fed := protected.Group("/federation")
		fed.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			fed.GET("/providers", federationHandler.ListProviders)
			fed.POST("/providers", federationHandler.CreateProvider)
			fed.PUT("/providers/:id", federationHandler.UpdateProvider)
			fed.DELETE("/providers/:id", federationHandler.DeleteProvider)
			fed.POST("/providers/:id/test", federationHandler.TestConnection)
			fed.POST("/sync", federationHandler.TriggerSync)
			fed.POST("/sync/:id", federationHandler.TriggerSync)
		}

		// ─── Incoming Federation Endpoints (called by remote panels) ──
		fedIn := engine.Group("/api/v1/federation")
		fedKeyValidator := func(key string) bool {
			providers, err := federationHandler.ListProvidersRaw()
			if err != nil {
				return false
			}
			for _, p := range providers {
				if p.APIKey == key {
					return true
				}
			}
			return false
		}
		fedIn.Use(middleware.FederationKeyMiddleware(fedKeyValidator))
		{
			fedIn.GET("/users", federationHandler.HandleFederationUsers)
			fedIn.POST("/users", federationHandler.HandleFederationUsers)
			fedIn.GET("/plans", federationHandler.HandleFederationPlans)
			fedIn.POST("/plans", federationHandler.HandleFederationPlans)
			fedIn.POST("/traffic", federationHandler.HandleFederationTraffic)
		}

		// ─── Cluster Management (Multi-Node Mesh) ────────────
		clusterHandler := handlers.NewClusterHandler(clusterMgr)
		clstr := protected.Group("/cluster")
		clstr.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			clstr.GET("/status", clusterHandler.Status)
			clstr.GET("/nodes", clusterHandler.ListNodes)
			clstr.GET("/nodes/:id", clusterHandler.GetNode)
			clstr.POST("/nodes", clusterHandler.AddNode)
			clstr.PUT("/nodes/:id", clusterHandler.UpdateNode)
			clstr.DELETE("/nodes/:id", clusterHandler.DeleteNode)
			clstr.GET("/election", clusterHandler.ElectionStats)
			clstr.GET("/sync-events", clusterHandler.SyncEvents)
			clstr.POST("/election/force", clusterHandler.ForceElection)
			clstr.GET("/topology", clusterHandler.Topology)
			clstr.GET("/pki", clusterHandler.PKIStatus)
		}

		// ─── RBAC: Admin Roles ──────────────────────────────
		roles := protected.Group("/roles")
		roles.Use(middleware.RoleOrPermissionMiddleware([]string{"super_admin", "admin"}, "roles", "view"))
		{
			roles.GET("", adminRoleHandler.List)
			roles.GET("/:id", adminRoleHandler.Get)
			roles.POST("", adminRoleHandler.Create)
			roles.PUT("/:id", adminRoleHandler.Update)
			roles.POST("/:id/duplicate", adminRoleHandler.Duplicate)
			roles.DELETE("/:id", adminRoleHandler.Delete)
		}

		adminMgmt := protected.Group("/admins")
		adminMgmt.Use(middleware.RoleOrPermissionMiddleware([]string{"super_admin", "admin"}, "admins", "view"))
		{
			adminMgmt.GET("", adminMgmtHandler.List)
			adminMgmt.GET("/:id", adminMgmtHandler.Get)
			adminMgmt.POST("", adminMgmtHandler.Create)
			adminMgmt.PUT("/:id", adminMgmtHandler.Update)
			adminMgmt.DELETE("/:id", adminMgmtHandler.Delete)
			adminMgmt.PUT("/:id/status", adminMgmtHandler.SetEnabled)
		}

		tokens := protected.Group("/api-tokens")
		tokens.Use(middleware.RoleOrPermissionMiddleware([]string{"super_admin"}, "settings", "view"))
		{
			tokens.GET("", apiTokenHandler.List)
			tokens.GET("/subjects", apiTokenHandler.ListDelegatedSubjects)
			tokens.POST("", apiTokenHandler.Create)
			tokens.DELETE("/:id", apiTokenHandler.Delete)
			tokens.PUT("/:id/status", apiTokenHandler.SetEnabled)
		}

		admin := protected.Group("/admin")
		admin.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			admin.GET("/users", userHandler.List)
			admin.GET("/users/:id", userHandler.Get)
			admin.POST("/users", userHandler.Create)
			admin.DELETE("/users/:id", userHandler.Delete)
			admin.PUT("/users/:id", userHandler.Update)
			admin.GET("/users/:id/clients", userHandler.ListClients)
			admin.POST("/users/:id/clients", userHandler.AddClient)
			admin.DELETE("/clients/:clientId", userHandler.DeleteClient)
			admin.POST("/users/:id/reset-traffic", userHandler.ResetTraffic)
		}

		inbounds := protected.Group("/inbounds")
		inbounds.Use(middleware.RoleMiddleware("super_admin", "admin", "reseller"))
		{
			inbounds.GET("", inboundHandler.List)
			inbounds.GET("/:id", inboundHandler.Get)
			inbounds.POST("", inboundHandler.Create)
			inbounds.PUT("/:id", inboundHandler.Update)
			inbounds.DELETE("/:id", inboundHandler.Delete)
			inbounds.GET("/xray-config", inboundHandler.GetXrayConfig)
		}

		nodes := protected.Group("/nodes")
		nodes.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			nodes.GET("", nodeHandler.List)
			nodes.GET("/:id", nodeHandler.Get)
			nodes.POST("", nodeHandler.Create)
			nodes.PUT("/:id", nodeHandler.Update)
			nodes.DELETE("/:id", nodeHandler.Delete)
		}

		tickets := protected.Group("/tickets")
		{
			tickets.GET("", ticketHandler.List)
			tickets.GET("/stats", ticketHandler.Stats)
			tickets.GET("/:id", ticketHandler.Get)
			tickets.POST("", ticketHandler.Create)
			tickets.POST("/:id/reply", ticketHandler.Reply)
			tickets.POST("/:id/close", ticketHandler.Close)
			tickets.DELETE("/:id", ticketHandler.Delete)
		}

		settings := protected.Group("/settings")
		settings.Use(middleware.RoleMiddleware("super_admin"))
		{
			settings.GET("", settingHandler.List)
			settings.GET("/:key", settingHandler.Get)
			settings.PUT("", settingHandler.Update)
		}

		// ─── Phase 8: Terminal ───────────────────────────────
		term := protected.Group("/terminal")
		term.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			term.GET("/sessions", terminalHandler.ListSessions)
			term.DELETE("/sessions/:id", terminalHandler.CloseSession)
		}

		// ─── Phase 8: Live Logs ───────────────────────────────
		logs := protected.Group("/logs")
		logs.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			logs.GET("/recent", logStreamHandler.GetRecentLogs)
		}

		// ─── Phase 8: WARP+ ───────────────────────────────────
		warp := protected.Group("/warp")
		warp.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			warp.GET("/config", warpHandler.GetConfig)
			warp.PUT("/config", warpHandler.UpdateConfig)
			warp.POST("/connect", warpHandler.Connect)
			warp.POST("/disconnect", warpHandler.Disconnect)
			warp.GET("/status", warpHandler.GetStatus)
			warp.GET("/xray-outbound", warpHandler.GetXrayOutbound)
		}

		// ─── Phase 8: TLS Tricks ──────────────────────────────
		tls := protected.Group("/tls-tricks")
		tls.Use(middleware.RoleMiddleware("super_admin", "admin"))
		{
			tls.GET("/profiles", tlsTricksHandler.ListProfiles)
			tls.GET("/profiles/:id", tlsTricksHandler.GetProfile)
			tls.POST("/profiles", tlsTricksHandler.SaveProfile)
			tls.PUT("/profiles/:id", tlsTricksHandler.EnableProfile)
			tls.DELETE("/profiles/:id", tlsTricksHandler.DeleteProfile)
			tls.GET("/generate", tlsTricksHandler.GenerateConfig)
		}

		// ─── Phase 8: Plugin System ────────────────────────────
		plgn := protected.Group("/plugins")
		plgn.Use(middleware.RoleMiddleware("super_admin"))
		{
			plgn.GET("", pluginHandler.ListPlugins)
			plgn.GET("/:id", pluginHandler.GetPlugin)
			plgn.POST("/load", pluginHandler.LoadPlugin)
			plgn.DELETE("/:id", pluginHandler.UnloadPlugin)
			plgn.PUT("/:id", pluginHandler.EnablePlugin)
		}

		system := protected.Group("/system")
		system.Use(middleware.RoleMiddleware("super_admin"))
		{
			system.GET("/status", systemHandler.Status)
			system.GET("/performance", systemHandler.Performance)
			system.GET("/core-status", systemHandler.CoreStatus)
			system.GET("/config", systemHandler.Config)
			system.GET("/logs", systemHandler.GetLogs)
			system.POST("/reset-traffic", systemHandler.ResetTraffic)
		}
	}

	// ─── Phase 8: WebSocket endpoints (public, no auth) ────────
	engine.GET("/api/v1/terminal/ws", terminalHandler.HandleTerminalWS)
	engine.GET("/api/v1/logs/ws", logStreamHandler.HandleLogStream)

	// ─── Phase 10: WebRTC (Direct Connections + P2P Mesh) ───
	webrtc := protected.Group("/webrtc")
	webrtc.Use(middleware.RoleMiddleware("super_admin", "admin"))
	{
		// ICE / STUN / TURN
		webrtc.GET("/ice-config", webrtcHandler.GetICEConfig)
		webrtc.GET("/turn-servers", webrtcHandler.ListTURNServers)
		webrtc.POST("/turn-servers", webrtcHandler.CreateTURNServer)
		webrtc.DELETE("/turn-servers/:id", webrtcHandler.DeleteTURNServer)
		webrtc.POST("/turn-servers/test", webrtcHandler.TestTURNServer)

		// P2P Mesh
		webrtc.GET("/mesh-config", webrtcHandler.GetMeshConfig)
		webrtc.PUT("/mesh-config", webrtcHandler.UpdateMeshConfig)

		// Peers
		webrtc.GET("/peers", webrtcHandler.ListPeers)
		webrtc.GET("/peers/:id", webrtcHandler.GetPeer)
		webrtc.DELETE("/peers/:id", webrtcHandler.DisconnectPeer)
		webrtc.GET("/peers/stats", webrtcHandler.GetPeerStats)

		// Signaling
		webrtc.POST("/signal", webrtcHandler.PostSignalingMessage)

		// Discovery & NAT
		webrtc.GET("/discover", webrtcHandler.DiscoverPeers)
		webrtc.GET("/nat-type", webrtcHandler.DetectNATType)
	}

	// Phase 10: WebRTC Signaling WebSocket (public for peer connections)
	engine.GET("/api/v1/webrtc/signal/ws", wsHub.HandleWebSocket)

	// ═══ Phase 15: Domain Fronting + CDN Proxy ════════════
	fronting := protected.Group("/domain-fronting")
	fronting.Use(middleware.RoleMiddleware("super_admin", "admin"))
	{
		fronting.GET("/providers", domainFrontingHandler.ListProviders)
		fronting.GET("/scan", domainFrontingHandler.ScanDomain)
		fronting.POST("/scan-all", domainFrontingHandler.ScanAll)
		fronting.GET("/domains", domainFrontingHandler.ListFrontable)
		fronting.GET("/generate-config", domainFrontingHandler.GenerateConfig)
		fronting.DELETE("/domains/:id", domainFrontingHandler.DeleteDomain)
	}

	// ═══ Phase 15: Smart DNS System ═════════════════════════
	dns := protected.Group("/dns")
	dns.Use(middleware.RoleMiddleware("super_admin", "admin"))
	{
		dns.GET("/resolve", smartDNSHandler.ResolveDNS)
		dns.GET("/configs", smartDNSHandler.ListDNSConfigs)
		dns.POST("/configs", smartDNSHandler.SaveDNSConfig)
		dns.DELETE("/configs/:id", smartDNSHandler.DeleteDNSConfig)
		dns.GET("/rules", smartDNSHandler.ListDNSRules)
		dns.POST("/rules", smartDNSHandler.SaveDNSRule)
		dns.DELETE("/rules/:id", smartDNSHandler.DeleteDNSRule)
		dns.POST("/load-ad-block", smartDNSHandler.LoadAdBlockList)
		dns.POST("/clear-cache", smartDNSHandler.ClearDNSCache)
	}

	// ═══ Phase 16: Xray gRPC Real Integration ═══════════
	xrayAPI := protected.Group("/xray")
	xrayAPI.Use(middleware.RoleMiddleware("super_admin", "admin"))
	{
		xrayAPI.GET("/process", xrayAPIHandler.GetProcessInfo)
		xrayAPI.GET("/logs", xrayAPIHandler.GetLogs)
		xrayAPI.POST("/validate", xrayAPIHandler.ValidateConfig)
		xrayAPI.GET("/online-users", xrayAPIHandler.GetOnlineUsers)
		xrayAPI.GET("/traffic", xrayAPIHandler.GetTrafficStats)
		xrayAPI.POST("/test-route", xrayAPIHandler.TestRoute)
		xrayAPI.GET("/balancers/:tag", xrayAPIHandler.GetBalancerInfo)
		xrayAPI.PUT("/balancers/target", xrayAPIHandler.SetBalancerTarget)
	}

	// ═══ Phase 15: Docker Native Mode ═══════════════════════
	docker := protected.Group("/docker")
	docker.Use(middleware.RoleMiddleware("super_admin"))
	{
		docker.GET("/status", dockerHandler.Status)
		docker.GET("/containers", dockerHandler.ListContainers)
		docker.POST("/containers", dockerHandler.CreateContainer)
		docker.POST("/containers/:id/start", dockerHandler.StartContainer)
		docker.POST("/containers/:id/stop", dockerHandler.StopContainer)
		docker.POST("/containers/:id/restart", dockerHandler.RestartContainer)
		docker.DELETE("/containers/:id", dockerHandler.RemoveContainer)
		docker.GET("/containers/:id/logs", dockerHandler.GetContainerLogs)
		docker.GET("/containers/:id/stats", dockerHandler.GetContainerStats)
		docker.GET("/images", dockerHandler.ListImages)
		docker.POST("/images/pull", dockerHandler.PullImage)
		docker.DELETE("/images/:id", dockerHandler.RemoveImage)
		docker.POST("/images/prune", dockerHandler.PruneImages)
	}

	// ─── Phase 12: Smart Health Check System ─────────────────
	hc := protected.Group("/health")
	{
		hc.GET("/configs", healthHandler.ListCheckConfigs)
		hc.POST("/configs", healthHandler.CreateCheckConfig)
		hc.PUT("/configs/:id", healthHandler.UpdateCheckConfig)
		hc.DELETE("/configs/:id", healthHandler.DeleteCheckConfig)
		hc.GET("/statuses", healthHandler.GetStatuses)
		hc.GET("/configs/:id/history", healthHandler.GetHistory)
		hc.POST("/manual-check", healthHandler.RunManualCheck)

		// Recovery Rules
		hc.GET("/recovery-rules", healthHandler.ListRecoveryRules)
		hc.POST("/recovery-rules", healthHandler.CreateRecoveryRule)
		hc.PUT("/recovery-rules/:id", healthHandler.UpdateRecoveryRule)
		hc.DELETE("/recovery-rules/:id", healthHandler.DeleteRecoveryRule)

		// Recovery History
		hc.GET("/recovery-history", healthHandler.GetRecoveryHistory)
	}

	// ─── Static Web UI (embedded) ────────────────────────────────
	distFS, err := fs.Sub(web.Dist, "dist")
	if err == nil {
		engine.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// Try to serve the file, fallback to index.html for SPA routing
			if _, err := distFS.Open(path[1:]); err == nil && path != "/" {
				http.FileServer(http.FS(distFS)).ServeHTTP(c.Writer, c.Request)
			} else {
				c.FileFromFS("index.html", http.FS(distFS))
			}
		})
	} else {
		engine.NoRoute(func(c *gin.Context) {
			c.JSON(404, gin.H{"error": "not found"})
		})
	}

	return &Router{engine: engine, Hub: wsHub}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}
