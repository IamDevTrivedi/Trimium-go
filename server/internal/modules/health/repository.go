package health

import "context"

// Pinger is the health-check operation required from a datastore client.
// Both database.Redis and database.Mongo satisfy it.
type Pinger interface {
	Ping(ctx context.Context) error
}

type Repository interface {
	PingRedis(ctx context.Context) error
	PingMongo(ctx context.Context) error
}

type repository struct {
	redis Pinger
	mongo Pinger
}

func NewRepository(redis, mongo Pinger) Repository {
	return &repository{redis: redis, mongo: mongo}
}

func (r *repository) PingRedis(ctx context.Context) error {
	return r.redis.Ping(ctx)
}

func (r *repository) PingMongo(ctx context.Context) error {
	return r.mongo.Ping(ctx)
}
