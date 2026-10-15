package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Port   int
	AppEnv string
}

var (
	once   sync.Once
	cached *Config
	err    error
)

// Get returns the singleton Config, loading it on first call.
// Subsequent calls return the same instance without re-reading the environment.
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

	return &Config{Port: port, AppEnv: appEnv}, nil
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
	raw := os.Getenv("PORT")

	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid PORT %q: not a number", raw)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid PORT %d: out of range 1-65535", port)
	}

	return port, nil
}
