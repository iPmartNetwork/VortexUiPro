package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/service"
)

// AntiCensorHandler provides anti-censorship tool endpoints.
type AntiCensorHandler struct {
	svc *service.AntiCensorshipService
}

// NewAntiCensorHandler creates a new anti-censorship handler.
func NewAntiCensorHandler(svc *service.AntiCensorshipService) *AntiCensorHandler {
	return &AntiCensorHandler{svc: svc}
}

// ListTricks returns all available TLS tricks.
func (h *AntiCensorHandler) ListTricks(c *gin.Context) {
	tricks := h.svc.GetAvailableTricks()
	c.JSON(http.StatusOK, gin.H{"tricks": tricks, "total": len(tricks)})
}

// ListFingerprints returns available TLS fingerprints.
func (h *AntiCensorHandler) ListFingerprints(c *gin.Context) {
	fps := h.svc.GetTLSFingerprints()
	c.JSON(http.StatusOK, gin.H{"fingerprints": fps})
}

// ScanTarget scans a target host for REALITY compatibility.
func (h *AntiCensorHandler) ScanTarget(c *gin.Context) {
	target := c.Query("target")
	portStr := c.DefaultQuery("port", "443")
	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 443
	}
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target query param required"})
		return
	}

	result := h.svc.ScanTarget(target, port)
	c.JSON(http.StatusOK, result)
}

// GenerateDecoyConfig generates a decoy site config.
func (h *AntiCensorHandler) GenerateDecoyConfig(c *gin.Context) {
	domain := c.Query("domain")
	proxyProto := c.DefaultQuery("proxy_proto", "tls")
	cfg := h.svc.GenerateDecoyConfig(domain, proxyProto)
	c.JSON(http.StatusOK, cfg)
}

// GenerateSelfSignedCert generates a self-signed TLS certificate.
func (h *AntiCensorHandler) GenerateSelfSignedCert(c *gin.Context) {
	domain := c.DefaultQuery("domain", "vortex.local")
	certPEM, keyPEM, err := h.svc.GenerateSelfSignedCert(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"certificate": string(certPEM),
		"private_key": string(keyPEM),
		"domain":      domain,
	})
}

// GetFragmentConfig returns a fragment config.
func (h *AntiCensorHandler) GetFragmentConfig(c *gin.Context) {
	cfg := h.svc.GenerateFragmentConfig()
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

// GetPaddingConfig returns a padding config.
func (h *AntiCensorHandler) GetPaddingConfig(c *gin.Context) {
	cfg := h.svc.GeneratePaddingConfig()
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

// GenerateMixConfig generates mixed HTTPS config.
func (h *AntiCensorHandler) GenerateMixConfig(c *gin.Context) {
	cfg := h.svc.GenerateMixConfig(nil)
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

// GenerateAntiDPIConfig returns a bundled anti-DPI configuration.
func (h *AntiCensorHandler) GenerateAntiDPIConfig(c *gin.Context) {
	transport := c.DefaultQuery("transport", "tcp")
	cfg := h.svc.GenerateAntiDPIConfig(transport)
	c.JSON(http.StatusOK, cfg)
}

// GenerateMTProtoConfig returns an MTProto proxy configuration.
func (h *AntiCensorHandler) GenerateMTProtoConfig(c *gin.Context) {
	cfg := h.svc.GenerateMTProtoConfig()
	c.JSON(http.StatusOK, cfg)
}

// GenerateWarpConfig returns a WARP integration configuration.
func (h *AntiCensorHandler) GenerateWarpConfig(c *gin.Context) {
	cfg := h.svc.GenerateWarpConfig()
	c.JSON(http.StatusOK, cfg)
}

// SaveCert saves generated cert to disk.
func (h *AntiCensorHandler) SaveCert(c *gin.Context) {
	var req struct {
		Certificate string `json:"certificate" binding:"required"`
		PrivateKey  string `json:"private_key" binding:"required"`
		CertPath    string `json:"cert_path"`
		KeyPath     string `json:"key_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "certificate and private_key required"})
		return
	}

	err := h.svc.SaveCertToFile([]byte(req.Certificate), []byte(req.PrivateKey), req.CertPath, req.KeyPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "certificate saved"})
}
