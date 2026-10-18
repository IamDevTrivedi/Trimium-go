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
	cfg, err := config.Get()
	if err != nil {
		fmt.Printf("config error: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println(cfg)
	}

	log := logger.Get()
	log.Info("Environment is valid")

	emailCtx, emailCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer emailCancel()

	err = mailer.Get().Verify(emailCtx)
	if err != nil {
		log.Error(fmt.Sprintf("Mailer verification failed: %v", err))
		os.Exit(1)
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
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		log.Info(fmt.Sprintf("Server listening on port %d", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(fmt.Sprintf("Server failed to start: %v", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	log.Info("Server stopped cleanly")
}
