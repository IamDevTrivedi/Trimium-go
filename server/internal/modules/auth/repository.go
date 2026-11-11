package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"server/internal/database"
)

type SignUpOTPData struct {
	Email    string `json:"email"`
	OTP      string `json:"OTP"`
	Attempts int    `json:"Attempts"`
	Status   string `json:"Status"`
}

type Repository interface {
	StoreSignUpOTP(ctx context.Context, email string, otpData *SignUpOTPData, ttl time.Duration) error
	GetSignUpOTP(ctx context.Context, email string) (*SignUpOTPData, error)
	DeleteSignUpOTP(ctx context.Context, email string) error
}

type repository struct {
	redis *database.Redis
	mongo *database.Mongo
}

func NewRepository(redis *database.Redis, mongo *database.Mongo) Repository {
	return &repository{redis: redis, mongo: mongo}
}

func (r *repository) StoreSignUpOTP(ctx context.Context, email string, otpData *SignUpOTPData, ttl time.Duration) error {
	payload, err := json.Marshal(otpData)
	if err != nil {
		return fmt.Errorf("auth: marshal otp: %w", err)
	}

	key := fmt.Sprintf("signup-otp:%s", email)
	expiration := ttl
	if expiration <= 0 {
		expiration = redis.KeepTTL
	}

	if err := r.redis.Client().Set(ctx, key, payload, expiration).Err(); err != nil {
		return fmt.Errorf("auth: store otp: %w", err)
	}

	return nil
}

func (r *repository) GetSignUpOTP(ctx context.Context, email string) (*SignUpOTPData, error) {
	key := fmt.Sprintf("signup-otp:%s", email)
	payload, err := r.redis.Client().Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrOTPNotFound
		}
		return nil, fmt.Errorf("auth: get otp: %w", err)
	}

	var otpData SignUpOTPData
	if err := json.Unmarshal(payload, &otpData); err != nil {
		return nil, fmt.Errorf("auth: unmarshal otp: %w", err)
	}

	return &otpData, nil
}

func (r *repository) DeleteSignUpOTP(ctx context.Context, email string) error {
	key := fmt.Sprintf("signup-otp:%s", email)
	if err := r.redis.Client().Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("auth: delete otp: %w", err)
	}

	return nil
}
