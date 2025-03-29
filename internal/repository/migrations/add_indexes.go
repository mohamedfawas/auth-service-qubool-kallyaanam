package migrations

import (
	"log"

	"gorm.io/gorm"
)

// AddIndexes adds database indexes to improve query performance
func AddIndexes(db *gorm.DB) error {
	log.Println("Running migration: AddIndexes")

	// Add index to User.CreatedAt
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)").Error; err != nil {
		return err
	}

	// Add indexes to PendingRegistration
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_pending_reg_email_verified ON pending_registrations(email_verified)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_pending_reg_phone_verified ON pending_registrations(phone_verified)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_pending_reg_created_at ON pending_registrations(created_at)").Error; err != nil {
		return err
	}

	// Add indexes to VerificationOTP
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_verification_otp_pending_id ON verification_otps(pending_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_verification_otp_type ON verification_otps(otp_type)").Error; err != nil {
		return err
	}

	log.Println("Migration completed: AddIndexes")
	return nil
}
