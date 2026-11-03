package main

import (
	"fmt"
	"server/internal/config"
	"server/internal/logger"
	"server/internal/tasks"
	"server/internal/workers"

	"github.com/hibiken/asynq"
)

const (
	asynqConcurrency = 10
)

func main() {
	cfg, err := config.Get()

	if err != nil {
		fmt.Printf("config error: %v", err)
		return
	}

	log := logger.Get()

	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:         fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
			Username:     cfg.RedisUsername,
			Password:     cfg.RedisPassword,
			DB:           cfg.RedisDB,
			DialTimeout:  cfg.RedisDialTimeout,
			ReadTimeout:  cfg.RedisReadTimeout,
			WriteTimeout: cfg.RedisWriteTimeout,
			PoolSize:     cfg.RedisPoolSize,
		},
		asynq.Config{Concurrency: asynqConcurrency},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeEmailSend, workers.HandleEmailSend)

	if err := server.Run(mux); err != nil {
		log.Error(fmt.Sprintf("could not run asynq server: %v", err))
	}
}
