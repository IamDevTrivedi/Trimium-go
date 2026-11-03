package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"

	"server/internal/database"
)

func Register(router gin.IRouter, redis *database.Redis, mongo *database.Mongo, queue *asynq.Client) {
	repo := NewRepository(redis, mongo)
	svc := NewService(repo, queue)
	h := NewHandler(svc)
	rl := NewRateLimiter(redis)

	RegisterRoutes(router, rl, h)
}
