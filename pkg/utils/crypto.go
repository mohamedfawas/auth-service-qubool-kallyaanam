package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword creates a bcrypt hash from the provided password
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// VerifyPassword compares a bcrypt hashed password with its possible plaintext equivalent
func VerifyPassword(hashedPassword []byte, password string) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
}
