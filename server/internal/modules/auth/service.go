package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"server/internal/crypto"
	"server/internal/emailtemplates"
	"server/internal/queue"
	"server/internal/tasks"
)

var (
	ErrEmailAlreadyRegistered   = errors.New("auth: email already registered")
	ErrEmailNotFound            = errors.New("auth: email not found")
	ErrUsernameAlreadyTaken     = errors.New("auth: username already taken")
	ErrSignUpOTPNotFound        = errors.New("auth: signup otp not found")
	ErrSignUpNotVerified        = errors.New("auth: signup not verified")
	ErrResetPasswordOTPNotFound = errors.New("auth: reset password otp not found")
	ErrResetPasswordNotVerified = errors.New("auth: reset password not verified")
	ErrInvalidOTP               = errors.New("auth: invalid otp")
	ErrMaxVerificationAttempts  = errors.New("auth: max verification attempts exceeded")
)

type Service interface {
	SignUpInit(ctx context.Context, email, ip, location, ua string) error
	SignUpVerify(ctx context.Context, email, otp string) error
	SignUpCreate(ctx context.Context, firstName, lastName, email, username, password string) error
	ResetPasswordInit(ctx context.Context, email, ip, location, ua string) error
	ResetPasswordVerify(ctx context.Context, email, otp string) error
	ResetPassword(ctx context.Context, email, password string) error
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

func (s *service) SignUpInit(ctx context.Context, email, ip, location, ua string) error {
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return fmt.Errorf("auth: check email: %w", err)
	}
	if exists {
		return ErrEmailAlreadyRegistered
	}

	otp, err := crypto.GenerateOTP(signUpOTPLength)
	if err != nil {
		return fmt.Errorf("auth: generate signup otp: %w", err)
	}

	err = s.repo.StoreSignUpOTP(ctx, email, signUpOTP{
		OTP:      otp,
		Attempts: 0,
		Status:   notVerified,
	}, otpVerificationTTL)
	if err != nil {
		return fmt.Errorf("auth: store signup otp: %w", err)
	}

	subject, htmlBody, err := emailtemplates.SignUpOTPEmail(emailtemplates.SignUpOTPData{
		Email:    email,
		OTP:      otp,
		TTL:      otpVerificationTTL,
		IP:       ip,
		Location: location,
		UA:       ua,
	})
	if err != nil {
		return fmt.Errorf("auth: render signup otp email: %w", err)
	}

	err = s.queue.SendEmail(ctx, tasks.EmailPayload{
		ToEmail:  []string{email},
		Subject:  subject,
		HTMLBody: htmlBody,
	})
	if err != nil {
		return fmt.Errorf("auth: enqueue signup otp email: %w", err)
	}

	return nil
}

func (s *service) SignUpVerify(ctx context.Context, email, otp string) error {
	record, err := s.repo.GetSignUpOTP(ctx, email)
	if err != nil {
		return fmt.Errorf("auth: get signup otp: %w", err)
	}

	if crypto.ConstantTimeEqual(record.OTP, otp) {
		record.Status = verified
		if err := s.repo.StoreSignUpOTP(ctx, email, *record, accountCreationTTL); err != nil {
			return fmt.Errorf("auth: mark signup verified: %w", err)
		}

		return nil
	}

	record.Attempts++
	if record.Attempts >= maxVerificationAttempts {
		if err := s.repo.DeleteSignUpOTP(ctx, email); err != nil {
			return fmt.Errorf("auth: delete exhausted signup otp: %w", err)
		}

		return ErrMaxVerificationAttempts
	}

	if err := s.repo.UpdateSignUpOTP(ctx, email, *record); err != nil {
		return fmt.Errorf("auth: update signup otp attempts: %w", err)
	}

	return ErrInvalidOTP
}

func (s *service) SignUpCreate(ctx context.Context, firstName, lastName, email, username, password string) error {
	record, err := s.repo.GetSignUpOTP(ctx, email)
	if err != nil {
		if errors.Is(err, ErrSignUpOTPNotFound) {
			return ErrSignUpNotVerified
		}

		return fmt.Errorf("auth: get signup otp: %w", err)
	}
	if record.Status != verified {
		return ErrSignUpNotVerified
	}

	emailExists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return fmt.Errorf("auth: check email: %w", err)
	}
	if emailExists {
		return ErrEmailAlreadyRegistered
	}

	usernameExists, err := s.repo.UsernameExists(ctx, username)
	if err != nil {
		return fmt.Errorf("auth: check username: %w", err)
	}
	if usernameExists {
		return ErrUsernameAlreadyTaken
	}

	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		return fmt.Errorf("auth: hash password: %w", err)
	}

	now := time.Now().UnixMilli()
	err = s.repo.CreateUser(ctx, user{
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return fmt.Errorf("auth: create user: %w", err)
	}

	return nil
}

func (s *service) ResetPasswordInit(ctx context.Context, email, ip, location, ua string) error {
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return fmt.Errorf("auth: check email: %w", err)
	}
	if !exists {
		return ErrEmailNotFound
	}

	otp, err := crypto.GenerateOTP(resetPasswordOTPLength)
	if err != nil {
		return fmt.Errorf("auth: generate reset password otp: %w", err)
	}

	err = s.repo.StoreResetPasswordOTP(ctx, email, resetPasswordOTP{
		OTP:      otp,
		Attempts: 0,
		Status:   notVerified,
	}, otpVerificationTTL)
	if err != nil {
		return fmt.Errorf("auth: store reset password otp: %w", err)
	}

	subject, htmlBody, err := emailtemplates.ResetPasswordOTPEmail(emailtemplates.ResetPasswordOTPData{
		Email:    email,
		OTP:      otp,
		TTL:      otpVerificationTTL,
		IP:       ip,
		Location: location,
		UA:       ua,
	})
	if err != nil {
		return fmt.Errorf("auth: render reset password otp email: %w", err)
	}

	err = s.queue.SendEmail(ctx, tasks.EmailPayload{
		ToEmail:  []string{email},
		Subject:  subject,
		HTMLBody: htmlBody,
	})
	if err != nil {
		return fmt.Errorf("auth: enqueue reset password otp email: %w", err)
	}

	return nil
}

func (s *service) ResetPasswordVerify(ctx context.Context, email, otp string) error {
	record, err := s.repo.GetResetPasswordOTP(ctx, email)
	if err != nil {
		return fmt.Errorf("auth: get reset password otp: %w", err)
	}

	if crypto.ConstantTimeEqual(record.OTP, otp) {
		record.Status = verified
		if err := s.repo.StoreResetPasswordOTP(ctx, email, *record, resetPasswordTTL); err != nil {
			return fmt.Errorf("auth: mark reset password verified: %w", err)
		}

		return nil
	}

	record.Attempts++
	if record.Attempts >= maxVerificationAttempts {
		if err := s.repo.DeleteResetPasswordOTP(ctx, email); err != nil {
			return fmt.Errorf("auth: delete exhausted reset password otp: %w", err)
		}

		return ErrMaxVerificationAttempts
	}

	if err := s.repo.UpdateResetPasswordOTP(ctx, email, *record); err != nil {
		return fmt.Errorf("auth: update reset password otp attempts: %w", err)
	}

	return ErrInvalidOTP
}

func (s *service) ResetPassword(ctx context.Context, email, password string) error {
	record, err := s.repo.GetResetPasswordOTP(ctx, email)
	if err != nil {
		if errors.Is(err, ErrResetPasswordOTPNotFound) {
			return ErrResetPasswordNotVerified
		}

		return fmt.Errorf("auth: get reset password otp: %w", err)
	}
	if record.Status != verified {
		return ErrResetPasswordNotVerified
	}

	emailExists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return fmt.Errorf("auth: check email: %w", err)
	}
	if !emailExists {
		return ErrEmailNotFound
	}

	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		return fmt.Errorf("auth: hash password: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, email, passwordHash); err != nil {
		return fmt.Errorf("auth: update password: %w", err)
	}

	if err := s.repo.DeleteResetPasswordOTP(ctx, email); err != nil {
		return fmt.Errorf("auth: cleanup reset password otp: %w", err)
	}

	return nil
}
