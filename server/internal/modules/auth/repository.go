package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"server/internal/database"
)

type SignUpOTPData struct {
	Email    string `json:"email"`
	OTP      string `json:"OTP"`
	Attempts int    `json:"Attempts"`
	Status   string `json:"Status"`
}

type Repository interface {
	StoreOTP(ctx context.Context, email string, otpData *SignUpOTPData, ttl time.Duration) error
}

type repository struct {
	redis *database.Redis
	mongo *database.Mongo
}

func NewRepository(redis *database.Redis, mongo *database.Mongo) Repository {
	return &repository{redis: redis, mongo: mongo}
}

func (r *repository) StoreOTP(ctx context.Context, email string, otpData *SignUpOTPData, ttl time.Duration) error {
	payload, err := json.Marshal(otpData)
	if err != nil {
		return fmt.Errorf("auth: marshal otp: %w", err)
	}

	key := fmt.Sprintf("signup-otp:%s", email)
	if err := r.redis.Client().Set(ctx, key, payload, ttl).Err(); err != nil {
		return fmt.Errorf("auth: store otp: %w", err)
	}

	return nil
}
