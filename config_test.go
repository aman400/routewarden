package routewarden_test

import (
	"net/http"
	"testing"

	"github.com/aman400/routewarden"
)

func TestCreateConfig_Defaults(t *testing.T) {
	cfg := routewarden.CreateConfig()

	if !cfg.Enabled {
		t.Errorf("expected Enabled to default to true")
	}
	if !cfg.EnableDefaultPatterns {
		t.Errorf("expected EnableDefaultPatterns to default to true")
	}
	if cfg.StatusCode != http.StatusForbidden {
		t.Errorf("expected StatusCode to default to %d, got %d", http.StatusForbidden, cfg.StatusCode)
	}
	if cfg.SilentDrop {
		t.Errorf("expected SilentDrop to default to false")
	}
	if cfg.CheckQuery {
		t.Errorf("expected CheckQuery to default to false")
	}
	if len(cfg.AllowPatterns) == 0 {
		t.Errorf("expected default AllowPatterns to not be empty")
	}
	if len(cfg.AllowedIPs) != 0 {
		t.Errorf("expected default AllowedIPs to be empty")
	}
	if cfg.Response != nil {
		t.Errorf("expected default Response to be nil so top-level configs are used cleanly")
	}
}

func TestDefaultBlockPatterns_ValidRegex(t *testing.T) {
	if len(routewarden.DefaultBlockPatterns) == 0 {
		t.Fatalf("DefaultBlockPatterns should not be empty")
	}
}

func TestDefaultAllowPatterns_ValidRegex(t *testing.T) {
	if len(routewarden.DefaultAllowPatterns) == 0 {
		t.Fatalf("DefaultAllowPatterns should not be empty")
	}
}
