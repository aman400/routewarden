package routewarden_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aman400/routewarden"
)

func TestResponseHandler_JSON(t *testing.T) {
	cfg := &routewarden.ResponseConfig{
		Mode:       "json",
		StatusCode: http.StatusTeapot,
		Body:       `{"error":"blocked","code":418}`,
		Headers: map[string]string{
			"X-Custom-Header": "WardenSec",
		},
	}

	handler, err := routewarden.NewResponseHandler(cfg, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Errorf("expected %d, got %d", http.StatusTeapot, rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
		t.Errorf("expected application/json content-type")
	}
	if rr.Header().Get("X-Custom-Header") != "WardenSec" {
		t.Errorf("expected custom header")
	}
	if !strings.Contains(rr.Body.String(), `"error":"blocked"`) {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}

func TestResponseHandler_HTML(t *testing.T) {
	cfg := &routewarden.ResponseConfig{
		Mode:       "html",
		StatusCode: http.StatusForbidden,
		Body:       "<html><body>Access Restricted</body></html>",
	}

	handler, err := routewarden.NewResponseHandler(cfg, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected %d, got %d", http.StatusForbidden, rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "text/html") {
		t.Errorf("expected text/html content-type")
	}
	if !strings.Contains(rr.Body.String(), "Access Restricted") {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}

func TestResponseHandler_Captcha(t *testing.T) {
	cfg := &routewarden.ResponseConfig{
		Mode:       "captcha",
		StatusCode: http.StatusForbidden,
		Captcha: &routewarden.CaptchaConfig{
			Provider: "turnstile",
			SiteKey:  "0x4AAAAAAtestkey",
			Title:    "Bot Check",
		},
	}

	handler, err := routewarden.NewResponseHandler(cfg, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected %d, got %d", http.StatusForbidden, rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "cf-turnstile") {
		t.Errorf("expected turnstile widget in body")
	}
	if !strings.Contains(body, "0x4AAAAAAtestkey") {
		t.Errorf("expected sitekey in body")
	}
}

func TestResponseHandler_Redirect(t *testing.T) {
	cfg := &routewarden.ResponseConfig{
		Mode:        "redirect",
		StatusCode:  http.StatusFound,
		RedirectURL: "https://example.com/blocked",
	}

	handler, err := routewarden.NewResponseHandler(cfg, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("expected %d, got %d", http.StatusFound, rr.Code)
	}
	if rr.Header().Get("Location") != "https://example.com/blocked" {
		t.Errorf("expected Location header")
	}
}

func TestResponseHandler_SilentDrop(t *testing.T) {
	handler, err := routewarden.NewResponseHandler(nil, http.StatusForbidden, "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected %d, got %d", http.StatusForbidden, rr.Code)
	}
	if rr.Body.Len() > 0 {
		t.Errorf("expected empty body for silent drop")
	}
}

func TestResponseHandler_InvalidCaptchaTemplate(t *testing.T) {
	cfg := &routewarden.ResponseConfig{
		Mode: "captcha",
		Captcha: &routewarden.CaptchaConfig{
			Template: "{{.UnclosedBracket",
		},
	}

	_, err := routewarden.NewResponseHandler(cfg, 0, "", false)
	if err == nil {
		t.Errorf("expected error for invalid captcha template")
	}
}

func TestResponseHandler_DefaultTextAndEmptyFallbacks(t *testing.T) {
	// 1. Default text mode with top-level message
	handlerText, err := routewarden.NewResponseHandler(nil, http.StatusForbidden, "Access Denied by Text", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr1 := httptest.NewRecorder()
	handlerText.ServeBlockedRequest(rr1, req1)

	if !strings.Contains(rr1.Body.String(), "Access Denied by Text") {
		t.Errorf("expected default text response, got %s", rr1.Body.String())
	}
	if !strings.Contains(rr1.Header().Get("Content-Type"), "text/plain") {
		t.Errorf("expected text/plain content-type, got %s", rr1.Header().Get("Content-Type"))
	}

	// 2. JSON mode with empty body fallback
	handlerJSON, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:       "json",
		StatusCode: http.StatusForbidden,
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr2 := httptest.NewRecorder()
	handlerJSON.ServeBlockedRequest(rr2, req2)

	if !strings.Contains(rr2.Body.String(), `"error":"Forbidden"`) {
		t.Errorf("expected default json payload, got %s", rr2.Body.String())
	}

	// 3. HTML mode with empty body fallback
	handlerHTML, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:       "html",
		StatusCode: http.StatusNotFound,
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr3 := httptest.NewRecorder()
	handlerHTML.ServeBlockedRequest(rr3, req3)

	if !strings.Contains(rr3.Body.String(), "404 Forbidden") && !strings.Contains(rr3.Body.String(), "Access to this resource is denied") {
		t.Errorf("expected default html payload, got %s", rr3.Body.String())
	}

	// 4. Custom Captcha Template
	handlerCustomCaptcha, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:       "captcha",
		StatusCode: http.StatusForbidden,
		Captcha: &routewarden.CaptchaConfig{
			Template: "<div>{{.Title}} - SiteKey: {{.SiteKey}}</div>",
			Title:    "Custom Challenge",
			SiteKey:  "my-custom-key-999",
		},
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req4 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr4 := httptest.NewRecorder()
	handlerCustomCaptcha.ServeBlockedRequest(rr4, req4)

	if !strings.Contains(rr4.Body.String(), "Custom Challenge - SiteKey: my-custom-key-999") {
		t.Errorf("expected custom captcha template output, got %s", rr4.Body.String())
	}
}

