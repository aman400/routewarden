// Package routewarden is a Traefik middleware plugin to block sensitive endpoints and files,
// with support for custom path regexes, and responses in JSON, HTML, Captcha, or custom status codes.
package routewarden

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

// DefaultBlockPatterns contains well-known sensitive endpoints and file extensions.
var DefaultBlockPatterns = []string{
	// Sensitive extensions & environment files (e.g. .env, .env.local, .txt, .log, .bak, .backup, .sql, .conf, .config, .ini, .yaml, .yml)
	`(?i)(^|/)(\.env.*|.*\.(txt|log|bak|backup|sql|conf|config|ini|yaml|yml))$`,
	// Version control & sensitive hidden directories
	`(?i)(^|/)\.(git|svn|hg|bzr|cvs)(/.*|$)`,
	// Cloud & infra credentials
	`(?i)(^|/)\.(aws|ssh|kube|docker)(/.*|$)`,
	// Database & server dump files / archives
	`(?i).*\.(tar|tar\.gz|tgz|zip|rar|7z|gz|bz2|iso|dump|sqlite|sqlite3|db)$`,
	// Common sensitive admin & debug endpoints
	`(?i)(^|/)(phpinfo\.php|info\.php|server-status|server-info|actuator(/.*)?|metrics|heapdump|trace|env)$`,
	// Package manager files & lockfiles
	`(?i)(^|/)(composer\.(json|lock)|package-lock\.json|yarn\.lock|pnpm-lock\.yaml|Pipfile|Pipfile\.lock|requirements\.txt)$`,
}

// DefaultAllowPatterns contains typical legitimate endpoints that might otherwise match broad patterns.
var DefaultAllowPatterns = []string{
	`(?i)^/robots\.txt$`,
	`(?i)^/sitemap.*\.xml$`,
	`(?i)^/ads\.txt$`,
	`(?i)^/security\.txt$`,
	`(?i)^/\.well-known(/.*)?$`,
}

// CaptchaConfig holds captcha configuration options.
type CaptchaConfig struct {
	Provider string `json:"provider,omitempty"` // "turnstile", "hcaptcha", "recaptcha", or "custom"
	SiteKey  string `json:"siteKey,omitempty"`  // Public site key
	Title    string `json:"title,omitempty"`    // Challenge page title
	Template string `json:"template,omitempty"` // Custom HTML template
}

// ResponseConfig defines how blocked requests should be answered.
type ResponseConfig struct {
	Mode        string            `json:"mode,omitempty"`        // "text", "json", "html", "captcha", "redirect"
	StatusCode  int               `json:"statusCode,omitempty"`  // HTTP status code (e.g. 403, 404, 429)
	ContentType string            `json:"contentType,omitempty"` // Custom Content-Type header override
	Body        string            `json:"body,omitempty"`        // Response payload (JSON string, HTML, or text)
	Headers     map[string]string `json:"headers,omitempty"`     // Custom response headers (e.g. Retry-After, X-Blocked-By)
	RedirectURL string            `json:"redirectUrl,omitempty"` // Target URL when Mode is "redirect"
	Captcha     *CaptchaConfig    `json:"captcha,omitempty"`     // Captcha settings when Mode is "captcha"
}

// Config holds the plugin configuration.
type Config struct {
	Enabled               bool            `json:"enabled,omitempty"`
	EnableDefaultPatterns bool            `json:"enableDefaultPatterns,omitempty"`
	PathPatterns          []string        `json:"pathPatterns,omitempty"` // Synonym for blockPatterns
	BlockPatterns         []string        `json:"blockPatterns,omitempty"`
	AllowPatterns         []string        `json:"allowPatterns,omitempty"`
	StatusCode            int             `json:"statusCode,omitempty"`
	CustomResponseText    string          `json:"customResponseText,omitempty"`
	SilentDrop            bool            `json:"silentDrop,omitempty"`
	CheckQuery            bool            `json:"checkQuery,omitempty"`
	Response              *ResponseConfig `json:"response,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Enabled:               true,
		EnableDefaultPatterns: true,
		PathPatterns:          []string{},
		BlockPatterns:         []string{},
		AllowPatterns:         DefaultAllowPatterns,
		StatusCode:            http.StatusForbidden,
		CustomResponseText:    "403 Forbidden: Access to sensitive endpoint is blocked",
		SilentDrop:            false,
		CheckQuery:            false,
		Response:              nil,
	}
}

// RouteWarden is the Traefik middleware plugin handler.
type RouteWarden struct {
	next            http.Handler
	name            string
	enabled         bool
	blockRegexes    []*regexp.Regexp
	allowRegexes    []*regexp.Regexp
	silentDrop      bool
	checkQuery      bool
	responseCfg     *ResponseConfig
	captchaTemplate *template.Template
}

// Default Captcha HTML template
const defaultCaptchaHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}}</title>
  <style>
    :root {
      --bg: #0f172a;
      --card: #1e293b;
      --text: #f8fafc;
      --subtext: #94a3b8;
      --accent: #3b82f6;
      --border: #334155;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      background: var(--bg);
      color: var(--text);
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
      padding: 1.5rem;
    }
    .card {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 1rem;
      padding: 2.5rem;
      max-width: 480px;
      width: 100%;
      text-align: center;
      box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5), 0 8px 10px -6px rgba(0, 0, 0, 0.5);
    }
    .shield-icon {
      width: 56px;
      height: 56px;
      margin: 0 auto 1.25rem;
      color: var(--accent);
    }
    h1 {
      font-size: 1.5rem;
      font-weight: 700;
      margin-bottom: 0.75rem;
      color: var(--text);
    }
    p {
      color: var(--subtext);
      font-size: 0.95rem;
      line-height: 1.5;
      margin-bottom: 2rem;
    }
    .captcha-container {
      display: flex;
      justify-content: center;
      margin-bottom: 1.5rem;
      min-height: 70px;
    }
    .footer {
      font-size: 0.8rem;
      color: var(--subtext);
      opacity: 0.75;
    }
  </style>
  {{if eq .Provider "turnstile"}}
  <script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>
  {{else if eq .Provider "hcaptcha"}}
  <script src="https://js.hcaptcha.com/1/api.js" async defer></script>
  {{else if eq .Provider "recaptcha"}}
  <script src="https://www.google.com/recaptcha/api.js" async defer></script>
  {{end}}
</head>
<body>
  <div class="card">
    <svg class="shield-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
      <path d="M9 12l2 2 4-4"/>
    </svg>
    <h1>{{.Title}}</h1>
    <p>Please complete the security challenge below to verify you are a human visitor before proceeding.</p>

    <form method="POST" action="">
      <div class="captcha-container">
        {{if eq .Provider "turnstile"}}
        <div class="cf-turnstile" data-sitekey="{{.SiteKey}}" data-theme="dark"></div>
        {{else if eq .Provider "hcaptcha"}}
        <div class="h-captcha" data-sitekey="{{.SiteKey}}" data-theme="dark"></div>
        {{else if eq .Provider "recaptcha"}}
        <div class="g-recaptcha" data-sitekey="{{.SiteKey}}" data-theme="dark"></div>
        {{else}}
        <div class="custom-captcha">{{.SiteKey}}</div>
        {{end}}
      </div>
    </form>
    <div class="footer">Protected by RouteWarden Security</div>
  </div>
</body>
</html>`

// New creates a new RouteWarden plugin handler.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config == nil {
		config = CreateConfig()
	}

	var blockPatterns []string
	if config.EnableDefaultPatterns {
		blockPatterns = append(blockPatterns, DefaultBlockPatterns...)
	}
	blockPatterns = append(blockPatterns, config.PathPatterns...)
	blockPatterns = append(blockPatterns, config.BlockPatterns...)

	compiledBlockRegexes := make([]*regexp.Regexp, 0, len(blockPatterns))
	for _, p := range blockPatterns {
		if strings.TrimSpace(p) == "" {
			continue
		}
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("routewarden [%s]: invalid block regex pattern %q: %w", name, p, err)
		}
		compiledBlockRegexes = append(compiledBlockRegexes, re)
	}

	compiledAllowRegexes := make([]*regexp.Regexp, 0, len(config.AllowPatterns))
	for _, p := range config.AllowPatterns {
		if strings.TrimSpace(p) == "" {
			continue
		}
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("routewarden [%s]: invalid allow regex pattern %q: %w", name, p, err)
		}
		compiledAllowRegexes = append(compiledAllowRegexes, re)
	}

	// Normalize response configuration
	respCfg := config.Response
	if respCfg == nil {
		statusCode := config.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusForbidden
		}
		respCfg = &ResponseConfig{
			Mode:       "text",
			StatusCode: statusCode,
			Body:       config.CustomResponseText,
		}
	} else {
		if respCfg.StatusCode == 0 {
			if config.StatusCode != 0 {
				respCfg.StatusCode = config.StatusCode
			} else {
				respCfg.StatusCode = http.StatusForbidden
			}
		}
		if respCfg.Mode == "" {
			respCfg.Mode = "text"
		}
	}

	var parsedTmpl *template.Template
	if respCfg.Mode == "captcha" {
		tmplText := defaultCaptchaHTML
		if respCfg.Captcha != nil && strings.TrimSpace(respCfg.Captcha.Template) != "" {
			tmplText = respCfg.Captcha.Template
		}
		var err error
		parsedTmpl, err = template.New("captcha").Parse(tmplText)
		if err != nil {
			return nil, fmt.Errorf("routewarden [%s]: invalid captcha template: %w", name, err)
		}
	}

	return &RouteWarden{
		next:            next,
		name:            name,
		enabled:         config.Enabled,
		blockRegexes:    compiledBlockRegexes,
		allowRegexes:    compiledAllowRegexes,
		silentDrop:      config.SilentDrop,
		checkQuery:      config.CheckQuery,
		responseCfg:     respCfg,
		captchaTemplate: parsedTmpl,
	}, nil
}

func (rw *RouteWarden) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !rw.enabled {
		rw.next.ServeHTTP(w, req)
		return
	}

	// Canonicalize and inspect paths.
	// 1. Check direct raw path
	pathsToCheck := []string{path.Clean(req.URL.Path)}

	// 2. Add RequestURI path before query to catch raw gateway discrepancies
	rawURIPath := req.RequestURI
	if idx := strings.IndexByte(rawURIPath, '?'); idx != -1 {
		rawURIPath = rawURIPath[:idx]
	}
	if rawURIPath != "" {
		pathsToCheck = append(pathsToCheck, path.Clean(rawURIPath))
	}

	// 3. Add RawPath if specified
	if req.URL.RawPath != "" && req.URL.RawPath != req.URL.Path {
		pathsToCheck = append(pathsToCheck, path.Clean(req.URL.RawPath))
	}

	// 4. Perform iterative unescaping to prevent multi-layer URL encoding evasion (e.g. %252e%252e)
	curPath := req.URL.Path
	for i := 0; i < 3; i++ {
		unescaped, err := url.PathUnescape(curPath)
		if err != nil || unescaped == curPath {
			break
		}
		pathsToCheck = append(pathsToCheck, path.Clean(unescaped))
		curPath = unescaped
	}

	// 5. Check backslash-converted paths (Windows / IIS style path traversal / separator evasion)
	for _, p := range append([]string{}, pathsToCheck...) {
		if strings.ContainsRune(p, '\\') {
			slashConverted := strings.ReplaceAll(p, "\\", "/")
			pathsToCheck = append(pathsToCheck, path.Clean(slashConverted))
		}
	}

	// 6. Semicolon matrix parameter handling (e.g. /;.env, /api;.env, /static;jsessionid=123/.env)
	for _, p := range append([]string{}, pathsToCheck...) {
		if strings.ContainsRune(p, ';') {
			// Option A: treat semicolon as delimiter / strip matrix params
			parts := strings.Split(p, "/")
			cleanedSegments := make([]string, len(parts))
			paramSegments := make([]string, 0)
			for i, seg := range parts {
				if semiIdx := strings.IndexByte(seg, ';'); semiIdx != -1 {
					cleanedSegments[i] = seg[:semiIdx]
					paramSegments = append(paramSegments, seg[semiIdx+1:])
				} else {
					cleanedSegments[i] = seg
				}
			}
			matrixStripped := strings.Join(cleanedSegments, "/")
			pathsToCheck = append(pathsToCheck, path.Clean(matrixStripped))

			// Option B: check stripped parameters individually as virtual paths
			for _, param := range paramSegments {
				if param != "" {
					pathsToCheck = append(pathsToCheck, "/"+param, path.Clean("/"+param))
				}
			}

			// Option C: replace semicolon with slash to inspect embedded subpaths (e.g. /api;.env -> /api/.env)
			semiAsSlash := strings.ReplaceAll(p, ";", "/")
			pathsToCheck = append(pathsToCheck, path.Clean(semiAsSlash))
		}
	}

	// 7. Strip null bytes
	for _, p := range append([]string{}, pathsToCheck...) {
		if strings.ContainsRune(p, '\x00') {
			pathsToCheck = append(pathsToCheck, path.Clean(strings.ReplaceAll(p, "\x00", "")))
		}
	}

	// Deduplicate candidates
	candidatePaths := make([]string, 0, len(pathsToCheck))
	seen := make(map[string]struct{}, len(pathsToCheck))
	for _, p := range pathsToCheck {
		if _, exists := seen[p]; !exists && p != "" {
			seen[p] = struct{}{}
			candidatePaths = append(candidatePaths, p)
		}
	}

	// 1. Check AllowPatterns first (Allowlist override)
	for _, p := range candidatePaths {
		if rw.isAllowed(p) {
			rw.next.ServeHTTP(w, req)
			return
		}
	}

	// 2. Check BlockPatterns against URL paths
	for _, p := range candidatePaths {
		if rw.isBlocked(p) {
			rw.blockRequest(w, req)
			return
		}
	}

	// 3. Optional: Check Query String if enabled
	if rw.checkQuery && req.URL.RawQuery != "" {
		unescapedQuery, err := url.QueryUnescape(req.URL.RawQuery)
		if err != nil {
			unescapedQuery = req.URL.RawQuery
		}

		if rw.isBlocked(unescapedQuery) || rw.isBlocked(req.URL.RawQuery) {
			rw.blockRequest(w, req)
			return
		}

		queryParams := req.URL.Query()
		for _, values := range queryParams {
			for _, val := range values {
				if rw.isBlocked(val) {
					rw.blockRequest(w, req)
					return
				}
			}
		}
	}

	rw.next.ServeHTTP(w, req)
}

func (rw *RouteWarden) isAllowed(target string) bool {
	for _, re := range rw.allowRegexes {
		if re.MatchString(target) {
			return true
		}
	}
	return false
}

func (rw *RouteWarden) isBlocked(target string) bool {
	for _, re := range rw.blockRegexes {
		if re.MatchString(target) {
			return true
		}
	}
	return false
}

func (rw *RouteWarden) blockRequest(w http.ResponseWriter, req *http.Request) {
	if rw.silentDrop {
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, err := hj.Hijack()
			if err == nil {
				_ = conn.Close()
				return
			}
		}
		w.WriteHeader(rw.responseCfg.StatusCode)
		return
	}

	// Apply custom headers
	for k, v := range rw.responseCfg.Headers {
		w.Header().Set(k, v)
	}

	switch strings.ToLower(rw.responseCfg.Mode) {
	case "redirect":
		target := rw.responseCfg.RedirectURL
		if target == "" {
			target = "/"
		}
		code := rw.responseCfg.StatusCode
		if code < 300 || code > 308 {
			code = http.StatusFound
		}
		http.Redirect(w, req, target, code)

	case "json":
		contentType := rw.responseCfg.ContentType
		if contentType == "" {
			contentType = "application/json"
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(rw.responseCfg.StatusCode)

		body := rw.responseCfg.Body
		if strings.TrimSpace(body) == "" {
			body = fmt.Sprintf(`{"error":"Forbidden","status":%d,"message":"Access to sensitive endpoint is blocked"}`, rw.responseCfg.StatusCode)
		}
		_, _ = fmt.Fprintln(w, body)

	case "html":
		contentType := rw.responseCfg.ContentType
		if contentType == "" {
			contentType = "text/html; charset=utf-8"
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(rw.responseCfg.StatusCode)

		body := rw.responseCfg.Body
		if strings.TrimSpace(body) == "" {
			body = fmt.Sprintf("<!DOCTYPE html><html><head><title>Access Denied</title></head><body><h1>%d Forbidden</h1><p>Access to this resource is denied.</p></body></html>", rw.responseCfg.StatusCode)
		}
		_, _ = fmt.Fprintln(w, body)

	case "captcha":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(rw.responseCfg.StatusCode)

		data := struct {
			Provider string
			SiteKey  string
			Title    string
		}{
			Provider: "turnstile",
			SiteKey:  "",
			Title:    "Security Check Required",
		}

		if rw.responseCfg.Captcha != nil {
			if rw.responseCfg.Captcha.Provider != "" {
				data.Provider = strings.ToLower(rw.responseCfg.Captcha.Provider)
			}
			data.SiteKey = rw.responseCfg.Captcha.SiteKey
			if rw.responseCfg.Captcha.Title != "" {
				data.Title = rw.responseCfg.Captcha.Title
			}
		}

		if rw.captchaTemplate != nil {
			var buf bytes.Buffer
			if err := rw.captchaTemplate.Execute(&buf, data); err == nil {
				_, _ = w.Write(buf.Bytes())
				return
			}
		}
		_, _ = fmt.Fprintln(w, "Security Challenge Required")

	default: // "text" or unspecified
		contentType := rw.responseCfg.ContentType
		if contentType == "" {
			contentType = "text/plain; charset=utf-8"
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(rw.responseCfg.StatusCode)
		if rw.responseCfg.Body != "" {
			_, _ = fmt.Fprintln(w, rw.responseCfg.Body)
		}
	}
}
