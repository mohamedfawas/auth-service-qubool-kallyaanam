package service

import (
	"context"
	"log"
	"time"

	"github.com/mohamedfawas/auth-service-qubool-kallyaanam/internal/repository"
	"github.com/robfig/cron/v3"
)

// CleanupService handles periodic cleanup of expired data
type CleanupService struct {
	regRepo repository.RegistrationRepository
	cron    *cron.Cron
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(regRepo repository.RegistrationRepository) *CleanupService {
	return &CleanupService{
		regRepo: regRepo,
		cron:    cron.New(),
	}
}

// StartCleanupJobs starts the cleanup cron jobs
func (s *CleanupService) StartCleanupJobs() {
	// Run cleanup every hour
	_, err := s.cron.AddFunc("0 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// We'll create a simple method that only cleans up expired registrations
		err := s.CleanupExpiredRegistrations(ctx)
		if err != nil {
			log.Printf("Error cleaning up expired registrations: %v", err)
		} else {
			log.Println("Successfully cleaned up expired registrations")
		}
	})

	if err != nil {
		log.Printf("Error setting up cleanup job: %v", err)
		return
	}

	s.cron.Start()
	log.Println("Cleanup jobs started")
}

// CleanupExpiredRegistrations removes expired registrations
func (s *CleanupService) CleanupExpiredRegistrations(ctx context.Context) error {
	// Assuming you have a method to clean up expired registrations in your repository
	// If you don't have this method yet, you'll need to add it
	count, err := s.regRepo.CleanExpiredRegistrations(ctx)
	if err != nil {
		return err
	}

	log.Printf("Cleaned up %d expired registrations", count)
	return nil
}

// StopCleanupJobs stops the cleanup cron jobs
func (s *CleanupService) StopCleanupJobs() {
	s.cron.Stop()
	log.Println("Cleanup jobs stopped")
}
