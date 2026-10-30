package root

import "server/internal/database"

// AppInfo is the metadata exposed by the API root endpoint.
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// Repository defines the data-access interface for the root module.
// The root module performs no datastore operations; this empty contract
// is maintained for architectural consistency across modules.
type Repository interface{}

type repository struct {
	redis *database.Redis
	mongo *database.Mongo
}

// NewRepository constructs an empty Repository for the root module.
func NewRepository(redis *database.Redis, mongo *database.Mongo) Repository {
	return &repository{
		redis: redis,
		mongo: mongo,
	}
}
