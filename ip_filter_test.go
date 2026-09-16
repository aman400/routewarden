package routewarden_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aman400/routewarden"
)

func TestIPFilter_Unit(t *testing.T) {
	filter, err := routewarden.NewIPFilter([]string{
		"192.168.1.10",
		"10.0.0.0/16",
		"2001:db8::/32",
	})
	if err != nil {
		t.Fatalf("unexpected error creating IPFilter: %v", err)
	}

	// 1. Exact IPv4 match in RemoteAddr
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "192.168.1.10:1234"
	if !filter.IsAllowed(req1) {
		t.Errorf("expected 192.168.1.10 to be allowed")
	}

	// 2. CIDR subnet match in RemoteAddr
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.5.20:5678"
	if !filter.IsAllowed(req2) {
		t.Errorf("expected 10.0.5.20 in 10.0.0.0/16 to be allowed")
	}

	// 3. IPv6 CIDR subnet match in RemoteAddr
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = "[2001:db8::1]:5678"
	if !filter.IsAllowed(req3) {
		t.Errorf("expected 2001:db8::1 in 2001:db8::/32 to be allowed")
	}

	// 4. X-Forwarded-For header match
	req4 := httptest.NewRequest(http.MethodGet, "/", nil)
	req4.RemoteAddr = "203.0.113.1:80"
	req4.Header.Set("X-Forwarded-For", "192.168.1.10, 203.0.113.1")
	if !filter.IsAllowed(req4) {
		t.Errorf("expected X-Forwarded-For first IP to be allowed")
	}

	// 5. X-Real-IP header match
	req5 := httptest.NewRequest(http.MethodGet, "/", nil)
	req5.RemoteAddr = "203.0.113.1:80"
	req5.Header.Set("X-Real-IP", "10.0.99.1")
	if !filter.IsAllowed(req5) {
		t.Errorf("expected X-Real-IP to be allowed")
	}

	// 6. Non-whitelisted IP
	req6 := httptest.NewRequest(http.MethodGet, "/", nil)
	req6.RemoteAddr = "203.0.113.50:4321"
	if filter.IsAllowed(req6) {
		t.Errorf("expected 203.0.113.50 to NOT be allowed")
	}

	// 7. Malformed / empty client IP
	req7 := httptest.NewRequest(http.MethodGet, "/", nil)
	req7.RemoteAddr = "invalid-address"
	if filter.IsAllowed(req7) {
		t.Errorf("expected invalid address to NOT be allowed")
	}
}

func TestIPFilter_EmptyFilter(t *testing.T) {
	filter, err := routewarden.NewIPFilter([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	if filter.IsAllowed(req) {
		t.Errorf("empty filter should never match")
	}
}

func TestIPFilter_InvalidInputs(t *testing.T) {
	_, err := routewarden.NewIPFilter([]string{"999.999.999.999"})
	if err == nil {
		t.Errorf("expected error for invalid IP address")
	}

	_, err2 := routewarden.NewIPFilter([]string{"10.0.0.0/999"})
	if err2 == nil {
		t.Errorf("expected error for invalid CIDR subnet")
	}
}
