package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net/mail"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         int
	AppEnv       string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	BackendURL   string
	FrontendURL  string
	GeoIPDBPath  string
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

	smtpUsername, err := loadSMTPUsername()
	if err != nil {
		return nil, err
	}

	smtpPassword, err := loadSMTPPassword()
	if err != nil {
		return nil, err
	}

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

	return &Config{
		Port:         port,
		AppEnv:       appEnv,
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUsername: smtpUsername,
		SMTPPassword: smtpPassword,
		FromEmail:    fromEmail,
		BackendURL:   backendURL,
		FrontendURL:  frontendURL,
		GeoIPDBPath:  geoIPDBPath,
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

func loadSMTPUsername() (string, error) {
	return os.Getenv("SMTP_USERNAME"), nil
}

func loadSMTPPassword() (string, error) {
	return os.Getenv("SMTP_PASSWORD"), nil
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
