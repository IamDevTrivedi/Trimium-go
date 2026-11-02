package auth

import (
	"server/internal/database"
)

type RateLimiter interface {
}

type rateLimiter struct {
}

func NewRateLimiter(_ *database.Redis) RateLimiter {
	return &rateLimiter{}
}
