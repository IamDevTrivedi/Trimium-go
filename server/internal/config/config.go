package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net/mail"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                int
	AppEnv              string
	SMTPHost            string
	SMTPPort            int
	SMTPUsername        string
	SMTPPassword        string
	FromEmail           string
	BackendURL          string
	FrontendURL         string
	GeoIPDBPath         string
	MongoURI            string
	MongoDatabase       string
	MongoMaxPoolSize    uint64
	MongoMinPoolSize    uint64
	MongoConnectTimeout time.Duration
	MongoPingTimeout    time.Duration
	RedisAddr           string
	RedisUsername       string
	RedisPassword       string
	RedisDB             int
	RedisPoolSize       int
	RedisMinIdleConns   int
	RedisDialTimeout    time.Duration
	RedisReadTimeout    time.Duration
	RedisWriteTimeout   time.Duration
	RedisPingTimeout    time.Duration
}

var (
	cached *Config
	err    error
	once   sync.Once
)

func Get() (*Config, error) {
	once.Do(func() {
		cached, err = load()
	})
	return cached, err
}

func load() (*Config, error) {
	appEnv, err := loadAppEnv()
	if err != nil {
		return nil, err
	}

	if err := loadEnvFile(appEnv); err != nil {
		return nil, err
	}

	port, err := loadPort()
	if err != nil {
		return nil, err
	}

	smtpHost, err := loadSMTPHost()
	if err != nil {
		return nil, err
	}

	smtpPort, err := loadSMTPPort()
	if err != nil {
		return nil, err
	}

	smtpUsername := loadSMTPUsername()
	smtpPassword := loadSMTPPassword()

	fromEmail, err := loadFromEmail()
	if err != nil {
		return nil, err
	}

	backendURL, err := loadBackendURL(appEnv)
	if err != nil {
		return nil, err
	}

	frontendURL, err := loadFrontendURL(appEnv)
	if err != nil {
		return nil, err
	}

	geoIPDBPath, err := loadGeoIPDBPath()
	if err != nil {
		return nil, err
	}

	mongoURI, err := loadMongoURI()
	if err != nil {
		return nil, err
	}

	mongoDatabase, err := loadMongoDatabase()
	if err != nil {
		return nil, err
	}

	mongoMaxPoolSize, err := loadMongoMaxPoolSize()
	if err != nil {
		return nil, err
	}

	mongoMinPoolSize, err := loadMongoMinPoolSize()
	if err != nil {
		return nil, err
	}

	mongoConnectTimeout, err := loadMongoConnectTimeout()
	if err != nil {
		return nil, err
	}

	mongoPingTimeout, err := loadMongoPingTimeout()
	if err != nil {
		return nil, err
	}

	redisAddr, err := loadRedisAddr()
	if err != nil {
		return nil, err
	}

	redisUsername := loadRedisUsername()
	redisPassword := loadRedisPassword()

	redisDB, err := loadRedisDB()
	if err != nil {
		return nil, err
	}

	redisPoolSize, err := loadRedisPoolSize()
	if err != nil {
		return nil, err
	}

	redisMinIdleConns, err := loadRedisMinIdleConns()
	if err != nil {
		return nil, err
	}

	redisDialTimeout, err := loadRedisDialTimeout()
	if err != nil {
		return nil, err
	}

	redisReadTimeout, err := loadRedisReadTimeout()
	if err != nil {
		return nil, err
	}

	redisWriteTimeout, err := loadRedisWriteTimeout()
	if err != nil {
		return nil, err
	}

	redisPingTimeout, err := loadRedisPingTimeout()
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:                port,
		AppEnv:              appEnv,
		SMTPHost:            smtpHost,
		SMTPPort:            smtpPort,
		SMTPUsername:        smtpUsername,
		SMTPPassword:        smtpPassword,
		FromEmail:           fromEmail,
		BackendURL:          backendURL,
		FrontendURL:         frontendURL,
		GeoIPDBPath:         geoIPDBPath,
		MongoURI:            mongoURI,
		MongoDatabase:       mongoDatabase,
		MongoMaxPoolSize:    mongoMaxPoolSize,
		MongoMinPoolSize:    mongoMinPoolSize,
		MongoConnectTimeout: mongoConnectTimeout,
		MongoPingTimeout:    mongoPingTimeout,
		RedisAddr:           redisAddr,
		RedisUsername:       redisUsername,
		RedisPassword:       redisPassword,
		RedisDB:             redisDB,
		RedisPoolSize:       redisPoolSize,
		RedisMinIdleConns:   redisMinIdleConns,
		RedisDialTimeout:    redisDialTimeout,
		RedisReadTimeout:    redisReadTimeout,
		RedisWriteTimeout:   redisWriteTimeout,
		RedisPingTimeout:    redisPingTimeout,
	}, nil
}

func requireString(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("missing %s", key)
	}
	return val, nil
}

func requirePort(key string) (int, error) {
	raw := os.Getenv(key)

	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: not a number", key, raw)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid %s %d: out of range 1-65535", key, port)
	}

	return port, nil
}

func requireUint64(key string) (uint64, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, fmt.Errorf("missing %s", key)
	}

	val, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: not a non-negative integer", key, raw)
	}

	return val, nil
}

func requireDuration(key string) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, fmt.Errorf("missing %s", key)
	}

	val, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: not a duration", key, raw)
	}
	if val <= 0 {
		return 0, fmt.Errorf("invalid %s %q: must be greater than zero", key, raw)
	}

	return val, nil
}

func requireNonNegativeInt(key string) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, fmt.Errorf("missing %s", key)
	}

	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: not an integer", key, raw)
	}
	if val < 0 {
		return 0, fmt.Errorf("invalid %s %d: must not be negative", key, val)
	}

	return val, nil
}

func requirePositiveInt(key string) (int, error) {
	val, err := requireNonNegativeInt(key)
	if err != nil {
		return 0, err
	}
	if val == 0 {
		return 0, fmt.Errorf("invalid %s: must be greater than zero", key)
	}

	return val, nil
}

func loadAppEnv() (string, error) {
	env := os.Getenv("APP_ENV")
	switch env {
	case "development", "production", "test":
		return env, nil
	default:
		return "", fmt.Errorf("invalid APP_ENV %q: want development, production or test", env)
	}
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

	return fmt.Errorf("load %s: %w", filename, err)
}

func loadPort() (int, error) {
	return requirePort("PORT")
}

func loadSMTPHost() (string, error) {
	return requireString("SMTP_HOST")
}

func loadSMTPPort() (int, error) {
	return requirePort("SMTP_PORT")
}

func loadSMTPUsername() string {
	return os.Getenv("SMTP_USERNAME")
}

func loadSMTPPassword() string {
	return os.Getenv("SMTP_PASSWORD")
}

func loadFromEmail() (string, error) {
	raw, err := requireString("FROM_EMAIL")
	if err != nil {
		return "", err
	}

	addr, err := mail.ParseAddress(raw)
	if err != nil {
		return "", fmt.Errorf("invalid FROM_EMAIL %q: %w", raw, err)
	}

	return addr.Address, nil
}

func loadBackendURL(appEnv string) (string, error) {
	key := "BACKEND_URL_PROD"
	if appEnv == "development" {
		key = "BACKEND_URL_DEV"
	}
	return requireString(key)
}

func loadFrontendURL(appEnv string) (string, error) {
	key := "FRONTEND_URL_PROD"
	if appEnv == "development" {
		key = "FRONTEND_URL_DEV"
	}
	return requireString(key)
}

func loadGeoIPDBPath() (string, error) {
	return requireString("GEOIP_DB_PATH")
}

func loadMongoURI() (string, error) {
	return requireString("MONGO_URI")
}

func loadMongoDatabase() (string, error) {
	return requireString("MONGO_DB_NAME")
}

func loadMongoMaxPoolSize() (uint64, error) {
	return requireUint64("MONGO_MAX_POOL_SIZE")
}

func loadMongoMinPoolSize() (uint64, error) {
	return requireUint64("MONGO_MIN_POOL_SIZE")
}

func loadMongoConnectTimeout() (time.Duration, error) {
	return requireDuration("MONGO_CONNECT_TIMEOUT")
}

func loadMongoPingTimeout() (time.Duration, error) {
	return requireDuration("MONGO_PING_TIMEOUT")
}

func loadRedisAddr() (string, error) {
	return requireString("REDIS_ADDR")
}

func loadRedisUsername() string {
	return os.Getenv("REDIS_USERNAME")
}

func loadRedisPassword() string {
	return os.Getenv("REDIS_PASSWORD")
}

func loadRedisDB() (int, error) {
	return requireNonNegativeInt("REDIS_DB")
}

func loadRedisPoolSize() (int, error) {
	return requirePositiveInt("REDIS_POOL_SIZE")
}

func loadRedisMinIdleConns() (int, error) {
	return requireNonNegativeInt("REDIS_MIN_IDLE_CONNS")
}

func loadRedisDialTimeout() (time.Duration, error) {
	return requireDuration("REDIS_DIAL_TIMEOUT")
}

func loadRedisReadTimeout() (time.Duration, error) {
	return requireDuration("REDIS_READ_TIMEOUT")
}

func loadRedisWriteTimeout() (time.Duration, error) {
	return requireDuration("REDIS_WRITE_TIMEOUT")
}

func loadRedisPingTimeout() (time.Duration, error) {
	return requireDuration("REDIS_PING_TIMEOUT")
}
