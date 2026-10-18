package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"server/internal/config"
	"server/internal/logger"
	"server/internal/mailer"
	"server/internal/middleware"
	"server/internal/modules/health"
	"server/internal/modules/root"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	log := logger.Get()
	log.Info("Environment is valid")

	emailCtx, emailCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer emailCancel()

	if err := mailer.Get().Verify(emailCtx); err != nil {
		return fmt.Errorf("mailer verification failed: %w", err)
	}

	log.Info("Mailer verification succeeded")

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.RequestLogger(), gin.Recovery())

	rootService := root.NewService()
	rootHandler := root.NewHandler(rootService)
	root.RegisterRoutes(router, rootHandler)

	healthService := health.NewService()
	healthHandler := health.NewHandler(healthService)
	health.RegisterRoutes(router, healthHandler)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info(fmt.Sprintf("Server listening on port %d", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case <-quit:
	}

	log.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Info("Server stopped cleanly")

	return nil
}
