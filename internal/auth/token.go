package auth

import (
	"errors"
	"time"

	"ticket-booking/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

// --- Interface for tokenClaims ---
type TokenClaims interface {
	jwt.Claims
	IsAccess() bool
}

// --- Claims ---
type AccessClaims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (a AccessClaims) IsAccess() bool { return true }

type RefreshClaims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

func (r RefreshClaims) IsAccess() bool { return false }

// --- Generate Tokens ---
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GenerateTokens creates both AccessToken and RefreshToken using separate secrets and TTLs
func GenerateTokens(cfg *config.Security, userID, role string) (*Tokens, error) {
	now := time.Now()

	accessClaims := AccessClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ticket-booking",
			Audience:  []string{"ticket-booking-client"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * time.Duration(cfg.AccessTTLMinute))),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshClaims := RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ticket-booking",
			Audience:  []string{"ticket-booking-client"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * time.Duration(cfg.RefreshTTLMinute))),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	at, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(cfg.JWTAccessSecret))
	if err != nil {
		return nil, err
	}
	rt, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(cfg.JWTRefreshSecret))
	if err != nil {
		return nil, err
	}

	return &Tokens{AccessToken: at, RefreshToken: rt}, nil
}

// --- Validate Tokens ---
func ValidateAccessToken(cfg *config.Security, token string) (*AccessClaims, error) {
	claims, err := validateToken(token, true, cfg)
	if err != nil {
		return nil, err
	}
	return claims.(*AccessClaims), nil
}

func ValidateRefreshToken(cfg *config.Security, token string) (*RefreshClaims, error) {
	claims, err := validateToken(token, false, cfg)
	if err != nil {
		return nil, err
	}
	return claims.(*RefreshClaims), nil
}

// --- Internal helper ---
func validateToken(token string, isAccess bool, cfg *config.Security) (TokenClaims, error) {
	var secret string
	if isAccess {
		secret = cfg.JWTAccessSecret
	} else {
		secret = cfg.JWTRefreshSecret
	}

	if secret == "" {
		return nil, errors.New("JWT secret not set")
	}

	var claims TokenClaims
	if isAccess {
		claims = &AccessClaims{}
	} else {
		claims = &RefreshClaims{}
	}

	tok, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
