package health

import (
	"context"

	"server/internal/database"
)

type Repository interface {
	PingRedis(ctx context.Context) error
	PingMongo(ctx context.Context) error
}

type repository struct {
	redis *database.Redis
	mongo *database.Mongo
}

func NewRepository(redis *database.Redis, mongo *database.Mongo) Repository {
	return &repository{redis: redis, mongo: mongo}
}

func (r *repository) PingRedis(ctx context.Context) error {
	return r.redis.Client().Ping(ctx).Err()
}

func (r *repository) PingMongo(ctx context.Context) error {
	return r.mongo.Client().Ping(ctx, nil)
}
