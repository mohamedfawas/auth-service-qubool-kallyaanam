package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenService handles JWT token operations
type TokenService struct {
	secretKey     []byte
	issuer        string
	cookieName    string
	tokenDuration time.Duration
}

// NewTokenService creates a new token service
func NewTokenService(secretKey, issuer string) *TokenService {
	return &TokenService{
		secretKey:     []byte(secretKey),
		issuer:        issuer,
		cookieName:    "registration_session",
		tokenDuration: 24 * time.Hour,
	}
}

// RegistrationClaims represents JWT claims for registration flow
type RegistrationClaims struct {
	jwt.RegisteredClaims
	PendingID string `json:"pending_id"`
}

// GenerateToken creates a JWT for a pending registration
func (s *TokenService) GenerateToken(pendingID uuid.UUID, expiresAt time.Time) (string, error) {
	claims := RegistrationClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
			Subject:   "registration",
		},
		PendingID: pendingID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ParseToken validates and extracts data from a token
func (s *TokenService) ParseToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&RegistrationClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return s.secretKey, nil
		},
	)

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(*RegistrationClaims); ok && token.Valid {
		pendingID, err := uuid.Parse(claims.PendingID)
		if err != nil {
			return uuid.Nil, err
		}
		return pendingID, nil
	}

	return uuid.Nil, jwt.ErrSignatureInvalid
}

// GenerateRefreshToken creates a refresh token with longer expiry
func (s *TokenService) GenerateRefreshToken(pendingID uuid.UUID) (string, error) {
	expiresAt := time.Now().Add(s.tokenDuration * 2) // Refresh token lasts longer

	claims := RegistrationClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
			Subject:   "refresh",
		},
		PendingID: pendingID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GetCookieName returns the cookie name used for storing tokens
func (s *TokenService) GetCookieName() string {
	return s.cookieName
}
