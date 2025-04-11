// internal/repository/postgres/user_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/model"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/repository"
	"gorm.io/gorm"
)

// Add error types
var (
	ErrDuplicateKey      = errors.New("duplicate key violation")
	ErrTransactionFailed = errors.New("transaction failed")
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// CreatePendingRegistration creates a new pending registration
func (r *UserRepository) CreatePendingRegistration(ctx context.Context, registration *model.PendingRegistration) error {
	// Check if there's a transaction in the context
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx.Create(registration).Error
	}

	return r.db.WithContext(ctx).Create(registration).Error
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) FindUserByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) FindPendingRegistrationByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PendingRegistration{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) FindPendingRegistrationByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PendingRegistration{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

// internal/repository/postgres/user_repository.go

// GetPendingRegistrationByEmail retrieves pending registration by email
func (r *UserRepository) GetPendingRegistrationByEmail(ctx context.Context, email string) (*model.PendingRegistration, error) {
	var pendingReg model.PendingRegistration

	result := r.db.WithContext(ctx).Where("email = ?", email).First(&pendingReg)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No pending registration found
		}
		return nil, result.Error
	}

	return &pendingReg, nil
}

// DeletePendingRegistration deletes a pending registration by ID
func (r *UserRepository) DeletePendingRegistration(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.PendingRegistration{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("pending registration not found")
	}

	return nil
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Create(user)
	return result.Error
}

// internal/repository/postgres/user_repository.go

// GetPendingRegistrationByEmailWithLock retrieves pending registration by email with row locking
func (r *UserRepository) GetPendingRegistrationByEmailWithLock(ctx context.Context, email string) (*model.PendingRegistration, error) {
	var pendingReg model.PendingRegistration

	// Use FOR UPDATE to lock the row and prevent concurrent modifications
	result := r.db.WithContext(ctx).
		Set("gorm:query_option", "FOR UPDATE").
		Where("email = ?", email).
		First(&pendingReg)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No pending registration found
		}
		return nil, result.Error
	}

	return &pendingReg, nil
}

// Improve WithTransaction to support retries
func (r *UserRepository) WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	// Start transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Use defer to ensure tx.Rollback() is executed if tx.Commit() is not called
	// tx.Rollback() is a no-op if the transaction has already been committed
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-throw panic after rollback
		}
	}()

	// Create new context with transaction
	txCtx := context.WithValue(ctx, "tx", tx)

	// Execute the provided function
	if err := fn(txCtx); err != nil {
		tx.Rollback()

		// Check for PostgreSQL error codes and translate to domain errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Handle unique constraint violations
			if pgErr.Code == "23505" { // Unique violation in PostgreSQL
				return ErrDuplicateKey
			}
		}

		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
