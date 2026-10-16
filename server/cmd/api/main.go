package main

import (
	"context"
	"fmt"
	"os"
	"server/internal/config"
	"server/internal/logger"
	"server/internal/mailer"
	"time"
)

func main() {
	_, err := config.Get()

	if err != nil {
		fmt.Printf("config error: %v\n", err)
		os.Exit(1)
	}

	logger := logger.Get()
	logger.Info("Environment is valid")

	emailCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = mailer.Get().Verify(emailCtx)

	if err != nil {
		logger.Error(fmt.Sprintf("Mailer verification failed: %v", err))
		os.Exit(1)
	}

	logger.Info("Mailer verification succeeded")
}
