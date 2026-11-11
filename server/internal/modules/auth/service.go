package auth

import (
	"context"
	"fmt"
	"time"

	"server/internal/crypto"
	"server/internal/emailtemplates"
	"server/internal/queue"
	"server/internal/tasks"
)

const (
	otpTTL               = 5 * time.Minute
	otpLength            = 6
	otpStatusVerified    = "VERIFIED"
	otpStatusNotVerified = "NOT_VERIFIED"
)

type Service interface {
	SignUpInit(email string) error
}

type service struct {
	repo  Repository
	queue queue.Queue
}

func NewService(repo Repository, queue queue.Queue) Service {
	return &service{
		repo:  repo,
		queue: queue,
	}
}

func (s *service) SignUpInit(email string) error {
	otp, err := crypto.GenerateOTP(otpLength)
	if err != nil {
		return fmt.Errorf("failed to generate otp: %w", err)
	}

	otpData := &SignUpOTPData{
		Email:    email,
		OTP:      otp,
		Attempts: 0,
		Status:   otpStatusNotVerified,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.repo.StoreOTP(ctx, email, otpData, otpTTL); err != nil {
		return fmt.Errorf("failed to store otp: %w", err)
	}

	subject, htmlBody := emailtemplates.SignUpOTPEmailTemplate(
		emailtemplates.SignUpOTPEmailData{
			Email: email,
			OTP:   otp,
			TTL:   otpTTL,
		},
	)

	payload := tasks.EmailPayload{
		ToEmail:  []string{email},
		Subject:  subject,
		HTMLBody: htmlBody,
	}

	if err := s.queue.SendEmail(ctx, payload); err != nil {
		return fmt.Errorf("failed to enqueue email task: %w", err)
	}

	return nil
}
