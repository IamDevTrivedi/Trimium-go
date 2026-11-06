package health

import (
	"github.com/gin-gonic/gin"

	"server/internal/database"
	"server/internal/queue"
)

func Register(router gin.IRouter, redis *database.Redis, mongo *database.Mongo, queue queue.Queue) {
	repo := NewRepository(redis, mongo)
	svc := NewService(repo, queue)
	h := NewHandler(svc)
	rl := NewRateLimiter(redis)

	RegisterRoutes(router, rl, h)
}
