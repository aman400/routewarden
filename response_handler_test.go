package routewarden_test

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/routewarden/traefik-warden"
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

func TestResponseHandler_GzipBomb(t *testing.T) {
	// 1. Test gzipBomb mode with default size (10MB)
	cfg := &routewarden.ResponseConfig{
		Mode:       "gzipBomb",
		StatusCode: http.StatusOK,
	}

	handler, err := routewarden.NewResponseHandler(cfg, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding: gzip, got %s", rr.Header().Get("Content-Encoding"))
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "text/html") {
		t.Errorf("expected text/html Content-Type, got %s", rr.Header().Get("Content-Type"))
	}

	// Verify the payload is valid gzip and expands
	gzReader, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader from response: %v", err)
	}
	defer gzReader.Close()

	// Read first 1MB of decompressed stream to verify it's zero bytes without exhausting test RAM
	buf := make([]byte, 1024*1024)
	n, err := io.ReadFull(gzReader, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		t.Fatalf("failed reading uncompressed stream: %v", err)
	}
	if n != len(buf) {
		t.Errorf("expected to read at least 1MB of decompressed zeroes, read %d bytes", n)
	}
	for i := 0; i < 1024; i++ {
		if buf[i] != 0 {
			t.Errorf("expected byte 0 at pos %d, got %d", i, buf[i])
			break
		}
	}

	// 2. Test alias mode "bomb" with custom size and custom status code
	cfgCustom := &routewarden.ResponseConfig{
		Mode:        "bomb",
		StatusCode:  http.StatusForbidden,
		GzipBombMB:  2,
		ContentType: "text/plain",
	}

	handlerCustom, err := routewarden.NewResponseHandler(cfgCustom, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rrCustom := httptest.NewRecorder()
	handlerCustom.ServeBlockedRequest(rrCustom, req)

	if rrCustom.Code != http.StatusForbidden {
		t.Errorf("expected %d, got %d", http.StatusForbidden, rrCustom.Code)
	}
	if rrCustom.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding: gzip")
	}
	if rrCustom.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("expected text/plain Content-Type, got %s", rrCustom.Header().Get("Content-Type"))
	}
}

func TestResponseHandler_XML(t *testing.T) {
	// Default XML
	handler, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:       "xml",
		StatusCode: http.StatusForbidden,
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/endpoint", nil)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected %d, got %d", http.StatusForbidden, rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/xml") {
		t.Errorf("expected application/xml content-type")
	}
	if !strings.Contains(rr.Body.String(), "<Error>") || !strings.Contains(rr.Body.String(), "<Status>403</Status>") {
		t.Errorf("unexpected xml body: %s", rr.Body.String())
	}

	// Custom XML body
	handlerCustom, _ := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:       "xml",
		StatusCode: http.StatusUnauthorized,
		Body:       "<soap:Fault><faultcode>Client</faultcode></soap:Fault>",
	}, 0, "", false)
	rrCustom := httptest.NewRecorder()
	handlerCustom.ServeBlockedRequest(rrCustom, req)
	if !strings.Contains(rrCustom.Body.String(), "<soap:Fault>") {
		t.Errorf("expected custom XML fault body")
	}
}

func TestResponseHandler_RateLimitChallenge(t *testing.T) {
	handler, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:              "rateLimitChallenge",
		RetryAfterSeconds: 600,
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") != "600" {
		t.Errorf("expected Retry-After: 600, got %s", rr.Header().Get("Retry-After"))
	}
	if !strings.Contains(rr.Body.String(), `"retryAfter":600`) {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}

func TestResponseHandler_FakeSuccessDecoy(t *testing.T) {
	handler, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode: "fakeSuccess",
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test .env synthetic response
	reqEnv := httptest.NewRequest(http.MethodGet, "/.env", nil)
	rrEnv := httptest.NewRecorder()
	handler.ServeBlockedRequest(rrEnv, reqEnv)
	if rrEnv.Code != http.StatusOK {
		t.Errorf("expected 200 OK for decoy")
	}
	if !strings.Contains(rrEnv.Body.String(), "APP_NAME=Laravel") || !strings.Contains(rrEnv.Body.String(), "DB_PASSWORD=") {
		t.Errorf("expected synthetic .env body, got: %s", rrEnv.Body.String())
	}

	// Test actuator health decoy
	reqActuator := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	rrActuator := httptest.NewRecorder()
	handler.ServeBlockedRequest(rrActuator, reqActuator)
	if !strings.Contains(rrActuator.Body.String(), `"status":"UP"`) {
		t.Errorf("expected actuator decoy, got: %s", rrActuator.Body.String())
	}

	// Test git/HEAD decoy
	reqGit := httptest.NewRequest(http.MethodGet, "/.git/HEAD", nil)
	rrGit := httptest.NewRecorder()
	handler.ServeBlockedRequest(rrGit, reqGit)
	if !strings.Contains(rrGit.Body.String(), "ref: refs/heads/master") {
		t.Errorf("expected git decoy, got: %s", rrGit.Body.String())
	}

	// Test phpinfo decoy
	reqPHP := httptest.NewRequest(http.MethodGet, "/phpinfo.php", nil)
	rrPHP := httptest.NewRecorder()
	handler.ServeBlockedRequest(rrPHP, reqPHP)
	if !strings.Contains(rrPHP.Body.String(), "phpinfo()") {
		t.Errorf("expected phpinfo decoy, got: %s", rrPHP.Body.String())
	}

	// Test wp-login decoy
	reqWP := httptest.NewRequest(http.MethodGet, "/wp-login.php", nil)
	rrWP := httptest.NewRecorder()
	handler.ServeBlockedRequest(rrWP, reqWP)
	if !strings.Contains(rrWP.Body.String(), "WordPress") {
		t.Errorf("expected wp-login decoy, got: %s", rrWP.Body.String())
	}

	// Test generic route decoy
	reqGeneric := httptest.NewRequest(http.MethodGet, "/api/something", nil)
	rrGeneric := httptest.NewRecorder()
	handler.ServeBlockedRequest(rrGeneric, reqGeneric)
	if !strings.Contains(rrGeneric.Body.String(), `"status":"success"`) {
		t.Errorf("expected generic success decoy")
	}
}

func TestResponseHandler_Proxy(t *testing.T) {
	// 1. Test Proxy handler with custom RoundTripper so it doesn't need to bind network ports in sandbox
	targetURL, _ := url.Parse("http://honeypot.local")
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	type testTransport struct{}
	proxy.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.Header().Set("X-Honeypot-Captured", "true")
		rec.WriteHeader(http.StatusTeapot)
		_, _ = rec.WriteString("honeypot-captured")
		resp := rec.Result()
		resp.Request = req
		return resp, nil
	})

	handler, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:     "proxy",
		ProxyURL: "http://honeypot.local",
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Inject test proxy
	handler.SetProxyHandlerForTest(proxy)

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Errorf("expected %d from honeypot backend, got %d", http.StatusTeapot, rr.Code)
	}
	if rr.Header().Get("X-Honeypot-Captured") != "true" {
		t.Errorf("expected proxy header")
	}
	if !strings.Contains(rr.Body.String(), "honeypot-captured") {
		t.Errorf("expected honeypot body")
	}

	// 2. Test invalid proxy URL error
	_, errInvalid := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:     "proxy",
		ProxyURL: "://invalid-url",
	}, 0, "", false)
	if errInvalid == nil {
		t.Errorf("expected error for invalid proxy URL")
	}

	// 3. Test empty proxy fallback
	handlerEmpty, _ := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode: "proxy",
	}, 0, "", false)
	rrEmpty := httptest.NewRecorder()
	handlerEmpty.ServeBlockedRequest(rrEmpty, req)
	if rrEmpty.Code != http.StatusBadGateway {
		t.Errorf("expected 502 for unconfigured proxy, got %d", rrEmpty.Code)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestResponseHandler_InfiniteStream(t *testing.T) {
	handler, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:         "infiniteStream",
		StatusCode:   http.StatusOK,
		StreamSizeMB: 1, // 1MB in test
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.Len() < 1024*1024 {
		t.Errorf("expected at least 1MB garbage stream, got %d bytes", rr.Body.Len())
	}
}

func TestResponseHandler_Tarpit(t *testing.T) {
	handler, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:                     "tarpit",
		StatusCode:               http.StatusOK,
		TarpitDelayMs:           5,  // Fast delay for testing
		TarpitMaxDurationSeconds: 1,  // 1 second max
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test context cancellation exits cleanly
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/probe", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeBlockedRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	// Test tarpit natural timeout branch
	handlerTimeout, _ := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:                     "tarpit",
		TarpitDelayMs:           5,
		TarpitMaxDurationSeconds: 1, // 1 second timeout
	}, 0, "", false)
	reqTimeout := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rrTimeout := httptest.NewRecorder()
	handlerTimeout.ServeBlockedRequest(rrTimeout, reqTimeout)
	if rrTimeout.Code != http.StatusForbidden {
		t.Errorf("expected default 403 for tarpit without explicit code")
	}
}

func TestResponseHandler_EdgeCases(t *testing.T) {
	// 1. Custom Body & Status for FakeSuccess
	handlerCustomDecoy, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:        "fakeSuccess",
		StatusCode:  http.StatusAccepted,
		ContentType: "application/json",
		Body:        `{"custom":"decoy_payload"}`,
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	reqDecoy := httptest.NewRequest(http.MethodGet, "/.env", nil)
	rrDecoy := httptest.NewRecorder()
	handlerCustomDecoy.ServeBlockedRequest(rrDecoy, reqDecoy)
	if rrDecoy.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", rrDecoy.Code)
	}
	if !strings.Contains(rrDecoy.Body.String(), `"custom":"decoy_payload"`) {
		t.Errorf("expected custom decoy payload")
	}

	// 2. Redirect without code (should default to 302 Found)
	handlerRedir, err := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:        "redirect",
		RedirectURL: "https://example.com/honeypot",
		StatusCode:  200, // Invalid redirect code should fallback to 302
	}, 0, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rrRedir := httptest.NewRecorder()
	handlerRedir.ServeBlockedRequest(rrRedir, reqDecoy)
	if rrRedir.Code != http.StatusFound {
		t.Errorf("expected fallback to 302 Found, got %d", rrRedir.Code)
	}

	// 3. Redirect without redirectURL (should default to "/")
	handlerRedirDefault, _ := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode: "redirect",
	}, 0, "", false)
	rrRedirDefault := httptest.NewRecorder()
	handlerRedirDefault.ServeBlockedRequest(rrRedirDefault, reqDecoy)
	if rrRedirDefault.Header().Get("Location") != "/" {
		t.Errorf("expected Location: /, got %s", rrRedirDefault.Header().Get("Location"))
	}

	// 4. RateLimitChallenge with custom body and status code
	handlerRL, _ := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:              "rateLimit",
		StatusCode:        http.StatusTooManyRequests,
		RetryAfterSeconds: 120,
		ContentType:       "text/plain",
		Body:              "Calm down bot",
	}, 0, "", false)
	rrRL := httptest.NewRecorder()
	handlerRL.ServeBlockedRequest(rrRL, reqDecoy)
	if rrRL.Header().Get("Retry-After") != "120" {
		t.Errorf("expected Retry-After: 120, got %s", rrRL.Header().Get("Retry-After"))
	}
	if !strings.Contains(rrRL.Body.String(), "Calm down bot") {
		t.Errorf("expected custom rate limit body")
	}

	// 5. XML with custom content type and status code
	handlerXML, _ := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:        "xml",
		StatusCode:  http.StatusPaymentRequired,
		ContentType: "application/soap+xml",
	}, 0, "", false)
	rrXML := httptest.NewRecorder()
	handlerXML.ServeBlockedRequest(rrXML, reqDecoy)
	if rrXML.Code != http.StatusPaymentRequired {
		t.Errorf("expected 402, got %d", rrXML.Code)
	}
	if rrXML.Header().Get("Content-Type") != "application/soap+xml" {
		t.Errorf("expected custom xml content type")
	}

	// 6. InfiniteStream default fallback size (<=0 MB defaults to 50MB in production, tested with 0)
	handlerStreamZero, _ := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:         "infiniteStream",
		StreamSizeMB: -1,
	}, 0, "", false)
	if handlerStreamZero == nil {
		t.Errorf("failed creating infiniteStream handler")
	}

	// 7. GzipBomb with 0 MB defaults
	handlerBombZero, _ := routewarden.NewResponseHandler(&routewarden.ResponseConfig{
		Mode:       "gzipBomb",
		GzipBombMB: 0,
	}, 0, "", false)
	rrBombZero := httptest.NewRecorder()
	handlerBombZero.ServeBlockedRequest(rrBombZero, reqDecoy)
	if rrBombZero.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip encoding")
	}
}

