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

type MongoOptions struct {
	URI            string
	Database       string
	MaxPoolSize    uint64
	MinPoolSize    uint64
	ConnectTimeout time.Duration
	PingTimeout    time.Duration
}

type Mongo struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongo(ctx context.Context, opts MongoOptions) (*Mongo, error) {
	if opts.URI == "" {
		return nil, errors.New("mongo: uri is required")
	}
	if opts.Database == "" {
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
		ApplyURI(opts.URI).
		SetMaxPoolSize(opts.MaxPoolSize).
		SetMinPoolSize(opts.MinPoolSize).
		SetConnectTimeout(opts.ConnectTimeout))
	if err != nil {
		return nil, fmt.Errorf("mongo: connect: %w", err)
	}

	m := &Mongo{
		client: client,
		db:     client.Database(opts.Database),
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

func (m *Mongo) Client() *mongo.Client {
	return m.client
}

func (m *Mongo) DB() *mongo.Database {
	return m.db
}

func (m *Mongo) Collection(name string) *mongo.Collection {
	return m.db.Collection(name)
}

func (m *Mongo) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}

func (m *Mongo) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
