package config_test

import (
	"testing"

	"github.com/andreypisarev/secret-santa/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("ENV", "")
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("BASE_URL", "")
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("EMAIL_FROM", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
	if !cfg.IsDev() {
		t.Error("IsDev() = false, want true")
	}
	if cfg.DatabasePath != "app.db" {
		t.Errorf("DatabasePath = %q, want %q", cfg.DatabasePath, "app.db")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("ENV", "production")
	t.Setenv("DATABASE_PATH", "/data/app.db")
	t.Setenv("BASE_URL", "https://example.com")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 3000 {
		t.Errorf("Port = %d, want 3000", cfg.Port)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want %q", cfg.Env, "production")
	}
	if cfg.IsDev() {
		t.Error("IsDev() = true, want false")
	}
	if cfg.DatabasePath != "/data/app.db" {
		t.Errorf("DatabasePath = %q, want %q", cfg.DatabasePath, "/data/app.db")
	}
	if cfg.BaseURL != "https://example.com" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://example.com")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("PORT", "abc")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid PORT")
	}
}

func TestLoad_PortOutOfRange(t *testing.T) {
	t.Setenv("PORT", "99999")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for out-of-range PORT")
	}
}
