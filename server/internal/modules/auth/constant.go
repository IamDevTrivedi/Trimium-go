package auth

import "time"

const verified = "VERIFIED"
const notVerified = "NOT_VERIFIED"
const otpVerificationTTL = 5 * time.Minute
const accountCreationTTL = 15 * time.Minute
const resetPasswordTTL = 15 * time.Minute

const signUpOTPLength = 6
const resetPasswordOTPLength = 6
const maxVerificationAttempts = 3

const signUpOTPKeyPrefix = "auth:signup-otp:"
const resetPasswordOTPKeyPrefix = "auth:reset-password-otp:" //nolint:gosec

const (
	minNameLength     = 1
	maxNameLength     = 50
	minUsernameLength = 3
	maxUsernameLength = 30
	minPasswordLength = 10
	maxPasswordLength = 128
)

const usersCollection = "users"
