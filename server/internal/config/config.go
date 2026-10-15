package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         int
	AppEnv       string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
}

var (
	cached *Config
	err    error
)

// Get returns the singleton Config, loading it on first call.
// Subsequent calls return the same instance without re-reading the environment.
func Get() (*Config, error) {
	if cached == nil && err == nil {
		cached, err = load()
	}
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

	return &Config{
		Port:         port,
		AppEnv:       appEnv,
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
	}, nil
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

func loadSMTPHost() (string, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return "", errors.New("missing SMTP_HOST")
	}
	return host, nil
}

func loadSMTPPort() (int, error) {
	raw := os.Getenv("SMTP_PORT")

	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid SMTP_PORT %q: not a number", raw)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid SMTP_PORT %d: out of range 1-65535", port)
	}

	return port, nil
}
