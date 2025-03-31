// Package auth provides authentication-related utilities.
package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plain text password using bcrypt.
func HashPassword(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword checks if the provided password matches the hashed password.
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateRandomCode generates a random numeric code of the specified length.
func GenerateRandomCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than zero")
	}

	// Define the maximum number for the first digit (9 for all digits)
	max := big.NewInt(9)

	// Build the format string based on the desired length
	formatString := fmt.Sprintf("%%0%dd", length)

	// Generate a random number with the specified number of digits
	num := big.NewInt(0)
	for i := 0; i < length; i++ {
		// For the first digit, we want 1-9 to avoid leading zeros
		if i == 0 {
			max = big.NewInt(9)
		} else {
			max = big.NewInt(10)
		}

		// Generate a random digit
		digit, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("failed to generate random digit: %w", err)
		}

		// For the first digit, add 1 to make it 1-9
		if i == 0 {
			digit.Add(digit, big.NewInt(1))
		}

		// Add the digit to our number
		num.Mul(num, big.NewInt(10))
		num.Add(num, digit)
	}

	return fmt.Sprintf(formatString, num), nil
}
