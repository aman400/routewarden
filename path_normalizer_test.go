package routewarden_test

import (
	"testing"

	"github.com/routewarden/traefik-warden"
)

func TestPathNormalizer_ExtractCandidatePaths(t *testing.T) {
	tests := []struct {
		name             string
		rawPath          string
		path             string
		requestURI       string
		expectedContains []string
	}{
		{
			name:             "Standard path",
			path:             "/api/v1/users",
			expectedContains: []string{"/api/v1/users"},
		},
		{
			name:             "Single URL encoded dot (%2eenv)",
			path:             "/%2eenv",
			expectedContains: []string{"/.env"},
		},
		{
			name:             "Double URL encoded dot (%252eenv)",
			path:             "/%252eenv",
			expectedContains: []string{"/.env"},
		},
		{
			name:             "Double URL encoded path traversal (%252e%252e)",
			path:             "/static/%252e%252e/.env",
			expectedContains: []string{"/.env"},
		},
		{
			name:             "RawPath difference provided",
			rawPath:          "/raw/%2eenv",
			path:             "/raw/.env",
			expectedContains: []string{"/raw/.env"},
		},
		{
			name:             "Semicolon matrix parameter prefix (/;.env)",
			path:             "/;.env",
			expectedContains: []string{"/.env"},
		},
		{
			name:             "Semicolon matrix parameter in segment (/app;jsessionid=123/.env)",
			path:             "/app;jsessionid=123/.env",
			expectedContains: []string{"/app/.env"},
		},
		{
			name:             "Semicolon embedded within segment (/api;.env/config.json)",
			path:             "/api;.env/config.json",
			expectedContains: []string{"/api/.env/config.json"},
		},
		{
			name:             "Windows backslash separator (\\..\\.env)",
			path:             "/static\\..\\.env",
			expectedContains: []string{"/.env"},
		},
		{
			name:             "Direct backslash path (/\\.env)",
			path:             "/\\.env",
			expectedContains: []string{"/.env"},
		},
		{
			name:             "Null byte in path",
			path:             "/.env\x00.png",
			expectedContains: []string{"/.env.png"},
		},
		{
			name:             "RequestURI query string stripped for candidate path",
			path:             "/search",
			requestURI:       "/search?file=.env",
			expectedContains: []string{"/search"},
		},
		{
			name:             "Dot slash canonical path (/./.env)",
			path:             "/./.env",
			expectedContains: []string{"/.env"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidates := routewarden.ExtractCandidatePaths(tc.rawPath, tc.path, tc.requestURI)

			for _, expected := range tc.expectedContains {
				found := false
				for _, c := range candidates {
					if c == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected candidate %q to be in candidates %v", expected, candidates)
				}
			}
		})
	}
}
