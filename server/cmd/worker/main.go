package main

import (
	"fmt"

	"github.com/hibiken/asynq"

	"server/internal/config"
	"server/internal/logger"
	"server/internal/mailer"
	"server/internal/tasks"
	"server/internal/workers"
)

func main() {
	cfg, err := config.Get()

	if err != nil {
		fmt.Printf("config error: %v", err)
		return
	}

	log := logger.Get()

	mailClient, err := mailer.New(mailer.Options{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUsername,
		Password: cfg.SMTPPassword,
		TLS:      mailer.TLSOpportunistic,
	})
	if err != nil {
		log.Error(fmt.Sprintf("mailer init error: %v", err))
		return
	}

	emailWorker := workers.NewEmailWorker(cfg.FromEmail, mailClient)

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
		asynq.Config{Concurrency: cfg.AsynqConcurrency},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeEmailSend, emailWorker.HandleEmailSend)

	if err := server.Run(mux); err != nil {
		log.Error(fmt.Sprintf("could not run asynq server: %v", err))
	}
}
