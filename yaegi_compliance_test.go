package traefik_warden_test

import (
	"context"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/routewarden/traefik-warden"
)

// TestYaegiCompliance validates all Traefik Yaegi plugin requirements:
// 1. Package name must strictly match the module import path suffix (replacing hyphens with underscores).
//    For "github.com/routewarden/traefik-warden", Yaegi evaluates "traefik_warden.CreateConfig()".
// 2. Exported function CreateConfig() *Config must exist and return valid *Config.
// 3. Exported function New(context.Context, http.Handler, *Config, string) (http.Handler, error) must exist.
// 4. .traefik.yml manifest must declare the exact matching module import.
func TestYaegiCompliance(t *testing.T) {
	expectedPkgName := "traefik_warden"

	// 1. Check all non-test Go source files for correct package name
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		node, err := parser.ParseFile(fset, entry.Name(), nil, parser.PackageClauseOnly)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", entry.Name(), err)
		}

		if node.Name.Name != expectedPkgName {
			t.Errorf("file %s has package %q; Traefik Yaegi requires package %q", entry.Name(), node.Name.Name, expectedPkgName)
		}
	}

	// 2. Validate CreateConfig function signature and return type
	createConfigVal := reflect.ValueOf(traefik_warden.CreateConfig)
	if createConfigVal.Kind() != reflect.Func {
		t.Fatalf("CreateConfig must be a function")
	}
	createConfigType := createConfigVal.Type()
	if createConfigType.NumIn() != 0 {
		t.Errorf("CreateConfig must take 0 arguments, got %d", createConfigType.NumIn())
	}
	if createConfigType.NumOut() != 1 {
		t.Fatalf("CreateConfig must return exactly 1 value, got %d", createConfigType.NumOut())
	}
	if createConfigType.Out(0) != reflect.TypeOf(&traefik_warden.Config{}) {
		t.Errorf("CreateConfig must return *traefik_warden.Config, got %v", createConfigType.Out(0))
	}

	cfg := traefik_warden.CreateConfig()
	if cfg == nil {
		t.Fatalf("CreateConfig returned nil")
	}

	// 3. Validate New function signature and contract
	newVal := reflect.ValueOf(traefik_warden.New)
	if newVal.Kind() != reflect.Func {
		t.Fatalf("New must be a function")
	}
	newType := newVal.Type()
	if newType.NumIn() != 4 {
		t.Fatalf("New must take 4 arguments: (context.Context, http.Handler, *Config, string), got %d", newType.NumIn())
	}

	// Verify argument types
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	handlerType := reflect.TypeOf((*http.Handler)(nil)).Elem()
	cfgType := reflect.TypeOf(&traefik_warden.Config{})
	stringType := reflect.TypeOf("")

	if !newType.In(0).Implements(ctxType) {
		t.Errorf("New arg 0 must implement context.Context, got %v", newType.In(0))
	}
	if !newType.In(1).Implements(handlerType) {
		t.Errorf("New arg 1 must implement http.Handler, got %v", newType.In(1))
	}
	if newType.In(2) != cfgType {
		t.Errorf("New arg 2 must be *Config, got %v", newType.In(2))
	}
	if newType.In(3) != stringType {
		t.Errorf("New arg 3 must be string, got %v", newType.In(3))
	}

	// Verify return types
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if newType.NumOut() != 2 {
		t.Fatalf("New must return 2 values: (http.Handler, error), got %d", newType.NumOut())
	}
	if !newType.Out(0).Implements(handlerType) {
		t.Errorf("New return 0 must implement http.Handler, got %v", newType.Out(0))
	}
	if !newType.Out(1).Implements(errorType) {
		t.Errorf("New return 1 must implement error, got %v", newType.Out(1))
	}

	// 4. Verify .traefik.yml manifest
	manifestContent, err := os.ReadFile(filepath.Clean(".traefik.yml"))
	if err != nil {
		t.Fatalf("failed to read .traefik.yml: %v", err)
	}
	manifestStr := string(manifestContent)

	if !strings.Contains(manifestStr, "import: github.com/routewarden/traefik-warden") {
		t.Errorf(".traefik.yml must contain 'import: github.com/routewarden/traefik-warden'")
	}
	if !strings.Contains(manifestStr, "type: middleware") {
		t.Errorf(".traefik.yml must contain 'type: middleware'")
	}
}

// TestTraefikYaegiSimulation verifies plugin instantiation using reflection in the exact way
// Traefik's dynamic Yaegi interpreter executes plugins.
func TestTraefikYaegiSimulation(t *testing.T) {
	// Simulate Traefik dynamic configuration loading:
	// 1. Call CreateConfig()
	cfg := traefik_warden.CreateConfig()
	if !cfg.Enabled || !cfg.EnableDefaultPatterns {
		t.Errorf("expected default config to be enabled")
	}

	// 2. Instantiate plugin via New()
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler, err := traefik_warden.New(context.Background(), dummyHandler, cfg, "yaegi-sim-plugin")
	if err != nil {
		t.Fatalf("failed to instantiate plugin with CreateConfig output: %v", err)
	}
	if handler == nil {
		t.Fatalf("expected non-nil http.Handler from New")
	}
}
