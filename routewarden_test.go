package routewarden_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aman400/routewarden"
)

func TestRouteWarden_Defaults(t *testing.T) {
	cfg := routewarden.CreateConfig()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler, err := routewarden.New(context.Background(), next, cfg, "routewarden-test")
	if err != nil {
		t.Fatalf("unexpected error initializing plugin: %v", err)
	}

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		// Blocked sensitive files
		{"Block .env", "/.env", http.StatusForbidden},
		{"Block .env.production", "/.env.production", http.StatusForbidden},
		{"Block subpath .env", "/config/.env", http.StatusForbidden},
		{"Block app.log", "/app.log", http.StatusForbidden},
		{"Block database.sql", "/backup/database.sql", http.StatusForbidden},
		{"Block dump.bak", "/dump.bak", http.StatusForbidden},
		{"Block config.ini", "/settings/config.ini", http.StatusForbidden},
		{"Block config.yaml", "/config.yaml", http.StatusForbidden},
		{"Block config.yml", "/config.yml", http.StatusForbidden},
		{"Block secret.conf", "/secret.conf", http.StatusForbidden},
		{"Block notes.txt", "/notes.txt", http.StatusForbidden},
		{"Block .git folder", "/.git/config", http.StatusForbidden},
		{"Block .git root", "/.git", http.StatusForbidden},
		{"Block .aws folder", "/.aws/credentials", http.StatusForbidden},
		{"Block zip archive", "/backup.zip", http.StatusForbidden},
		{"Block tar archive", "/site.tar.gz", http.StatusForbidden},
		{"Block composer.lock", "/composer.lock", http.StatusForbidden},
		{"Block package-lock.json", "/package-lock.json", http.StatusForbidden},
		{"Block phpinfo", "/phpinfo.php", http.StatusForbidden},
		{"Block actuator", "/actuator/health", http.StatusForbidden},

		// URL-encoded evasion attempts
		{"Block encoded .env (%2eenv)", "/%2eenv", http.StatusForbidden},
		{"Block double encoded path traversal", "/static/%2e%2e/.env", http.StatusForbidden},

		// Allowed / legitimate files matching broad patterns
		{"Allow robots.txt", "/robots.txt", http.StatusOK},
		{"Allow ads.txt", "/ads.txt", http.StatusOK},
		{"Allow security.txt", "/security.txt", http.StatusOK},
		{"Allow .well-known", "/.well-known/acme-challenge/test", http.StatusOK},

		// Normal clean endpoints
		{"Allow normal root", "/", http.StatusOK},
		{"Allow normal API", "/api/v1/users", http.StatusOK},
		{"Allow normal page", "/dashboard", http.StatusOK},
		{"Allow normal image", "/assets/logo.png", http.StatusOK},
		{"Allow normal js", "/bundle.js", http.StatusOK},
		{"Allow normal css", "/styles.css", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedCode {
				t.Errorf("Path %q: expected status %d, got %d", tc.url, tc.expectedCode, rr.Code)
			}
		})
	}
}

func TestRouteWarden_CustomBlockPatterns(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.EnableDefaultPatterns = false
	// Add user's exact requested pattern
	cfg.BlockPatterns = []string{
		`(?i)(^|/)(\.env.*|.*\.(txt|log|bak|backup|sql|conf|config|ini|yaml|yml))`,
	}
	cfg.StatusCode = http.StatusNotFound // Custom 404
	cfg.CustomResponseText = "Not Found"

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler, err := routewarden.New(context.Background(), next, cfg, "custom-test")
	if err != nil {
		t.Fatalf("unexpected error initializing plugin: %v", err)
	}

	tests := []struct {
		url          string
		expectedCode int
	}{
		{"/.env", http.StatusNotFound},
		{"/.env.local", http.StatusNotFound},
		{"/app.log", http.StatusNotFound},
		{"/db.backup", http.StatusNotFound},
		{"/my.sql", http.StatusNotFound},
		{"/app.conf", http.StatusNotFound},
		{"/test.ini", http.StatusNotFound},
		{"/server.yaml", http.StatusNotFound},
		{"/server.yml", http.StatusNotFound},
		{"/secret.txt", http.StatusNotFound},
		// Unblocked paths since default patterns were disabled
		{"/.git/config", http.StatusOK},
		{"/normal/page", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedCode {
				t.Errorf("Path %q: expected status %d, got %d", tc.url, tc.expectedCode, rr.Code)
			}
		})
	}
}

func TestRouteWarden_JSONResponse(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.PathPatterns = []string{`^/api/admin/.*`}
	cfg.Response = &routewarden.ResponseConfig{
		Mode:       "json",
		StatusCode: http.StatusTeapot, // 418 or 403 / 429
		Body:       `{"error":"unauthorized_resource","status":418,"success":false}`,
		Headers: map[string]string{
			"X-RouteWarden-Blocked": "true",
		},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler, err := routewarden.New(context.Background(), next, cfg, "json-test")
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/secrets", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Errorf("expected status %d, got %d", http.StatusTeapot, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}

	if rr.Header().Get("X-RouteWarden-Blocked") != "true" {
		t.Errorf("expected custom header X-RouteWarden-Blocked to be 'true'")
	}

	body := rr.Body.String()
	if !strings.Contains(body, `"error":"unauthorized_resource"`) {
		t.Errorf("unexpected body content: %s", body)
	}
}

func TestRouteWarden_HTMLResponse(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.PathPatterns = []string{`(?i)^/admin/login`}
	cfg.Response = &routewarden.ResponseConfig{
		Mode:       "html",
		StatusCode: http.StatusForbidden,
		Body:       `<!DOCTYPE html><html><body><h1>Access Denied</h1><p>Restricted area.</p></body></html>`,
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler, err := routewarden.New(context.Background(), next, cfg, "html-test")
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %q", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, `<h1>Access Denied</h1>`) {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestRouteWarden_CaptchaResponse(t *testing.T) {
	providers := []struct {
		provider string
		element  string
	}{
		{"turnstile", "cf-turnstile"},
		{"hcaptcha", "h-captcha"},
		{"recaptcha", "g-recaptcha"},
	}

	for _, p := range providers {
		t.Run(p.provider, func(t *testing.T) {
			cfg := routewarden.CreateConfig()
			cfg.PathPatterns = []string{`^/login`}
			cfg.Response = &routewarden.ResponseConfig{
				Mode:       "captcha",
				StatusCode: http.StatusForbidden,
				Captcha: &routewarden.CaptchaConfig{
					Provider: p.provider,
					SiteKey:  "0x4AAAAAAtestkey123",
					Title:    "Custom Security Check",
				},
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handler, err := routewarden.New(context.Background(), next, cfg, "captcha-test")
			if err != nil {
				t.Fatalf("failed to create handler: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/login", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("expected status 403, got %d", rr.Code)
			}

			body := rr.Body.String()
			if !strings.Contains(body, p.element) {
				t.Errorf("expected captcha container with %q, got body:\n%s", p.element, body)
			}
			if !strings.Contains(body, "0x4AAAAAAtestkey123") {
				t.Errorf("expected sitekey to be in HTML output")
			}
			if !strings.Contains(body, "Custom Security Check") {
				t.Errorf("expected title to be in HTML output")
			}
		})
	}
}

func TestRouteWarden_RedirectResponse(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.PathPatterns = []string{`^/trap`}
	cfg.Response = &routewarden.ResponseConfig{
		Mode:        "redirect",
		StatusCode:  http.StatusTemporaryRedirect, // 307
		RedirectURL: "https://example.com/blocked",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler, err := routewarden.New(context.Background(), next, cfg, "redirect-test")
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/trap", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status 307, got %d", rr.Code)
	}

	loc := rr.Header().Get("Location")
	if loc != "https://example.com/blocked" {
		t.Errorf("expected redirect location https://example.com/blocked, got %q", loc)
	}
}

func TestRouteWarden_AllowPatternsOverride(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.AllowPatterns = append(cfg.AllowPatterns, `(?i)^/public/.*\.txt$`)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler, err := routewarden.New(context.Background(), next, cfg, "allow-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// /public/info.txt would normally match .txt block rule, but is explicitly allowed
	req1 := httptest.NewRequest(http.MethodGet, "/public/info.txt", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("expected 200 for allowed pattern, got %d", rr1.Code)
	}

	// /private/info.txt should be blocked
	req2 := httptest.NewRequest(http.MethodGet, "/private/info.txt", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Errorf("expected 403 for blocked pattern, got %d", rr2.Code)
	}
}

func TestRouteWarden_CheckQuery(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.CheckQuery = true

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler, err := routewarden.New(context.Background(), next, cfg, "query-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/search?file=.env", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 when query matches sensitive block pattern, got %d", rr.Code)
	}
}

func TestRouteWarden_Disabled(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.Enabled = false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler, err := routewarden.New(context.Background(), next, cfg, "disabled-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 when plugin is disabled, got %d", rr.Code)
	}
}

func TestRouteWarden_InvalidRegex(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.BlockPatterns = []string{"(unclosed parenthesis"}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	_, err := routewarden.New(context.Background(), next, cfg, "error-test")
	if err == nil {
		t.Errorf("expected error for invalid regex pattern, got nil")
	}
}

func TestRouteWarden_SecurityEvasionVectors(t *testing.T) {
	cfg := routewarden.CreateConfig()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("SHOULD NOT BE REACHED"))
	})

	handler, err := routewarden.New(context.Background(), next, cfg, "security-evasion-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	evasionTests := []struct {
		name string
		path string
	}{
		// Double URL encoding
		{"Double encoded dot (%252eenv)", "/%252eenv"},
		{"Double encoded traversal (%252e%252e/.env)", "/static/%252e%252e/.env"},

		// Matrix parameter evasion (Spring / Java / reverse-proxy semicolon bypass)
		{"Semicolon matrix parameter prefix", "/;.env"},
		{"Semicolon path segment suffix", "/static;param=123/.env"},
		{"Semicolon within path", "/api;.env/config.json"},

		// Backslash path separation (Windows / IIS / reverse-proxy normalizer evasion)
		{"Backslash traversal", "/static\\..\\.env"},
		{"Direct backslash path", "/\\.env"},

		// Encoded null byte injection attempt
		{"Null byte in path (%00)", "/.env%00.png"},

		// Direct and subpath variations
		{"Dot slash path variation", "/./.env"},
		{"Trailing slash directory .git", "/.git/"},
		{"Case insensitivity (.ENV)", "/.ENV"},
		{"Case insensitivity (.YML)", "/config.YML"},
		{"Case insensitivity (.SQL)", "/DUMP.SQL"},
	}

	for _, tc := range evasionTests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("Security evasion test %q failed: path %q got status %d, expected %d",
					tc.name, tc.path, rr.Code, http.StatusForbidden)
			}
		})
	}
}

func TestRouteWarden_SilentDrop(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.SilentDrop = true
	cfg.StatusCode = http.StatusForbidden

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler, err := routewarden.New(context.Background(), next, cfg, "silent-drop-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Since httptest.ResponseRecorder doesn't implement http.Hijacker, it falls back to writing status code with empty body
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
	if rr.Body.Len() > 0 {
		t.Errorf("expected empty body for silent drop fallback, got %q", rr.Body.String())
	}
}

func TestRouteWarden_IPWhitelist(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.AllowedIPs = []string{
		"192.168.1.50",       // Exact IP
		"10.0.0.0/24",        // CIDR subnet
		"2001:db8::/32",      // IPv6 subnet
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ALLOWED"))
	})

	handler, err := routewarden.New(context.Background(), next, cfg, "ip-whitelist-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name         string
		remoteAddr   string
		xff          string
		xrip         string
		path         string
		expectedCode int
	}{
		{
			name:         "Whitelisted exact IP in RemoteAddr",
			remoteAddr:   "192.168.1.50:54321",
			path:         "/.env",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Whitelisted CIDR IP in RemoteAddr",
			remoteAddr:   "10.0.0.42:12345",
			path:         "/.env",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Whitelisted IP via X-Forwarded-For",
			remoteAddr:   "203.0.113.195:8080",
			xff:          "192.168.1.50, 10.0.0.1",
			path:         "/.env.production",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Whitelisted IP via X-Real-IP",
			remoteAddr:   "203.0.113.195:8080",
			xrip:         "10.0.0.99",
			path:         "/backup/database.sql",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Non-whitelisted IP accessing sensitive file is blocked",
			remoteAddr:   "203.0.113.5:12345",
			path:         "/.env",
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "Non-whitelisted IP accessing normal page is allowed",
			remoteAddr:   "203.0.113.5:12345",
			path:         "/api/users",
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.RemoteAddr = tc.remoteAddr
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}
			if tc.xrip != "" {
				req.Header.Set("X-Real-IP", tc.xrip)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedCode {
				t.Errorf("expected code %d, got %d", tc.expectedCode, rr.Code)
			}
		})
	}
}

func TestRouteWarden_InvalidAllowedIPs(t *testing.T) {
	cfg := routewarden.CreateConfig()
	cfg.AllowedIPs = []string{"not-an-ip-or-cidr"}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	_, err := routewarden.New(context.Background(), next, cfg, "invalid-ip-test")
	if err == nil {
		t.Errorf("expected error on invalid IP format, got nil")
	}

	cfg2 := routewarden.CreateConfig()
	cfg2.AllowedIPs = []string{"10.0.0.0/99"} // invalid CIDR mask
	_, err2 := routewarden.New(context.Background(), next, cfg2, "invalid-cidr-test")
	if err2 == nil {
		t.Errorf("expected error on invalid CIDR mask, got nil")
	}
}

