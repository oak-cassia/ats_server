package config

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	port := 8080
	t.Setenv("PORT", fmt.Sprint(port))

	cfg, err := New()
	if err != nil {
		t.Fatalf("cannot create config: %v", err)
	}
	if cfg.Port != port {
		t.Errorf("expected port %d, got %d", port, cfg.Port)
	}
	env := "dev"
	if cfg.Env != env {
		t.Errorf("expected env %s, got %s", env, cfg.Env)
	}
}
