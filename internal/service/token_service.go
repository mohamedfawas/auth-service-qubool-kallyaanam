package service

import (
	"errors"
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

// RefreshToken validates a refresh token and generates a new access token
func (s *TokenService) RefreshToken(refreshToken string) (string, string, error) {
	// Parse the refresh token
	token, err := jwt.ParseWithClaims(
		refreshToken,
		&RegistrationClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return s.secretKey, nil
		},
	)

	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(*RegistrationClaims); ok && token.Valid {
		// Verify it's a refresh token
		if claims.Subject != "refresh" {
			return "", "", errors.New("not a refresh token")
		}

		pendingID, err := uuid.Parse(claims.PendingID)
		if err != nil {
			return "", "", err
		}

		// Generate new access token that expires sooner than refresh token
		newAccessToken, err := s.GenerateToken(pendingID, time.Now().Add(15*time.Minute))
		if err != nil {
			return "", "", err
		}

		// Generate new refresh token to implement refresh token rotation
		newRefreshToken, err := s.GenerateRefreshToken(pendingID)
		if err != nil {
			return "", "", err
		}

		return newAccessToken, newRefreshToken, nil
	}

	return "", "", jwt.ErrSignatureInvalid
}

// ValidateTokens checks if tokens are valid and returns the pendingID
func (s *TokenService) ValidateTokens(accessToken string) (uuid.UUID, bool, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&RegistrationClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return s.secretKey, nil
		},
	)

	if err != nil {
		// Check if the error is due to an expired token
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, true, err
		}
		return uuid.Nil, false, err
	}

	if claims, ok := token.Claims.(*RegistrationClaims); ok && token.Valid {
		pendingID, err := uuid.Parse(claims.PendingID)
		if err != nil {
			return uuid.Nil, false, err
		}
		return pendingID, false, nil
	}

	return uuid.Nil, false, jwt.ErrSignatureInvalid
}

// InvalidateToken adds a token to a blacklist
// In a production environment, this should store the token in Redis with an expiry
func (s *TokenService) InvalidateToken(token string) error {
	// For a full implementation, you'd add the token to a Redis blacklist
	// with an expiry matching the token's expiry time
	// This is simplified for this example
	return nil
}
