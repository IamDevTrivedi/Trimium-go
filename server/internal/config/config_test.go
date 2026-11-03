package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func resetConfigState() {
	cached = nil
	loadErr = nil
	once = sync.Once{}
}

func validEnvValues(geoIPPath string) map[string]string {
	return map[string]string{
		"APP_ENV":               "development",
		"PORT":                  "8080",
		"BACKEND_URL_DEV":       "http://localhost:8080",
		"FRONTEND_URL_DEV":      "http://localhost:3000",
		"BACKEND_URL_PROD":      "https://api.example.com",
		"FRONTEND_URL_PROD":     "https://example.com",
		"SMTP_HOST":             "127.0.0.1",
		"SMTP_PORT":             "587",
		"SMTP_USERNAME":         "smtp-user",
		"SMTP_PASSWORD":         "smtp-password",
		"FROM_EMAIL":            "noreply@example.com",
		"GEOIP_DB_PATH":         geoIPPath,
		"MONGO_URI":             "mongodb://localhost:27017",
		"MONGO_DATABASE":        "trimium",
		"MONGO_MAX_POOL_SIZE":   "10",
		"MONGO_MIN_POOL_SIZE":   "2",
		"MONGO_CONNECT_TIMEOUT": "5s",
		"MONGO_PING_TIMEOUT":    "3s",
		"REDIS_HOST":            "127.0.0.1",
		"REDIS_PORT":            "6379",
		"REDIS_USERNAME":        "",
		"REDIS_PASSWORD":        "",
		"ASYNQ_CONCURRENCY":     "10",
		"REDIS_POOL_SIZE":       "10",
		"REDIS_MIN_IDLE_CONNS":  "2",
		"REDIS_DIAL_TIMEOUT":    "5s",
		"REDIS_READ_TIMEOUT":    "3s",
		"REDIS_WRITE_TIMEOUT":   "3s",
		"REDIS_PING_TIMEOUT":    "2s",
	}
}

func setValidEnv(t *testing.T) string {
	t.Helper()

	geoIPPath := filepath.Join(t.TempDir(), "GeoLite2-City.mmdb")
	if err := os.WriteFile(geoIPPath, []byte("test-geoip-db"), 0o600); err != nil {
		t.Fatalf("write geoip db: %v", err)
	}

	for key, value := range validEnvValues(geoIPPath) {
		t.Setenv(key, value)
	}

	return geoIPPath
}

func chdirToTemp(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	t.Chdir(dir)
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	oldValue, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}

	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, oldValue)
			return
		}
		_ = os.Unsetenv(key)
	})
}

func TestLoad_DevelopmentSuccess(t *testing.T) {
	resetConfigState()
	geoIPPath := setValidEnv(t)
	dir := chdirToTemp(t)
	writeFile(t, filepath.Join(dir, ".env.development"), "")

	cfg, err := load()
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, "development")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.GeoIPDBPath != geoIPPath {
		t.Errorf("GeoIPDBPath = %q, want %q", cfg.GeoIPDBPath, geoIPPath)
	}
	if cfg.MongoConnectTimeout != 5*time.Second {
		t.Errorf("MongoConnectTimeout = %v, want %v", cfg.MongoConnectTimeout, 5*time.Second)
	}
	if cfg.MongoPingTimeout != 3*time.Second {
		t.Errorf("MongoPingTimeout = %v, want %v", cfg.MongoPingTimeout, 3*time.Second)
	}
	if cfg.RedisReadTimeout != 3*time.Second {
		t.Errorf("RedisReadTimeout = %v, want %v", cfg.RedisReadTimeout, 3*time.Second)
	}
	if cfg.RedisWriteTimeout != 3*time.Second {
		t.Errorf("RedisWriteTimeout = %v, want %v", cfg.RedisWriteTimeout, 3*time.Second)
	}
	if cfg.RedisPingTimeout != 2*time.Second {
		t.Errorf("RedisPingTimeout = %v, want %v", cfg.RedisPingTimeout, 2*time.Second)
	}
}

func TestLoad_ProductionWithoutEnvFile(t *testing.T) {
	resetConfigState()
	setValidEnv(t)
	t.Setenv("APP_ENV", "production")
	chdirToTemp(t)

	cfg, err := load()
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}
	if cfg.AppEnv != "production" {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, "production")
	}
}

func TestLoad_ReadsValuesFromDevelopmentEnvFile(t *testing.T) {
	resetConfigState()

	geoIPPath := filepath.Join(t.TempDir(), "GeoLite2-City.mmdb")
	writeFile(t, geoIPPath, "test-geoip-db")

	values := validEnvValues(geoIPPath)
	values["APP_ENV"] = "development"

	dir := chdirToTemp(t)
	var envFile strings.Builder
	for key, value := range values {
		fmt.Fprintf(&envFile, "%s=%s\n", key, value)
	}
	writeFile(t, filepath.Join(dir, ".env.development"), envFile.String())

	t.Setenv("APP_ENV", "development")
	for key := range values {
		if key == "APP_ENV" {
			continue
		}
		unsetEnv(t, key)
	}

	cfg, err := load()
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.MongoDatabase != "trimium" {
		t.Errorf("MongoDatabase = %q, want %q", cfg.MongoDatabase, "trimium")
	}
	if cfg.GeoIPDBPath != geoIPPath {
		t.Errorf("GeoIPDBPath = %q, want %q", cfg.GeoIPDBPath, geoIPPath)
	}
}

func TestLoad_InvalidAppEnv(t *testing.T) {
	tests := []struct {
		name   string
		appEnv string
	}{
		{name: "empty", appEnv: ""},
		{name: "unknown", appEnv: "staging"},
		{name: "wrong case", appEnv: "Production"},
		{name: "short form", appEnv: "prod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigState()
			setValidEnv(t)
			t.Setenv("APP_ENV", tt.appEnv)
			chdirToTemp(t)

			cfg, err := load()
			if cfg != nil {
				t.Fatalf("load() cfg = %#v, want nil", cfg)
			}
			if err == nil {
				t.Fatal("load() error = nil, want error")
			}
			if !strings.Contains(err.Error(), "APP_ENV must be exactly 'production' or 'development'") {
				t.Errorf("load() error = %q, want APP_ENV validation error", err)
			}
		})
	}
}

func TestLoad_EnvFileErrors(t *testing.T) {
	t.Run("development file missing", func(t *testing.T) {
		resetConfigState()
		setValidEnv(t)
		t.Setenv("APP_ENV", "development")
		chdirToTemp(t)

		cfg, err := load()
		if cfg != nil {
			t.Fatalf("load() cfg = %#v, want nil", cfg)
		}
		if err == nil {
			t.Fatal("load() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "failed to load environment file .env.development") {
			t.Errorf("load() error = %q, want env file error", err)
		}
	})

	t.Run("production file malformed", func(t *testing.T) {
		resetConfigState()
		setValidEnv(t)
		t.Setenv("APP_ENV", "production")
		dir := chdirToTemp(t)
		writeFile(t, filepath.Join(dir, ".env.production"), "BAD$KEY=value\n")

		cfg, err := load()
		if cfg != nil {
			t.Fatalf("load() cfg = %#v, want nil", cfg)
		}
		if err == nil {
			t.Fatal("load() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "failed to load environment file .env.production") {
			t.Errorf("load() error = %q, want env file error", err)
		}
	})
}

func TestLoad_ParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "invalid int", key: "PORT", value: "not-a-number"},
		{name: "invalid uint", key: "MONGO_MAX_POOL_SIZE", value: "-1"},
		{name: "invalid duration", key: "MONGO_CONNECT_TIMEOUT", value: "soon"},
		{name: "empty required variable", key: "MONGO_DATABASE", value: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigState()
			setValidEnv(t)
			t.Setenv("APP_ENV", "production")
			chdirToTemp(t)
			t.Setenv(tt.key, tt.value)

			cfg, err := load()
			if cfg != nil {
				t.Fatalf("load() cfg = %#v, want nil", cfg)
			}
			if err == nil {
				t.Fatal("load() error = nil, want error")
			}
			if !strings.Contains(err.Error(), "config parse") {
				t.Errorf("load() error = %q, want parse error", err)
			}
		})
	}
}

func TestLoad_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantSub []string
	}{
		{name: "port too low", key: "PORT", value: "0", wantSub: []string{"Port", "min"}},
		{name: "port too high", key: "PORT", value: "65536", wantSub: []string{"Port", "max"}},
		{name: "invalid dev backend url", key: "BACKEND_URL_DEV", value: "not-a-url", wantSub: []string{"BackendURLDev", "http_url"}},
		{name: "invalid dev frontend url", key: "FRONTEND_URL_DEV", value: "ftp://example.com", wantSub: []string{"FrontendURLDev", "http_url"}},
		{name: "invalid prod backend url", key: "BACKEND_URL_PROD", value: "example.com", wantSub: []string{"BackendURLProd", "http_url"}},
		{name: "invalid smtp host", key: "SMTP_HOST", value: "not a host!", wantSub: []string{"SMTPHost"}},
		{name: "smtp port too low", key: "SMTP_PORT", value: "0", wantSub: []string{"SMTPPort", "min"}},
		{name: "invalid from email", key: "FROM_EMAIL", value: "not-an-email", wantSub: []string{"FromEmail", "email"}},
		{name: "geoip file missing", key: "GEOIP_DB_PATH", value: "/definitely/not/a/real/geoip.mmdb", wantSub: []string{"GeoIPDBPath", "file"}},
		{name: "invalid mongo uri", key: "MONGO_URI", value: "not a uri", wantSub: []string{"MongoURI", "uri"}},
		{name: "mongo max pool zero", key: "MONGO_MAX_POOL_SIZE", value: "0", wantSub: []string{"MongoMaxPoolSize", "min"}},
		{name: "mongo min pool above max", key: "MONGO_MIN_POOL_SIZE", value: "11", wantSub: []string{"MongoMinPoolSize", "ltefield"}},
		{name: "mongo connect timeout zero", key: "MONGO_CONNECT_TIMEOUT", value: "0s", wantSub: []string{"MongoConnectTimeout", "gt"}},
		{name: "mongo ping timeout negative", key: "MONGO_PING_TIMEOUT", value: "-1s", wantSub: []string{"MongoPingTimeout", "gt"}},
		{name: "redis host empty", key: "REDIS_HOST", value: "", wantSub: []string{"REDIS_HOST"}},
		{name: "redis port too low", key: "REDIS_PORT", value: "0", wantSub: []string{"RedisPort", "min"}},
		{name: "redis port too high", key: "REDIS_PORT", value: "65536", wantSub: []string{"RedisPort", "max"}},
		{name: "asynq concurrency zero", key: "ASYNQ_CONCURRENCY", value: "0", wantSub: []string{"AsynqConcurrency", "min"}},
		{name: "redis pool zero", key: "REDIS_POOL_SIZE", value: "0", wantSub: []string{"RedisPoolSize", "min"}},
		{name: "redis min idle above pool", key: "REDIS_MIN_IDLE_CONNS", value: "11", wantSub: []string{"RedisMinIdleConns", "ltefield"}},
		{name: "redis dial timeout zero", key: "REDIS_DIAL_TIMEOUT", value: "0s", wantSub: []string{"RedisDialTimeout", "gt"}},
		{name: "redis read timeout zero", key: "REDIS_READ_TIMEOUT", value: "0s", wantSub: []string{"RedisReadTimeout", "gt"}},
		{name: "redis write timeout zero", key: "REDIS_WRITE_TIMEOUT", value: "0s", wantSub: []string{"RedisWriteTimeout", "gt"}},
		{name: "redis ping timeout zero", key: "REDIS_PING_TIMEOUT", value: "0s", wantSub: []string{"RedisPingTimeout", "gt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigState()
			setValidEnv(t)
			t.Setenv("APP_ENV", "production")
			chdirToTemp(t)
			t.Setenv(tt.key, tt.value)

			cfg, err := load()
			if cfg != nil {
				t.Fatalf("load() cfg = %#v, want nil", cfg)
			}
			if err == nil {
				t.Fatal("load() error = nil, want error")
			}
			for _, want := range tt.wantSub {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("load() error = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}

func TestLoadEnvFile(t *testing.T) {
	t.Run("loads development file", func(t *testing.T) {
		const key = "CONFIG_TEST_DEVELOPMENT_VALUE"
		unsetEnv(t, key)

		dir := chdirToTemp(t)
		writeFile(t, filepath.Join(dir, ".env.development"), key+"=from-development\n")

		if err := loadEnvFile("development"); err != nil {
			t.Fatalf("loadEnvFile() error = %v", err)
		}
		if got := os.Getenv(key); got != "from-development" {
			t.Errorf("%s = %q, want %q", key, got, "from-development")
		}
	})

	t.Run("loads production file", func(t *testing.T) {
		const key = "CONFIG_TEST_PRODUCTION_VALUE"
		unsetEnv(t, key)

		dir := chdirToTemp(t)
		writeFile(t, filepath.Join(dir, ".env.production"), key+"=from-production\n")

		if err := loadEnvFile("production"); err != nil {
			t.Fatalf("loadEnvFile() error = %v", err)
		}
		if got := os.Getenv(key); got != "from-production" {
			t.Errorf("%s = %q, want %q", key, got, "from-production")
		}
	})

	t.Run("missing development file", func(t *testing.T) {
		chdirToTemp(t)

		err := loadEnvFile("development")
		if err == nil {
			t.Fatal("loadEnvFile() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "failed to load environment file .env.development") {
			t.Errorf("loadEnvFile() error = %q, want env file error", err)
		}
	})

	t.Run("missing production file is allowed", func(t *testing.T) {
		chdirToTemp(t)

		if err := loadEnvFile("production"); err != nil {
			t.Errorf("loadEnvFile() error = %v, want nil", err)
		}
	})

	t.Run("does not override existing environment", func(t *testing.T) {
		const key = "CONFIG_TEST_PRECEDENCE_VALUE"
		t.Setenv(key, "from-environment")

		dir := chdirToTemp(t)
		writeFile(t, filepath.Join(dir, ".env.development"), key+"=from-file\n")

		if err := loadEnvFile("development"); err != nil {
			t.Fatalf("loadEnvFile() error = %v", err)
		}
		if got := os.Getenv(key); got != "from-environment" {
			t.Errorf("%s = %q, want %q", key, got, "from-environment")
		}
	})

	t.Run("malformed production file", func(t *testing.T) {
		dir := chdirToTemp(t)
		writeFile(t, filepath.Join(dir, ".env.production"), "BAD$KEY=value\n")

		err := loadEnvFile("production")
		if err == nil {
			t.Fatal("loadEnvFile() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "failed to load environment file .env.production") {
			t.Errorf("loadEnvFile() error = %q, want env file error", err)
		}
	})
}

func TestGet_CachesConfig(t *testing.T) {
	resetConfigState()
	t.Cleanup(resetConfigState)

	setValidEnv(t)
	t.Setenv("APP_ENV", "production")
	chdirToTemp(t)

	first, err := Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if first == nil {
		t.Fatal("Get() config = nil, want config")
	}

	second, err := Get()
	if err != nil {
		t.Fatalf("second Get() error = %v", err)
	}
	if second != first {
		t.Errorf("Get() returned different pointers: %p and %p", first, second)
	}
}

func TestGet_CachesLoadError(t *testing.T) {
	resetConfigState()
	t.Cleanup(resetConfigState)

	setValidEnv(t)
	t.Setenv("APP_ENV", "staging")
	chdirToTemp(t)

	first, firstErr := Get()
	if first != nil {
		t.Fatalf("Get() config = %#v, want nil", first)
	}
	if firstErr == nil {
		t.Fatal("Get() error = nil, want error")
	}

	second, secondErr := Get()
	if second != nil {
		t.Fatalf("second Get() config = %#v, want nil", second)
	}
	if secondErr == nil {
		t.Fatal("second Get() error = nil, want error")
	}
	if secondErr.Error() != firstErr.Error() {
		t.Errorf("second Get() error = %q, want cached error %q", secondErr, firstErr)
	}
}
