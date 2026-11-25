package crypto

import (
	"fmt"

	"github.com/google/uuid"
)

func NewUUID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("crypto: generate uuid v7: %w", err)
	}

	return id.String(), nil
}
