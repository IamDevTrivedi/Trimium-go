package crypto

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const otpCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateOTP(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("otp length must be greater than zero")
	}

	max := big.NewInt(int64(len(otpCharset)))
	otp := make([]byte, length)

	for i := range otp {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		otp[i] = otpCharset[n.Int64()]
	}

	return string(otp), nil
}
