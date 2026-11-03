package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string        `env:"APP_ENV,required,notEmpty" validate:"oneof=production development"`
	Port                int           `env:"PORT,required,notEmpty"              validate:"min=1,max=65535"`
	BackendURLDev       string        `env:"BACKEND_URL_DEV,required,notEmpty"   validate:"http_url"`
	FrontendURLDev      string        `env:"FRONTEND_URL_DEV,required,notEmpty"  validate:"http_url"`
	BackendURLProd      string        `env:"BACKEND_URL_PROD,required,notEmpty"  validate:"http_url"`
	FrontendURLProd     string        `env:"FRONTEND_URL_PROD,required,notEmpty" validate:"http_url"`
	SMTPHost            string        `env:"SMTP_HOST,required,notEmpty"     validate:"hostname|ip"`
	SMTPPort            int           `env:"SMTP_PORT,required,notEmpty"     validate:"min=1,max=65535"`
	SMTPUsername        string        `env:"SMTP_USERNAME,required"`
	SMTPPassword        string        `env:"SMTP_PASSWORD,required"`
	FromEmail           string        `env:"FROM_EMAIL,required,notEmpty"    validate:"email"`
	GeoIPDBPath         string        `env:"GEOIP_DB_PATH,required,notEmpty" validate:"file"`
	MongoURI            string        `env:"MONGO_URI,required,notEmpty"             validate:"uri"`
	MongoDatabase       string        `env:"MONGO_DATABASE,required,notEmpty"`
	MongoMaxPoolSize    uint64        `env:"MONGO_MAX_POOL_SIZE,required,notEmpty"   validate:"min=1"`
	MongoMinPoolSize    uint64        `env:"MONGO_MIN_POOL_SIZE,required,notEmpty"   validate:"ltefield=MongoMaxPoolSize"`
	MongoConnectTimeout time.Duration `env:"MONGO_CONNECT_TIMEOUT,required,notEmpty" validate:"gt=0"`
	MongoPingTimeout    time.Duration `env:"MONGO_PING_TIMEOUT,required,notEmpty"    validate:"gt=0"`
	RedisHost           string        `env:"REDIS_HOST,required,notEmpty"           validate:"hostname|ip"`
	RedisPort           int           `env:"REDIS_PORT,required,notEmpty"           validate:"min=1,max=65535"`
	RedisUsername       string        `env:"REDIS_USERNAME"`
	RedisPassword       string        `env:"REDIS_PASSWORD"`
	RedisDB             int           `env:"REDIS_DB,required,notEmpty"             validate:"min=0,max=15"`
	RedisPoolSize       int           `env:"REDIS_POOL_SIZE,required,notEmpty"      validate:"min=1"`
	RedisMinIdleConns   int           `env:"REDIS_MIN_IDLE_CONNS,required,notEmpty" validate:"min=0,ltefield=RedisPoolSize"`
	RedisDialTimeout    time.Duration `env:"REDIS_DIAL_TIMEOUT,required,notEmpty"   validate:"gt=0"`
	RedisReadTimeout    time.Duration `env:"REDIS_READ_TIMEOUT,required,notEmpty"   validate:"gt=0"`
	RedisWriteTimeout   time.Duration `env:"REDIS_WRITE_TIMEOUT,required,notEmpty"  validate:"gt=0"`
	RedisPingTimeout    time.Duration `env:"REDIS_PING_TIMEOUT,required,notEmpty"   validate:"gt=0"`
	AsynqConcurrency    int           `env:"ASYNQ_CONCURRENCY,required,notEmpty"    validate:"min=1"`
}

var (
	cached  *Config
	loadErr error
	once    sync.Once
)

func Get() (*Config, error) {
	once.Do(func() {
		cached, loadErr = load()
	})
	return cached, loadErr
}

func load() (*Config, error) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv != "production" && appEnv != "development" {
		return nil, fmt.Errorf("APP_ENV must be exactly 'production' or 'development', got %q", appEnv)
	}

	if err := loadEnvFile(appEnv); err != nil {
		return nil, err
	}

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("config parse: %w", err)
	}

	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadEnvFile(appEnv string) error {
	filename := fmt.Sprintf(".env.%s", appEnv)

	err := godotenv.Load(filename)
	if err == nil {
		return nil
	}

	if appEnv == "production" && errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return fmt.Errorf("failed to load environment file %s: %w", filename, err)
}
