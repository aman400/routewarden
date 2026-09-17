package routewarden

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// Default Captcha HTML challenge template
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

// ResponseHandler manages custom response execution (JSON, HTML, Captcha, Redirect, Text).
type ResponseHandler struct {
	config          *ResponseConfig
	silentDrop      bool
	captchaTemplate *template.Template
}

// NewResponseHandler initializes a ResponseHandler with compiled templates.
func NewResponseHandler(respCfg *ResponseConfig, topStatusCode int, topCustomText string, silentDrop bool) (*ResponseHandler, error) {
	if respCfg == nil {
		code := topStatusCode
		if code == 0 {
			code = http.StatusForbidden
		}
		respCfg = &ResponseConfig{
			Mode:       "text",
			StatusCode: code,
			Body:       topCustomText,
		}
	} else {
		if respCfg.StatusCode == 0 {
			if topStatusCode != 0 {
				respCfg.StatusCode = topStatusCode
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
			return nil, fmt.Errorf("invalid captcha template: %w", err)
		}
	}

	return &ResponseHandler{
		config:          respCfg,
		silentDrop:      silentDrop,
		captchaTemplate: parsedTmpl,
	}, nil
}

// ServeBlockedRequest handles writing the configured response to the client.
func (h *ResponseHandler) ServeBlockedRequest(w http.ResponseWriter, req *http.Request) {
	if h.silentDrop {
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, err := hj.Hijack()
			if err == nil {
				_ = conn.Close()
				return
			}
		}
		w.WriteHeader(h.config.StatusCode)
		return
	}

	// Apply custom headers
	for k, v := range h.config.Headers {
		w.Header().Set(k, v)
	}

	switch strings.ToLower(h.config.Mode) {
	case "redirect":
		target := h.config.RedirectURL
		if target == "" {
			target = "/"
		}
		code := h.config.StatusCode
		if code < 300 || code > 308 {
			code = http.StatusFound
		}
		http.Redirect(w, req, target, code)

	case "json":
		contentType := h.config.ContentType
		if contentType == "" {
			contentType = "application/json"
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(h.config.StatusCode)

		body := h.config.Body
		if strings.TrimSpace(body) == "" {
			body = fmt.Sprintf(`{"error":"Forbidden","status":%d,"message":"Access to sensitive endpoint is blocked"}`, h.config.StatusCode)
		}
		_, _ = fmt.Fprintln(w, body)

	case "html":
		contentType := h.config.ContentType
		if contentType == "" {
			contentType = "text/html; charset=utf-8"
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(h.config.StatusCode)

		body := h.config.Body
		if strings.TrimSpace(body) == "" {
			body = fmt.Sprintf("<!DOCTYPE html><html><head><title>Access Denied</title></head><body><h1>%d Forbidden</h1><p>Access to this resource is denied.</p></body></html>", h.config.StatusCode)
		}
		_, _ = fmt.Fprintln(w, body)

	case "captcha":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(h.config.StatusCode)

		data := struct {
			Provider string
			SiteKey  string
			Title    string
		}{
			Provider: "turnstile",
			SiteKey:  "",
			Title:    "Security Check Required",
		}

		if h.config.Captcha != nil {
			if h.config.Captcha.Provider != "" {
				data.Provider = strings.ToLower(h.config.Captcha.Provider)
			}
			data.SiteKey = h.config.Captcha.SiteKey
			if h.config.Captcha.Title != "" {
				data.Title = h.config.Captcha.Title
			}
		}

		if h.captchaTemplate != nil {
			var buf bytes.Buffer
			if err := h.captchaTemplate.Execute(&buf, data); err == nil {
				_, _ = w.Write(buf.Bytes())
				return
			}
		}
		_, _ = fmt.Fprintln(w, "Security Challenge Required")

	case "gzipbomb", "bomb":
		contentType := h.config.ContentType
		if contentType == "" {
			contentType = "text/html; charset=UTF-8"
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(h.config.StatusCode)

		// Size in MB to generate. Each 1MB of zero-bytes compresses to ~1KB in gzip,
		// expanding ~1000x on the client during decompression.
		targetMB := h.config.GzipBombMB
		if targetMB <= 0 {
			targetMB = 10 // Default: 10MB expands to ~10GB on decompression
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			gz = gzip.NewWriter(w)
		}
		defer gz.Close()

		// Stream 32KB zero chunks through gzip writer
		zeroChunk := make([]byte, 32*1024)
		totalChunks := (targetMB * 1024 * 1024) / len(zeroChunk)
		if totalChunks <= 0 {
			totalChunks = 32
		}

		for i := 0; i < totalChunks; i++ {
			if _, err := gz.Write(zeroChunk); err != nil {
				// Client hung up, timed out, or crashed due to memory exhaustion
				break
			}
		}

	default: // "text"
		contentType := h.config.ContentType
		if contentType == "" {
			contentType = "text/plain; charset=utf-8"
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(h.config.StatusCode)
		if h.config.Body != "" {
			_, _ = fmt.Fprintln(w, h.config.Body)
		}
	}
}
