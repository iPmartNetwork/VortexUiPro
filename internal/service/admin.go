package service

import (
	"fmt"
	"time"

	"vortexuipro/internal/auth"
	"vortexuipro/internal/database"
	"vortexuipro/internal/domain"
	"vortexuipro/internal/events"
)

// AdminService manages panel administrators and authentication.
type AdminService struct {
	authCfg  auth.JWTConfig
	eventBus events.Publisher
}

// NewAdminService creates a new admin service with database backend.
func NewAdminService(jwtSecret string, bus events.Publisher) *AdminService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &AdminService{
		authCfg: auth.DefaultJWTConfig(jwtSecret),
		eventBus: bus,
	}
}

// Login authenticates an admin and returns JWT tokens.
func (s *AdminService) Login(username, password string) (*auth.TokenPair, error) {
	admin, err := database.GetAdminByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if locked
	if admin.LockedUntil > time.Now().UnixMilli() {
		return nil, fmt.Errorf("account locked. try again later")
	}

	// Verify password
	if !auth.VerifyPassword(password, admin.PasswordHash) {
		admin.LoginAttempts++
		if admin.LoginAttempts >= 5 {
			admin.LockedUntil = time.Now().Add(15 * time.Minute).UnixMilli()
		}
		database.UpdateAdmin(admin)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Reset attempts on success
	admin.LoginAttempts = 0
	admin.LockedUntil = 0
	database.UpdateAdmin(admin)

	tokens, err := auth.GenerateTokenPair(s.authCfg, admin.ID, admin.Username, admin.Role, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}
	return tokens, nil
}

// ValidateToken validates a JWT and returns the claims.
func (s *AdminService) ValidateToken(tokenString string) (*auth.Claims, error) {
	return auth.ValidateToken(s.authCfg, tokenString)
}

// RefreshToken generates a new access token from a refresh token.
func (s *AdminService) RefreshToken(refreshToken string) (*auth.TokenPair, error) {
	return auth.RefreshAccessToken(s.authCfg, refreshToken)
}

// CreateAdmin adds a new admin user.
func (s *AdminService) CreateAdmin(username, password string, role domain.AdminRole) (*database.Admin, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	admin := &database.Admin{
		Username:     username,
		PasswordHash: hash,
		Role:         string(role),
	}
	if err := database.CreateAdmin(admin); err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}
	return admin, nil
}

// GetAdmin retrieves an admin by ID.
func (s *AdminService) GetAdmin(id int64) (*database.Admin, error) {
	return database.GetAdminByID(id)
}

// ListAdmins returns all admins.
func (s *AdminService) ListAdmins() ([]database.Admin, error) {
	return database.ListAdmins()
}

// ChangePassword updates an admin's password.
func (s *AdminService) ChangePassword(id int64, oldPassword, newPassword string) error {
	admin, err := database.GetAdminByID(id)
	if err != nil {
		return fmt.Errorf("admin not found")
	}
	if !auth.VerifyPassword(oldPassword, admin.PasswordHash) {
		return fmt.Errorf("invalid current password")
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	admin.PasswordHash = hash
	return database.UpdateAdmin(admin)
}
