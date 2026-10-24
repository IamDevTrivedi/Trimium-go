package config

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func resetState() {
	cached = nil
	err = nil
	once = sync.Once{}
}

func TestLoadAppEnv(t *testing.T) {
	tests := []struct {
		name      string
		envVal    string
		wantEnv   string
		expectErr bool
	}{
		{"valid development", "development", "development", false},
		{"valid production", "production", "production", false},
		{"valid test", "test", "test", false},
		{"invalid environment", "staging", "", true},
		{"empty environment", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tt.envVal)

			got, err := loadAppEnv()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantEnv {
				t.Errorf("got %q, want %q", got, tt.wantEnv)
			}
		})
	}
}

func TestLoadPort(t *testing.T) {
	tests := []struct {
		name      string
		portVal   string
		wantPort  int
		expectErr bool
	}{
		{"valid port 8080", "8080", 8080, false},
		{"valid port 1", "1", 1, false},
		{"valid port 65535", "65535", 65535, false},
		{"invalid port non-numeric", "abc", 0, true},
		{"invalid port out of range low", "0", 0, true},
		{"invalid port out of range negative", "-10", 0, true},
		{"invalid port out of range high", "65536", 0, true},
		{"empty port", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.portVal)

			got, err := loadPort()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantPort {
				t.Errorf("got %d, want %d", got, tt.wantPort)
			}
		})
	}
}

func TestLoadSMTPHost(t *testing.T) {
	tests := []struct {
		name      string
		hostVal   string
		wantHost  string
		expectErr bool
	}{
		{"valid localhost", "localhost", "localhost", false},
		{"valid domain", "smtp.example.com", "smtp.example.com", false},
		{"empty host", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SMTP_HOST", tt.hostVal)

			got, err := loadSMTPHost()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantHost {
				t.Errorf("got %q, want %q", got, tt.wantHost)
			}
		})
	}
}

func TestLoadSMTPPort(t *testing.T) {
	tests := []struct {
		name      string
		portVal   string
		wantPort  int
		expectErr bool
	}{
		{"valid port 1025", "1025", 1025, false},
		{"valid port 587", "587", 587, false},
		{"valid port 1", "1", 1, false},
		{"valid port 65535", "65535", 65535, false},
		{"invalid port non-numeric", "abc", 0, true},
		{"invalid port out of range low", "0", 0, true},
		{"invalid port out of range negative", "-10", 0, true},
		{"invalid port out of range high", "65536", 0, true},
		{"empty port", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SMTP_PORT", tt.portVal)

			got, err := loadSMTPPort()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantPort {
				t.Errorf("got %d, want %d", got, tt.wantPort)
			}
		})
	}
}

func TestLoadSMTPUsername(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"value set", "smtp-user", "smtp-user"},
		{"empty allowed", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SMTP_USERNAME", tt.value)

			if got := loadSMTPUsername(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadSMTPPassword(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"value set", "smtp-pass", "smtp-pass"},
		{"empty allowed", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SMTP_PASSWORD", tt.value)

			if got := loadSMTPPassword(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadFromEmail(t *testing.T) {
	tests := []struct {
		name      string
		emailVal  string
		wantEmail string
		expectErr bool
	}{
		{"valid email", "user@example.com", "user@example.com", false},
		{"valid formatted email", "Sender Name <user@example.com>", "user@example.com", false},
		{"invalid syntax plain text", "GEMINI is ASSHOLE", "", true},
		{"invalid missing at sign", "userexample.com", "", true},
		{"empty email", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FROM_EMAIL", tt.emailVal)

			got, err := loadFromEmail()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantEmail {
				t.Errorf("got %q, want %q", got, tt.wantEmail)
			}
		})
	}
}

func TestLoadBackendURL(t *testing.T) {
	tests := []struct {
		name      string
		appEnv    string
		devURL    string
		prodURL   string
		want      string
		expectErr bool
	}{
		{"development uses dev URL", "development", "http://localhost:5000", "https://api.example.com", "http://localhost:5000", false},
		{"production uses prod URL", "production", "http://localhost:5000", "https://api.example.com", "https://api.example.com", false},
		{"test uses prod URL", "test", "http://localhost:5000", "https://api.example.com", "https://api.example.com", false},
		{"development missing dev URL", "development", "", "https://api.example.com", "", true},
		{"production missing prod URL", "production", "http://localhost:5000", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("BACKEND_URL_DEV", tt.devURL)
			t.Setenv("BACKEND_URL_PROD", tt.prodURL)

			got, err := loadBackendURL(tt.appEnv)
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadFrontendURL(t *testing.T) {
	tests := []struct {
		name      string
		appEnv    string
		devURL    string
		prodURL   string
		want      string
		expectErr bool
	}{
		{"development uses dev URL", "development", "http://localhost:3000", "https://trimium.example.com", "http://localhost:3000", false},
		{"production uses prod URL", "production", "http://localhost:3000", "https://trimium.example.com", "https://trimium.example.com", false},
		{"test uses prod URL", "test", "http://localhost:3000", "https://trimium.example.com", "https://trimium.example.com", false},
		{"development missing dev URL", "development", "", "https://trimium.example.com", "", true},
		{"production missing prod URL", "production", "http://localhost:3000", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FRONTEND_URL_DEV", tt.devURL)
			t.Setenv("FRONTEND_URL_PROD", tt.prodURL)

			got, err := loadFrontendURL(tt.appEnv)
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadGeoIPDBPath(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      string
		expectErr bool
	}{
		{"valid path", "./data/geoip/GeoLite2-City.mmdb", "./data/geoip/GeoLite2-City.mmdb", false},
		{"empty path", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GEOIP_DB_PATH", tt.value)

			got, err := loadGeoIPDBPath()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadMongoURI(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      string
		expectErr bool
	}{
		{"valid uri", "mongodb://localhost:27017", "mongodb://localhost:27017", false},
		{"valid uri with auth", "mongodb://root:root@localhost:5002/?authSource=admin", "mongodb://root:root@localhost:5002/?authSource=admin", false},
		{"empty uri", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONGO_URI", tt.value)

			got, err := loadMongoURI()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadMongoDatabase(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      string
		expectErr bool
	}{
		{"valid database", "trimium", "trimium", false},
		{"empty database", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONGO_DB_NAME", tt.value)

			got, err := loadMongoDatabase()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadMongoMaxPoolSize(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      uint64
		expectErr bool
	}{
		{"missing", "", 0, true},
		{"valid size", "50", 50, false},
		{"zero means unlimited", "0", 0, false},
		{"invalid non-numeric", "abc", 0, true},
		{"invalid negative", "-1", 0, true},
		{"invalid fractional", "1.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONGO_MAX_POOL_SIZE", tt.value)

			got, err := loadMongoMaxPoolSize()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLoadMongoMinPoolSize(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      uint64
		expectErr bool
	}{
		{"missing", "", 0, true},
		{"valid size", "10", 10, false},
		{"invalid non-numeric", "abc", 0, true},
		{"invalid negative", "-1", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONGO_MIN_POOL_SIZE", tt.value)

			got, err := loadMongoMinPoolSize()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLoadMongoConnectTimeout(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      time.Duration
		expectErr bool
	}{
		{"missing", "", 0, true},
		{"valid seconds", "5s", 5 * time.Second, false},
		{"valid milliseconds", "1500ms", 1500 * time.Millisecond, false},
		{"invalid non-duration", "abc", 0, true},
		{"invalid zero", "0s", 0, true},
		{"invalid negative", "-1s", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONGO_CONNECT_TIMEOUT", tt.value)

			got, err := loadMongoConnectTimeout()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadMongoPingTimeout(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      time.Duration
		expectErr bool
	}{
		{"missing", "", 0, true},
		{"valid seconds", "3s", 3 * time.Second, false},
		{"invalid non-duration", "abc", 0, true},
		{"invalid zero", "0s", 0, true},
		{"invalid negative", "-1s", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONGO_PING_TIMEOUT", tt.value)

			got, err := loadMongoPingTimeout()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadRedisAddr(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      string
		expectErr bool
	}{
		{"valid address", "localhost:5001", "localhost:5001", false},
		{"valid hostname", "redis:6379", "redis:6379", false},
		{"empty address", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REDIS_ADDR", tt.value)

			got, err := loadRedisAddr()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadRedisUsername(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"value set", "redis-user", "redis-user"},
		{"empty allowed", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REDIS_USERNAME", tt.value)

			if got := loadRedisUsername(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadRedisPassword(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"value set", "redis-pass", "redis-pass"},
		{"empty allowed", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REDIS_PASSWORD", tt.value)

			if got := loadRedisPassword(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadRedisDB(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      int
		expectErr bool
	}{
		{"zero db", "0", 0, false},
		{"valid db", "2", 2, false},
		{"missing", "", 0, true},
		{"invalid non-numeric", "abc", 0, true},
		{"invalid negative", "-1", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REDIS_DB", tt.value)

			got, err := loadRedisDB()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLoadRedisPoolSize(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      int
		expectErr bool
	}{
		{"valid size", "10", 10, false},
		{"missing", "", 0, true},
		{"invalid zero", "0", 0, true},
		{"invalid negative", "-1", 0, true},
		{"invalid non-numeric", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REDIS_POOL_SIZE", tt.value)

			got, err := loadRedisPoolSize()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLoadRedisMinIdleConns(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      int
		expectErr bool
	}{
		{"zero allowed", "0", 0, false},
		{"valid count", "5", 5, false},
		{"missing", "", 0, true},
		{"invalid negative", "-1", 0, true},
		{"invalid non-numeric", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REDIS_MIN_IDLE_CONNS", tt.value)

			got, err := loadRedisMinIdleConns()
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLoadRedisTimeouts(t *testing.T) {
	loaders := []struct {
		name string
		key  string
		load func() (time.Duration, error)
	}{
		{"dial timeout", "REDIS_DIAL_TIMEOUT", loadRedisDialTimeout},
		{"read timeout", "REDIS_READ_TIMEOUT", loadRedisReadTimeout},
		{"write timeout", "REDIS_WRITE_TIMEOUT", loadRedisWriteTimeout},
		{"ping timeout", "REDIS_PING_TIMEOUT", loadRedisPingTimeout},
	}

	tests := []struct {
		name      string
		value     string
		want      time.Duration
		expectErr bool
	}{
		{"missing", "", 0, true},
		{"valid", "3s", 3 * time.Second, false},
		{"valid milliseconds", "1500ms", 1500 * time.Millisecond, false},
		{"invalid non-duration", "abc", 0, true},
		{"invalid zero", "0s", 0, true},
		{"invalid negative", "-1s", 0, true},
	}

	for _, loader := range loaders {
		t.Run(loader.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					t.Setenv(loader.key, tt.value)

					got, err := loader.load()
					if tt.expectErr && err == nil {
						t.Fatalf("expected error, got nil")
					}
					if !tt.expectErr && err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if got != tt.want {
						t.Errorf("got %v, want %v", got, tt.want)
					}
				})
			}
		})
	}
}

func TestLoadEnvFile(t *testing.T) {
	t.Run("production allows missing env file", func(t *testing.T) {
		err := loadEnvFile("production")
		if err != nil {
			t.Fatalf("unexpected error in production: %v", err)
		}
	})

	t.Run("missing env file for nonexistent env returns error", func(t *testing.T) {
		err := loadEnvFile("nonexistent_env")
		if err == nil {
			t.Fatalf("expected error for nonexistent file, got nil")
		}
	})
}

func TestGet(t *testing.T) {
	t.Cleanup(resetState)

	t.Run("invalid APP_ENV returns error", func(t *testing.T) {
		resetState()
		t.Setenv("APP_ENV", "invalid")
		t.Setenv("PORT", "8080")
		t.Setenv("SMTP_HOST", "localhost")
		t.Setenv("SMTP_PORT", "1025")

		cfg, err := Get()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if cfg != nil {
			t.Fatalf("expected nil config, got %v", cfg)
		}
	})

	t.Run("missing mongo settings returns error", func(t *testing.T) {
		resetState()
		t.Setenv("APP_ENV", "production")
		t.Setenv("PORT", "3000")
		t.Setenv("SMTP_HOST", "smtp.example.com")
		t.Setenv("SMTP_PORT", "587")
		t.Setenv("FROM_EMAIL", "noreply@example.com")
		t.Setenv("BACKEND_URL_PROD", "https://api.trimium.example.com")
		t.Setenv("FRONTEND_URL_PROD", "https://trimium.example.com")
		t.Setenv("GEOIP_DB_PATH", "./data/geoip/GeoLite2-City.mmdb")
		t.Setenv("MONGO_URI", "mongodb://localhost:27017")
		t.Setenv("MONGO_DB_NAME", "trimium")
		t.Setenv("MONGO_MAX_POOL_SIZE", "")
		t.Setenv("MONGO_MIN_POOL_SIZE", "")
		t.Setenv("MONGO_CONNECT_TIMEOUT", "")
		t.Setenv("MONGO_PING_TIMEOUT", "")

		cfg, err := Get()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if cfg != nil {
			t.Fatalf("expected nil config, got %v", cfg)
		}
		if !strings.Contains(err.Error(), "missing MONGO_MAX_POOL_SIZE") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing redis settings returns error", func(t *testing.T) {
		resetState()
		t.Setenv("APP_ENV", "production")
		t.Setenv("PORT", "3000")
		t.Setenv("SMTP_HOST", "smtp.example.com")
		t.Setenv("SMTP_PORT", "587")
		t.Setenv("FROM_EMAIL", "noreply@example.com")
		t.Setenv("BACKEND_URL_PROD", "https://api.trimium.example.com")
		t.Setenv("FRONTEND_URL_PROD", "https://trimium.example.com")
		t.Setenv("GEOIP_DB_PATH", "./data/geoip/GeoLite2-City.mmdb")
		t.Setenv("MONGO_URI", "mongodb://localhost:27017")
		t.Setenv("MONGO_DB_NAME", "trimium")
		t.Setenv("MONGO_MAX_POOL_SIZE", "100")
		t.Setenv("MONGO_MIN_POOL_SIZE", "0")
		t.Setenv("MONGO_CONNECT_TIMEOUT", "10s")
		t.Setenv("MONGO_PING_TIMEOUT", "10s")
		t.Setenv("REDIS_ADDR", "localhost:5001")
		t.Setenv("REDIS_USERNAME", "")
		t.Setenv("REDIS_PASSWORD", "")
		t.Setenv("REDIS_DB", "0")
		t.Setenv("REDIS_POOL_SIZE", "10")
		t.Setenv("REDIS_MIN_IDLE_CONNS", "0")
		t.Setenv("REDIS_DIAL_TIMEOUT", "5s")
		t.Setenv("REDIS_READ_TIMEOUT", "3s")
		t.Setenv("REDIS_WRITE_TIMEOUT", "3s")
		t.Setenv("REDIS_PING_TIMEOUT", "")

		cfg, err := Get()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if cfg != nil {
			t.Fatalf("expected nil config, got %v", cfg)
		}
		if !strings.Contains(err.Error(), "missing REDIS_PING_TIMEOUT") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid configuration loads and caches instance", func(t *testing.T) {
		resetState()
		t.Setenv("APP_ENV", "production")
		t.Setenv("PORT", "3000")
		t.Setenv("SMTP_HOST", "smtp.mailtrap.io")
		t.Setenv("SMTP_PORT", "2525")
		t.Setenv("SMTP_USERNAME", "testuser")
		t.Setenv("SMTP_PASSWORD", "testpass")
		t.Setenv("FROM_EMAIL", "noreply@example.com")
		t.Setenv("BACKEND_URL_PROD", "https://api.trimium.example.com")
		t.Setenv("FRONTEND_URL_PROD", "https://trimium.example.com")
		t.Setenv("GEOIP_DB_PATH", "./data/geoip/GeoLite2-City.mmdb")
		t.Setenv("MONGO_URI", "mongodb://localhost:27017")
		t.Setenv("MONGO_DB_NAME", "trimium")
		t.Setenv("MONGO_MAX_POOL_SIZE", "50")
		t.Setenv("MONGO_MIN_POOL_SIZE", "5")
		t.Setenv("MONGO_CONNECT_TIMEOUT", "3s")
		t.Setenv("MONGO_PING_TIMEOUT", "2s")
		t.Setenv("REDIS_ADDR", "localhost:5001")
		t.Setenv("REDIS_USERNAME", "redis-user")
		t.Setenv("REDIS_PASSWORD", "redis-pass")
		t.Setenv("REDIS_DB", "2")
		t.Setenv("REDIS_POOL_SIZE", "20")
		t.Setenv("REDIS_MIN_IDLE_CONNS", "5")
		t.Setenv("REDIS_DIAL_TIMEOUT", "4s")
		t.Setenv("REDIS_READ_TIMEOUT", "3s")
		t.Setenv("REDIS_WRITE_TIMEOUT", "2s")
		t.Setenv("REDIS_PING_TIMEOUT", "1s")

		cfg1, err := Get()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg1 == nil {
			t.Fatal("expected config instance, got nil")
		}
		if cfg1.Port != 3000 {
			t.Errorf("got port %d, want 3000", cfg1.Port)
		}
		if cfg1.AppEnv != "production" {
			t.Errorf("got appEnv %q, want production", cfg1.AppEnv)
		}
		if cfg1.SMTPHost != "smtp.mailtrap.io" {
			t.Errorf("got smtpHost %q, want smtp.mailtrap.io", cfg1.SMTPHost)
		}
		if cfg1.SMTPPort != 2525 {
			t.Errorf("got smtpPort %d, want 2525", cfg1.SMTPPort)
		}
		if cfg1.SMTPUsername != "testuser" {
			t.Errorf("got smtpUsername %q, want testuser", cfg1.SMTPUsername)
		}
		if cfg1.SMTPPassword != "testpass" {
			t.Errorf("got smtpPassword %q, want testpass", cfg1.SMTPPassword)
		}
		if cfg1.FromEmail != "noreply@example.com" {
			t.Errorf("got fromEmail %q, want noreply@example.com", cfg1.FromEmail)
		}
		if cfg1.BackendURL != "https://api.trimium.example.com" {
			t.Errorf("got backendURL %q, want https://api.trimium.example.com", cfg1.BackendURL)
		}
		if cfg1.FrontendURL != "https://trimium.example.com" {
			t.Errorf("got frontendURL %q, want https://trimium.example.com", cfg1.FrontendURL)
		}
		if cfg1.GeoIPDBPath != "./data/geoip/GeoLite2-City.mmdb" {
			t.Errorf("got geoIPDBPath %q, want ./data/geoip/GeoLite2-City.mmdb", cfg1.GeoIPDBPath)
		}
		if cfg1.MongoURI != "mongodb://localhost:27017" {
			t.Errorf("got mongoURI %q, want mongodb://localhost:27017", cfg1.MongoURI)
		}
		if cfg1.MongoDatabase != "trimium" {
			t.Errorf("got mongoDatabase %q, want trimium", cfg1.MongoDatabase)
		}
		if cfg1.MongoMaxPoolSize != 50 {
			t.Errorf("got mongoMaxPoolSize %d, want 50", cfg1.MongoMaxPoolSize)
		}
		if cfg1.MongoMinPoolSize != 5 {
			t.Errorf("got mongoMinPoolSize %d, want 5", cfg1.MongoMinPoolSize)
		}
		if cfg1.MongoConnectTimeout != 3*time.Second {
			t.Errorf("got mongoConnectTimeout %v, want 3s", cfg1.MongoConnectTimeout)
		}
		if cfg1.MongoPingTimeout != 2*time.Second {
			t.Errorf("got mongoPingTimeout %v, want 2s", cfg1.MongoPingTimeout)
		}
		if cfg1.RedisAddr != "localhost:5001" {
			t.Errorf("got redisAddr %q, want localhost:5001", cfg1.RedisAddr)
		}
		if cfg1.RedisUsername != "redis-user" {
			t.Errorf("got redisUsername %q, want redis-user", cfg1.RedisUsername)
		}
		if cfg1.RedisPassword != "redis-pass" {
			t.Errorf("got redisPassword %q, want redis-pass", cfg1.RedisPassword)
		}
		if cfg1.RedisDB != 2 {
			t.Errorf("got redisDB %d, want 2", cfg1.RedisDB)
		}
		if cfg1.RedisPoolSize != 20 {
			t.Errorf("got redisPoolSize %d, want 20", cfg1.RedisPoolSize)
		}
		if cfg1.RedisMinIdleConns != 5 {
			t.Errorf("got redisMinIdleConns %d, want 5", cfg1.RedisMinIdleConns)
		}
		if cfg1.RedisDialTimeout != 4*time.Second {
			t.Errorf("got redisDialTimeout %v, want 4s", cfg1.RedisDialTimeout)
		}
		if cfg1.RedisReadTimeout != 3*time.Second {
			t.Errorf("got redisReadTimeout %v, want 3s", cfg1.RedisReadTimeout)
		}
		if cfg1.RedisWriteTimeout != 2*time.Second {
			t.Errorf("got redisWriteTimeout %v, want 2s", cfg1.RedisWriteTimeout)
		}
		if cfg1.RedisPingTimeout != time.Second {
			t.Errorf("got redisPingTimeout %v, want 1s", cfg1.RedisPingTimeout)
		}

		// Subsequent call returns cached instance even if environment changes
		t.Setenv("PORT", "9000")
		t.Setenv("SMTP_PORT", "9999")
		cfg2, err := Get()
		if err != nil {
			t.Fatalf("unexpected error on second call: %v", err)
		}
		if cfg2 != cfg1 {
			t.Errorf("expected cached instance %p, got %p", cfg1, cfg2)
		}
		if cfg2.Port != 3000 {
			t.Errorf("expected cached port 3000, got %d", cfg2.Port)
		}
		if cfg2.SMTPPort != 2525 {
			t.Errorf("expected cached smtpPort 2525, got %d", cfg2.SMTPPort)
		}
	})

	t.Run("development environment selects dev URLs", func(t *testing.T) {
		resetState()
		if err := os.WriteFile(".env.development", []byte(""), 0600); err != nil {
			t.Fatalf("failed to create temp .env.development: %v", err)
		}
		t.Cleanup(func() {
			os.Remove(".env.development")
		})

		t.Setenv("APP_ENV", "development")
		t.Setenv("PORT", "5000")
		t.Setenv("SMTP_HOST", "localhost")
		t.Setenv("SMTP_PORT", "5003")
		t.Setenv("FROM_EMAIL", "trimium@localhost")
		t.Setenv("BACKEND_URL_DEV", "http://localhost:5000")
		t.Setenv("FRONTEND_URL_DEV", "http://localhost:3000")
		t.Setenv("BACKEND_URL_PROD", "https://api.trimium.trivedi.dev")
		t.Setenv("FRONTEND_URL_PROD", "https://trimium.vercel.app")
		t.Setenv("GEOIP_DB_PATH", "./data/geoip/GeoLite2-City.mmdb")
		t.Setenv("MONGO_URI", "mongodb://root:root@localhost:5002/?authSource=admin")
		t.Setenv("MONGO_DB_NAME", "trimium")
		t.Setenv("MONGO_MAX_POOL_SIZE", "100")
		t.Setenv("MONGO_MIN_POOL_SIZE", "0")
		t.Setenv("MONGO_CONNECT_TIMEOUT", "10s")
		t.Setenv("MONGO_PING_TIMEOUT", "10s")
		t.Setenv("REDIS_ADDR", "localhost:5001")
		t.Setenv("REDIS_USERNAME", "")
		t.Setenv("REDIS_PASSWORD", "")
		t.Setenv("REDIS_DB", "0")
		t.Setenv("REDIS_POOL_SIZE", "10")
		t.Setenv("REDIS_MIN_IDLE_CONNS", "0")
		t.Setenv("REDIS_DIAL_TIMEOUT", "5s")
		t.Setenv("REDIS_READ_TIMEOUT", "3s")
		t.Setenv("REDIS_WRITE_TIMEOUT", "3s")
		t.Setenv("REDIS_PING_TIMEOUT", "5s")

		cfg, err := Get()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.BackendURL != "http://localhost:5000" {
			t.Errorf("got backendURL %q, want http://localhost:5000", cfg.BackendURL)
		}
		if cfg.FrontendURL != "http://localhost:3000" {
			t.Errorf("got frontendURL %q, want http://localhost:3000", cfg.FrontendURL)
		}
	})

	t.Run("production environment selects prod URLs over dev URLs", func(t *testing.T) {
		resetState()
		t.Setenv("APP_ENV", "production")
		t.Setenv("PORT", "5000")
		t.Setenv("SMTP_HOST", "smtp.example.com")
		t.Setenv("SMTP_PORT", "587")
		t.Setenv("FROM_EMAIL", "admin@example.com")
		t.Setenv("BACKEND_URL_DEV", "http://localhost:5000")
		t.Setenv("FRONTEND_URL_DEV", "http://localhost:3000")
		t.Setenv("BACKEND_URL_PROD", "https://api.trimium.trivedi.dev")
		t.Setenv("FRONTEND_URL_PROD", "https://trimium.vercel.app")
		t.Setenv("GEOIP_DB_PATH", "./data/geoip/GeoLite2-City.mmdb")
		t.Setenv("MONGO_URI", "mongodb://root:root@localhost:5002/?authSource=admin")
		t.Setenv("MONGO_DB_NAME", "trimium")
		t.Setenv("MONGO_MAX_POOL_SIZE", "100")
		t.Setenv("MONGO_MIN_POOL_SIZE", "0")
		t.Setenv("MONGO_CONNECT_TIMEOUT", "10s")
		t.Setenv("MONGO_PING_TIMEOUT", "10s")
		t.Setenv("REDIS_ADDR", "localhost:5001")
		t.Setenv("REDIS_USERNAME", "")
		t.Setenv("REDIS_PASSWORD", "")
		t.Setenv("REDIS_DB", "0")
		t.Setenv("REDIS_POOL_SIZE", "10")
		t.Setenv("REDIS_MIN_IDLE_CONNS", "0")
		t.Setenv("REDIS_DIAL_TIMEOUT", "5s")
		t.Setenv("REDIS_READ_TIMEOUT", "3s")
		t.Setenv("REDIS_WRITE_TIMEOUT", "3s")
		t.Setenv("REDIS_PING_TIMEOUT", "5s")

		cfg, err := Get()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.BackendURL != "https://api.trimium.trivedi.dev" {
			t.Errorf("got backendURL %q, want https://api.trimium.trivedi.dev", cfg.BackendURL)
		}
		if cfg.FrontendURL != "https://trimium.vercel.app" {
			t.Errorf("got frontendURL %q, want https://trimium.vercel.app", cfg.FrontendURL)
		}
	})
}
