package spindle

import "testing"

func TestConfigDefault(t *testing.T) {
	t.Parallel()

	cfg := configDefault()
	if cfg.PageKey != "page" {
		t.Errorf("PageKey = %q, want %q", cfg.PageKey, "page")
	}
	if cfg.DefaultPage != 1 {
		t.Errorf("DefaultPage = %d, want %d", cfg.DefaultPage, 1)
	}
	if cfg.LimitKey != "limit" {
		t.Errorf("LimitKey = %q, want %q", cfg.LimitKey, "limit")
	}
	if cfg.DefaultLimit != 10 {
		t.Errorf("DefaultLimit = %d, want %d", cfg.DefaultLimit, 10)
	}
}

func TestConfigOverride(t *testing.T) {
	t.Parallel()

	cfg := configDefault(Config{
		PageKey:      "p",
		LimitKey:     "l",
		DefaultPage:  5,
		DefaultLimit: 50,
	})
	if cfg.PageKey != "p" {
		t.Errorf("PageKey = %q, want %q", cfg.PageKey, "p")
	}
	if cfg.LimitKey != "l" {
		t.Errorf("LimitKey = %q, want %q", cfg.LimitKey, "l")
	}
	if cfg.DefaultPage != 5 {
		t.Errorf("DefaultPage = %d, want %d", cfg.DefaultPage, 5)
	}
	if cfg.DefaultLimit != 50 {
		t.Errorf("DefaultLimit = %d, want %d", cfg.DefaultLimit, 50)
	}
}

func TestConfigNegativeDefaults(t *testing.T) {
	t.Parallel()

	cfg := configDefault(Config{
		DefaultPage:  -1,
		DefaultLimit: -1,
	})
	if cfg.DefaultPage != 1 {
		t.Errorf("DefaultPage = %d, want %d", cfg.DefaultPage, 1)
	}
	if cfg.DefaultLimit != 10 {
		t.Errorf("DefaultLimit = %d, want %d", cfg.DefaultLimit, 10)
	}
}
