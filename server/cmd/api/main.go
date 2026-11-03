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
	"server/internal/middleware"
	"server/internal/modules/health"
	"server/internal/modules/root"
)

func main() {
	log := logger.Get()
	if err := run(); err != nil {
		log.Error(fmt.Sprintf("Application error: %v", err))
		os.Exit(1)
	}
}

func run() error {
	// getting the logger instance
	log := logger.Get()
	log.Info("Initialize the process")

	// loading the application configuration
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	log.Info("Environment is valid")

	// making a connection with the mongo database
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

	// making a connection with the redis database
	redisStore, err := database.NewRedis(context.Background(), cfg.RedisURI, database.RedisOptions{
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

	// setting the gin mode based on the environment
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// registering middlewares
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
		middleware.RateLimiter(
			middleware.RateLimiterConfig{
				Client: redisStore.Client(),
				Limit:  1000,
				Window: time.Minute,
				KeyGen: middleware.GetRequestIP,
			},
		),
	)

	// registering routes
	root.Register(router, redisStore, mongoStore)
	health.Register(router, redisStore, mongoStore)

	// starting the HTTP server
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// starting the server in a separate goroutine to allow graceful shutdown
	serverErr := make(chan error, 1)

	go func() {
		log.Info(fmt.Sprintf("Server listening on port %d", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// waiting for termination signal to gracefully shutdown the server
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
