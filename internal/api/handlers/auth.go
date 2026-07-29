package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/auth"
	"vortexuipro/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	adminSvc    *service.AdminService
	totpManager *auth.TOTPManager
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(adminSvc *service.AdminService) *AuthHandler {
	return &AuthHandler{
		adminSvc:    adminSvc,
		totpManager: auth.NewTOTPManager("VortexUiPro"),
	}
}

// LoginRequest is the login request body.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles user login and returns JWT tokens.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	tokens, err := h.adminSvc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// RefreshRequest is the token refresh request body.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh handles token refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	tokens, err := h.adminSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Me returns the current authenticated user's info.
func (h *AuthHandler) Me(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"admin_id": adminID,
		"username": username,
		"role":     role,
	})
}

// SetupTOTP initiates TOTP enrollment.
func (h *AuthHandler) SetupTOTP(c *gin.Context) {
	username, _ := c.Get("username")
	usernameStr := username.(string)

	secret := h.totpManager.GenerateSecret()
	qrBytes, err := h.totpManager.GenerateQR(secret, usernameStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate QR code"})
		return
	}

	uri := h.totpManager.ProvisioningURI(secret, usernameStr)

	c.JSON(http.StatusOK, gin.H{
		"secret":          secret,
		"provisioning_uri": uri,
		"qr_code":         qrBytes,
	})
}

// ValidateTOTP validates a TOTP code.
type ValidateTOTPRequest struct {
	Secret string `json:"secret" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

func (h *AuthHandler) ValidateTOTP(c *gin.Context) {
	var req ValidateTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "secret and code are required"})
		return
	}

	valid := h.totpManager.Validate(req.Secret, req.Code)
	c.JSON(http.StatusOK, gin.H{"valid": valid})
}

// ChangePassword handles password changes.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	adminID, _ := c.Get("admin_id")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	err := h.adminSvc.ChangePassword(adminID.(int64), req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
