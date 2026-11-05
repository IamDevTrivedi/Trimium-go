package template

import "server/internal/database"

type Repository interface {
}

type repository struct {
	redis *database.Redis
	mongo *database.Mongo
}

func NewRepository(redis *database.Redis, mongo *database.Mongo) Repository {
	return &repository{redis: redis, mongo: mongo}
}
