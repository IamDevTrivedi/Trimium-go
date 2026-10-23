package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// MongoOptions configures the MongoDB client connection pool and timeouts.
type MongoOptions struct {
	MaxPoolSize    uint64
	MinPoolSize    uint64
	ConnectTimeout time.Duration
	PingTimeout    time.Duration
}

// Mongo wraps a MongoDB client and the default database used by the app.
type Mongo struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewMongo connects to the MongoDB instance at uri, selects database and
// verifies the connection with a ping before returning.
func NewMongo(ctx context.Context, uri, database string, opts MongoOptions) (*Mongo, error) {
	if uri == "" {
		return nil, errors.New("mongo: uri is required")
	}
	if database == "" {
		return nil, errors.New("mongo: database name is required")
	}
	if opts.MaxPoolSize > 0 && opts.MinPoolSize > opts.MaxPoolSize {
		return nil, fmt.Errorf(
			"mongo: min pool size %d cannot exceed max pool size %d",
			opts.MinPoolSize,
			opts.MaxPoolSize,
		)
	}

	client, err := mongo.Connect(options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(opts.MaxPoolSize).
		SetMinPoolSize(opts.MinPoolSize).
		SetConnectTimeout(opts.ConnectTimeout))
	if err != nil {
		return nil, fmt.Errorf("mongo: connect: %w", err)
	}

	m := &Mongo{
		client: client,
		db:     client.Database(database),
	}

	pingCtx := ctx
	if opts.PingTimeout > 0 {
		var cancel context.CancelFunc
		pingCtx, cancel = context.WithTimeout(ctx, opts.PingTimeout)
		defer cancel()
	}

	if err := m.Ping(pingCtx); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo: ping: %w", err)
	}

	return m, nil
}

// Client returns the underlying MongoDB client.
func (m *Mongo) Client() *mongo.Client {
	return m.client
}

// DB returns the default database.
func (m *Mongo) DB() *mongo.Database {
	return m.db
}

// Collection returns a collection from the default database.
func (m *Mongo) Collection(name string) *mongo.Collection {
	return m.db.Collection(name)
}

// Ping checks that the deployment is reachable by pinging the primary.
func (m *Mongo) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}

// Close disconnects the client from MongoDB.
func (m *Mongo) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
