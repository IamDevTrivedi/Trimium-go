package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"

	"server/internal/database"
)

type signUpOTP struct {
	OTP      string `json:"otp"`
	Attempts int    `json:"attempts"`
	Status   string `json:"status"`
}

type user struct {
	FirstName    string `bson:"first_name"`
	LastName     string `bson:"last_name"`
	Email        string `bson:"email"`
	Username     string `bson:"username"`
	PasswordHash string `bson:"password_hash"`
	CreatedAt    int64  `bson:"created_at"`
	UpdatedAt    int64  `bson:"updated_at"`
}

type resetPasswordOTP struct {
	OTP      string `json:"otp"`
	Attempts int    `json:"attempts"`
	Status   string `json:"status"`
}

type Repository interface {
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	CreateUser(ctx context.Context, user user) error
	GetSignUpOTP(ctx context.Context, email string) (*signUpOTP, error)
	StoreSignUpOTP(ctx context.Context, email string, otp signUpOTP, ttl time.Duration) error
	UpdateSignUpOTP(ctx context.Context, email string, otp signUpOTP) error
	DeleteSignUpOTP(ctx context.Context, email string) error
	StoreResetPasswordOTP(ctx context.Context, email string, otp resetPasswordOTP, ttl time.Duration) error
	GetResetPasswordOTP(ctx context.Context, email string) (*resetPasswordOTP, error)
	UpdateResetPasswordOTP(ctx context.Context, email string, otp resetPasswordOTP) error
	DeleteResetPasswordOTP(ctx context.Context, email string) error
	UpdatePassword(ctx context.Context, email, passwordHash string) error
}

type repository struct {
	redis *database.Redis
	mongo *database.Mongo
}

func NewRepository(redis *database.Redis, mongo *database.Mongo) Repository {
	return &repository{redis: redis, mongo: mongo}
}

func (r *repository) EmailExists(ctx context.Context, email string) (bool, error) {
	count, err := r.mongo.Collection(usersCollection).CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, fmt.Errorf("auth: check email existence: %w", err)
	}

	return count > 0, nil
}

func (r *repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	count, err := r.mongo.Collection(usersCollection).CountDocuments(ctx, bson.M{"username": username})
	if err != nil {
		return false, fmt.Errorf("auth: check username existence: %w", err)
	}

	return count > 0, nil
}

func (r *repository) CreateUser(ctx context.Context, user user) error {
	if _, err := r.mongo.Collection(usersCollection).InsertOne(ctx, user); err != nil {
		return fmt.Errorf("auth: create user: %w", err)
	}

	return nil
}

func (r *repository) GetSignUpOTP(ctx context.Context, email string) (*signUpOTP, error) {
	payload, err := r.redis.Client().Get(ctx, signUpOTPKeyPrefix+email).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrSignUpOTPNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("auth: get signup otp: %w", err)
	}

	var otp signUpOTP
	if err := json.Unmarshal(payload, &otp); err != nil {
		return nil, fmt.Errorf("auth: unmarshal signup otp: %w", err)
	}

	return &otp, nil
}

func (r *repository) StoreSignUpOTP(
	ctx context.Context,
	email string,
	otp signUpOTP,
	ttl time.Duration,
) error {
	payload, err := json.Marshal(otp)
	if err != nil {
		return fmt.Errorf("auth: marshal signup otp: %w", err)
	}

	key := signUpOTPKeyPrefix + email
	if err := r.redis.Client().Set(ctx, key, payload, ttl).Err(); err != nil {
		return fmt.Errorf("auth: store signup otp: %w", err)
	}

	return nil
}

func (r *repository) UpdateSignUpOTP(ctx context.Context, email string, otp signUpOTP) error {
	payload, err := json.Marshal(otp)
	if err != nil {
		return fmt.Errorf("auth: marshal signup otp: %w", err)
	}

	key := signUpOTPKeyPrefix + email
	if err := r.redis.Client().Set(ctx, key, payload, redis.KeepTTL).Err(); err != nil {
		return fmt.Errorf("auth: update signup otp: %w", err)
	}

	return nil
}

func (r *repository) DeleteSignUpOTP(ctx context.Context, email string) error {
	if err := r.redis.Client().Del(ctx, signUpOTPKeyPrefix+email).Err(); err != nil {
		return fmt.Errorf("auth: delete signup otp: %w", err)
	}

	return nil
}

func (r *repository) StoreResetPasswordOTP(
	ctx context.Context,
	email string,
	otp resetPasswordOTP,
	ttl time.Duration,
) error {
	payload, err := json.Marshal(otp)
	if err != nil {
		return fmt.Errorf("auth: marshal reset password otp: %w", err)
	}

	key := resetPasswordOTPKeyPrefix + email
	if err := r.redis.Client().Set(ctx, key, payload, ttl).Err(); err != nil {
		return fmt.Errorf("auth: store reset password otp: %w", err)
	}

	return nil
}

func (r *repository) GetResetPasswordOTP(ctx context.Context, email string) (*resetPasswordOTP, error) {
	payload, err := r.redis.Client().Get(ctx, resetPasswordOTPKeyPrefix+email).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrResetPasswordOTPNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("auth: get reset password otp: %w", err)
	}

	var otp resetPasswordOTP
	if err := json.Unmarshal(payload, &otp); err != nil {
		return nil, fmt.Errorf("auth: unmarshal reset password otp: %w", err)
	}

	return &otp, nil
}

func (r *repository) UpdateResetPasswordOTP(ctx context.Context, email string, otp resetPasswordOTP) error {
	payload, err := json.Marshal(otp)
	if err != nil {
		return fmt.Errorf("auth: marshal reset password otp: %w", err)
	}

	key := resetPasswordOTPKeyPrefix + email
	if err := r.redis.Client().Set(ctx, key, payload, redis.KeepTTL).Err(); err != nil {
		return fmt.Errorf("auth: update reset password otp: %w", err)
	}

	return nil
}

func (r *repository) DeleteResetPasswordOTP(ctx context.Context, email string) error {
	if err := r.redis.Client().Del(ctx, resetPasswordOTPKeyPrefix+email).Err(); err != nil {
		return fmt.Errorf("auth: delete reset password otp: %w", err)
	}

	return nil
}

func (r *repository) UpdatePassword(ctx context.Context, email, passwordHash string) error {
	update := bson.M{
		"$set": bson.M{
			"password_hash": passwordHash,
			"updated_at":    time.Now().UnixMilli(),
		},
	}

	_, err := r.mongo.Collection(usersCollection).UpdateOne(ctx, bson.M{"email": email}, update)
	if err != nil {
		return fmt.Errorf("auth: update password: %w", err)
	}

	return nil
}
