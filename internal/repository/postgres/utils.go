// internal/repository/postgres/utils.go

package postgres

import (
	"errors"
	"strings"

	"github.com/jackc/pgconn"
)

// Error codes for PostgreSQL constraint violations
const (
	UniqueViolationCode = "23505"
)

// IsUniqueConstraintViolation checks if error is a unique constraint violation
// and optionally checks against a specific constraint name
func IsUniqueConstraintViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == UniqueViolationCode {
		// If no specific constraint name is provided, any unique violation matches
		if constraintName == "" {
			return true
		}
		// Otherwise check if the constraint name matches
		return strings.Contains(pgErr.ConstraintName, constraintName)
	}
	return false
}
