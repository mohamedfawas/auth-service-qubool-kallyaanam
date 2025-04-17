// internal/service/security_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Update SecurityConfig struct to include JWT settings
type SecurityConfig struct {
	BcryptCost       int
	MinPasswordChars int
	JWTSecret        string
	TokenExpiry      time.Duration
	RefreshExpiry    time.Duration
	Issuer           string
}

// Implementation of the SecurityService interface
type securityService struct {
	config       SecurityConfig
	redisService RedisService
}

func NewSecurityService(config SecurityConfig, redisService RedisService) (SecurityService, error) {
	// Validate JWT secret length (at least 32 characters)
	if len(config.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT secret is too short, must be at least 32 characters")
	}

	return &securityService{
		config:       config,
		redisService: redisService,
	}, nil
}

// SanitizeInput cleans input to prevent XSS
func (s *securityService) SanitizeInput(ctx context.Context, input string) string {
	// Escape HTML
	sanitized := html.EscapeString(input)
	// Trim spaces
	sanitized = strings.TrimSpace(sanitized)
	return sanitized
}

// ValidatePassword checks if a password meets security requirements
func (s *securityService) ValidatePassword(ctx context.Context, password string) (bool, string) {
	if len(password) < s.config.MinPasswordChars {
		return false, "Password must be at least " + string(rune(s.config.MinPasswordChars)) + " characters long"
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "Password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "Password must contain at least one lowercase letter"
	}
	if !hasNumber {
		return false, "Password must contain at least one number"
	}
	if !hasSpecial {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}

// HashPassword hashes a password securely
func (s *securityService) HashPassword(ctx context.Context, password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.config.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword checks if a password matches its hash
func (s *securityService) VerifyPassword(ctx context.Context, hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *securityService) GenerateJWT(ctx context.Context, userID, role string, lastLogin time.Time) (string, error) {
	// Create token ID
	tokenID := uuid.New().String()

	// Create roles array based on role string
	roles := []string{role}

	// Create the claims with additional context
	claims := jwt.MapClaims{
		"sub":        userID,                                      // Subject (user ID)
		"roles":      roles,                                       // User roles as array
		"iss":        s.config.Issuer,                             // Issuer
		"iat":        time.Now().Unix(),                           // Issued at
		"exp":        time.Now().Add(s.config.TokenExpiry).Unix(), // Expiry
		"jti":        tokenID,                                     // JWT ID
		"last_login": lastLogin.Unix(),                            // Last login timestamp
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Set header values for better security
	token.Header["kid"] = "auth-key-1" // Key ID for future key rotation

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a refresh token for the authenticated user
func (s *securityService) GenerateRefreshToken(ctx context.Context, userID string) (string, string, error) {
	// Create token ID
	tokenID := uuid.New().String()

	// Create the claims
	claims := jwt.MapClaims{
		"sub": userID,                                        // Subject (user ID)
		"iss": s.config.Issuer,                               // Issuer
		"iat": time.Now().Unix(),                             // Issued at
		"exp": time.Now().Add(s.config.RefreshExpiry).Unix(), // Expiry
		"jti": tokenID,                                       // JWT ID
		"typ": "refresh",                                     // Token type
	}

	// Create the token with explicit HS256 method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Set header values for better security
	token.Header["kid"] = "refresh-key-1" // Key ID for future key rotation

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return tokenString, tokenID, nil
}

func (s *securityService) ExtractTokenID(ctx context.Context, tokenString string) (string, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	tokenID, ok := claims["jti"].(string)
	if !ok {
		return "", errors.New("token ID not found")
	}

	return tokenID, nil
}

// ValidateJWT validates the JWT token and returns the claims
func (s *securityService) ValidateJWT(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		// Check for specific error types
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, errors.New("token not yet valid")
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, errors.New("malformed token")
		}
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	// Extract the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrSignatureInvalid
	}

	// Check expiration time
	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		if time.Now().After(expTime) {
			return nil, errors.New("token has expired")
		}
	} else {
		return nil, errors.New("invalid token: missing expiration claim")
	}

	// Check if token is blacklisted
	tokenID, ok := claims["jti"].(string)
	if !ok {
		return nil, errors.New("invalid token: missing jti claim")
	}

	isBlacklisted, err := s.redisService.IsTokenBlacklisted(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("error checking blacklist: %w", err)
	}

	if isBlacklisted {
		return nil, errors.New("token is blacklisted")
	}

	// Convert to map
	claimsMap := make(map[string]interface{})
	for key, value := range claims {
		claimsMap[key] = value
	}

	return claimsMap, nil
}
