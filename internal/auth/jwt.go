package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType distinguishes access from refresh tokens.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims carries the JWT payload for panel authentication.
type Claims struct {
	jwt.RegisteredClaims
	Type      TokenType `json:"type"`
	AdminID   int64     `json:"admin_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	SessionID string    `json:"session_id"`
	Scope     []string  `json:"scope,omitempty"`
}

// TokenPair contains an access and a refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// JWTConfig holds signing parameters.
type JWTConfig struct {
	Secret         []byte
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	Issuer         string
}

// DefaultJWTConfig returns sensible defaults for JWT configuration.
func DefaultJWTConfig(secret string) JWTConfig {
	return JWTConfig{
		Secret:     []byte(secret),
		AccessTTL:  24 * time.Hour,
		RefreshTTL: 7 * 24 * time.Hour,
		Issuer:     "vortexuipro",
	}
}

// GenerateTokenPair creates an access and refresh token for the given admin.
func GenerateTokenPair(cfg JWTConfig, adminID int64, username, role string, scope []string) (*TokenPair, error) {
	if len(cfg.Secret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 bytes")
	}
	now := time.Now()
	sessionID := uuid.New().String()

	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", adminID),
			Issuer:    cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessTTL)),
			ID:        uuid.New().String(),
		},
		Type:      AccessToken,
		AdminID:   adminID,
		Username:  username,
		Role:      role,
		SessionID: sessionID,
		Scope:     scope,
	}

	refreshClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", adminID),
			Issuer:    cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.RefreshTTL)),
			ID:        uuid.New().String(),
		},
		Type:      RefreshToken,
		AdminID:   adminID,
		Username:  username,
		Role:      role,
		SessionID: sessionID,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(cfg.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(cfg.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(cfg.AccessTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken parses and validates a JWT token string.
func ValidateToken(cfg JWTConfig, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return cfg.Secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token.
func RefreshAccessToken(cfg JWTConfig, refreshToken string) (*TokenPair, error) {
	claims, err := ValidateToken(cfg, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	if claims.Type != RefreshToken {
		return nil, errors.New("not a refresh token")
	}

	return GenerateTokenPair(cfg, claims.AdminID, claims.Username, claims.Role, claims.Scope)
}
