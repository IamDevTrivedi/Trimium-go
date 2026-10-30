package root

import (
	"github.com/gin-gonic/gin"

	"server/internal/database"
)

func Register(router gin.IRouter, redis *database.Redis, mongo *database.Mongo) {
	repo := NewRepository(redis, mongo)
	svc := NewService(repo)
	h := NewHandler(svc)
	rl := NewRateLimiter(redis)

	RegisterRoutes(router, rl, h)
}
