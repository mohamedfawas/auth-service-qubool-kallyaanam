package service

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// CleanupService handles periodic cleanup of expired data
type CleanupService struct {
	otpService *OTPService
	cron       *cron.Cron
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(otpService *OTPService) *CleanupService {
	return &CleanupService{
		otpService: otpService,
		cron:       cron.New(),
	}
}

// StartCleanupJobs starts the cleanup cron jobs
func (s *CleanupService) StartCleanupJobs() {
	// Run cleanup every hour
	_, err := s.cron.AddFunc("0 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := s.otpService.CleanupExpiredData(ctx)
		if err != nil {
			log.Printf("Error cleaning up expired data: %v", err)
		} else {
			log.Println("Successfully cleaned up expired data")
		}
	})

	if err != nil {
		log.Printf("Error setting up cleanup job: %v", err)
		return
	}

	s.cron.Start()
	log.Println("Cleanup jobs started")
}

// StopCleanupJobs stops the cleanup cron jobs
func (s *CleanupService) StopCleanupJobs() {
	s.cron.Stop()
	log.Println("Cleanup jobs stopped")
}
