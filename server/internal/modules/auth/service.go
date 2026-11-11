package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"server/internal/crypto"
	"server/internal/emailtemplates"
	"server/internal/queue"
	"server/internal/tasks"
)

var (
	ErrOTPNotFound         = errors.New("otp not found or expired")
	ErrInvalidOTP          = errors.New("invalid otp")
	ErrMaxAttemptsExceeded = errors.New("max verification attempts exceeded")
)

const (
	otpTTL               = 5 * time.Minute
	otpLength            = 6
	otpMaxAttempts       = 3
	otpStatusVerified    = "VERIFIED"
	otpStatusNotVerified = "NOT_VERIFIED"
)

type Service interface {
	SignUpInit(email string) error
	SignUpVerify(email, otp string) error
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

	if err := s.repo.StoreSignUpOTP(ctx, email, otpData, otpTTL); err != nil {
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

func (s *service) SignUpVerify(email, otp string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	otpData, err := s.repo.GetSignUpOTP(ctx, email)
	if err != nil {
		return err
	}

	if otpData.Status == otpStatusVerified {
		return nil
	}

	if subtle.ConstantTimeCompare([]byte(otpData.OTP), []byte(otp)) != 1 {
		otpData.Attempts++
		if otpData.Attempts >= otpMaxAttempts {
			if err := s.repo.DeleteSignUpOTP(ctx, email); err != nil {
				return fmt.Errorf("failed to delete otp after max attempts: %w", err)
			}
			return ErrMaxAttemptsExceeded
		}

		if err := s.repo.StoreSignUpOTP(ctx, email, otpData, 0); err != nil {
			return fmt.Errorf("failed to update otp attempts: %w", err)
		}

		return ErrInvalidOTP
	}

	otpData.Status = otpStatusVerified
	if err := s.repo.StoreSignUpOTP(ctx, email, otpData, 0); err != nil {
		return fmt.Errorf("failed to update otp status: %w", err)
	}

	return nil
}
