package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
)

const otpCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	ErrOTPLength  = errors.New("otp has incorrect length")
	ErrOTPCharset = errors.New("otp contains invalid characters")
)

func GenerateOTP(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("otp length must be greater than zero")
	}

	limit := big.NewInt(int64(len(otpCharset)))
	otp := make([]byte, length)

	for i := range otp {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		otp[i] = otpCharset[n.Int64()]
	}

	return string(otp), nil
}

func ValidateOTP(otp string, expectedLength int) error {
	if expectedLength <= 0 {
		return errors.New("expected otp length must be greater than zero")
	}
	if len(otp) != expectedLength {
		return fmt.Errorf("%w: expected %d, got %d", ErrOTPLength, expectedLength, len(otp))
	}
	for i := 0; i < len(otp); i++ {
		if !isOTPChar(otp[i]) {
			return fmt.Errorf("%w: invalid character at position %d", ErrOTPCharset, i)
		}
	}
	return nil
}

func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func isOTPChar(c byte) bool {
	for i := range len(otpCharset) {
		if otpCharset[i] == c {
			return true
		}
	}
	return false
}
