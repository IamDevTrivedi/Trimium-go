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
	"server/internal/database"
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
		fmt.Printf("mailer verification failed: %v\n", err)
	}

	log.Info("Mailer verification succeeded")

	mongoStore, err := database.NewMongo(context.Background(), cfg.MongoURI, cfg.MongoDatabase, database.MongoOptions{
		MaxPoolSize:    cfg.MongoMaxPoolSize,
		MinPoolSize:    cfg.MongoMinPoolSize,
		ConnectTimeout: cfg.MongoConnectTimeout,
		PingTimeout:    cfg.MongoPingTimeout,
	})
	if err != nil {
		return fmt.Errorf("mongo error: %w", err)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()

		if err := mongoStore.Close(closeCtx); err != nil {
			log.Error(fmt.Sprintf("failed to close MongoDB connection: %v", err))
		}
	}()

	log.Info("MongoDB connection established")

	redisStore, err := database.NewRedis(context.Background(), cfg.RedisAddr, cfg.RedisUsername, cfg.RedisPassword, cfg.RedisDB, database.RedisOptions{
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
		PingTimeout:  cfg.RedisPingTimeout,
	})
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}
	defer func() {
		if err := redisStore.Close(); err != nil {
			log.Error(fmt.Sprintf("failed to close Redis connection: %v", err))
		}
	}()

	log.Info("Redis connection established")

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	geo, err := middleware.NewGeoResolver(middleware.GeoResolverConfig{
		DBPath: "./data/geoip/GeoLite2-City.mmdb",
	})
	if err != nil {
		return fmt.Errorf("failed to initialize GeoResolver: %w", err)
	}
	defer geo.Close()

	router.Use(
		gin.Recovery(),
		middleware.RequestID(),
		middleware.RequestIP(),
		middleware.RequestLogger(),
		middleware.RequestUserAgent(),
		middleware.RequestGeo(geo),
		middleware.RateLimiter(100, 60_000, middleware.GetRequestIP),
	)

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
