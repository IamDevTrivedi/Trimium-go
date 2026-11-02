package auth

import (
	"server/internal/database"
)

type RateLimiter interface {
}

type rateLimiter struct {
}

func NewRateLimiter(store *database.Redis) RateLimiter {
	return &rateLimiter{}
}
