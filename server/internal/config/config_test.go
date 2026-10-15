package config

import (
	"testing"
)

func resetState() {
	cached = nil
	err = nil
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

		cfg, err := Get()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if cfg != nil {
			t.Fatalf("expected nil config, got %v", cfg)
		}
	})

	t.Run("valid configuration loads and caches instance", func(t *testing.T) {
		resetState()
		t.Setenv("APP_ENV", "production")
		t.Setenv("PORT", "3000")

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

		// Subsequent call returns cached instance even if environment changes
		t.Setenv("PORT", "9000")
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
	})
}
