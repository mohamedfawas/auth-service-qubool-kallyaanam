// internal/repository/repository.go
package repository

import (
	"context"

	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model"
)

type UserRepository interface {
	CreatePendingRegistration(ctx context.Context, registration *model.PendingRegistration) error
	FindUserByEmail(ctx context.Context, email string) (bool, error)
	FindUserByPhone(ctx context.Context, phone string) (bool, error)
	FindPendingRegistrationByEmail(ctx context.Context, email string) (bool, error)
	FindPendingRegistrationByPhone(ctx context.Context, phone string) (bool, error)
	// WithTransaction executes operations within a transaction
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}
