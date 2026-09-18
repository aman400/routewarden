package traefik_warden

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// RouteWarden is the Traefik middleware plugin handler.
type RouteWarden struct {
	next            http.Handler
	name            string
	enabled         bool
	debug           bool
	methods         map[string]struct{}
	blockRegexes    []*regexp.Regexp
	allowRegexes    []*regexp.Regexp
	ipFilter        *IPFilter
	checkQuery      bool
	responseHandler *ResponseHandler
}

// New creates a new RouteWarden plugin handler.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config == nil {
		config = CreateConfig()
	}

	methodsMap := make(map[string]struct{})
	if len(config.Methods) == 0 {
		methodsMap["GET"] = struct{}{}
	} else {
		for _, m := range config.Methods {
			m = strings.ToUpper(strings.TrimSpace(m))
			if m != "" {
				methodsMap[m] = struct{}{}
			}
		}
		if len(methodsMap) == 0 {
			methodsMap["GET"] = struct{}{}
		}
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

	var allowPatterns []string
	if config.EnableDefaultAllowPatterns {
		allowPatterns = append(allowPatterns, DefaultAllowPatterns...)
	}
	allowPatterns = append(allowPatterns, config.AllowPatterns...)

	compiledAllowRegexes := make([]*regexp.Regexp, 0, len(allowPatterns))
	for _, p := range allowPatterns {
		if strings.TrimSpace(p) == "" {
			continue
		}
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("routewarden [%s]: invalid allow regex pattern %q: %w", name, p, err)
		}
		compiledAllowRegexes = append(compiledAllowRegexes, re)
	}

	ipFilter, err := NewIPFilter(config.AllowedIPs)
	if err != nil {
		return nil, fmt.Errorf("routewarden [%s]: %w", name, err)
	}

	respHandler, err := NewResponseHandler(config.Response, config.StatusCode, config.CustomResponseText, config.SilentDrop)
	if err != nil {
		return nil, fmt.Errorf("routewarden [%s]: %w", name, err)
	}

	return &RouteWarden{
		next:            next,
		name:            name,
		enabled:         config.Enabled,
		debug:           config.Debug,
		methods:         methodsMap,
		blockRegexes:    compiledBlockRegexes,
		allowRegexes:    compiledAllowRegexes,
		ipFilter:        ipFilter,
		checkQuery:      config.CheckQuery,
		responseHandler: respHandler,
	}, nil
}

func (rw *RouteWarden) logDebug(format string, v ...interface{}) {
	if rw.debug {
		log.Printf("[DEBUG] routewarden [%s]: "+format, append([]interface{}{rw.name}, v...)...)
	}
}

func (rw *RouteWarden) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !rw.enabled {
		rw.next.ServeHTTP(w, req)
		return
	}

	// Only inspect requests whose HTTP method matches configured verbs (default: GET)
	if _, matchesMethod := rw.methods[strings.ToUpper(req.Method)]; !matchesMethod {
		rw.logDebug("method %s not in inspected methods, bypassing", req.Method)
		rw.next.ServeHTTP(w, req)
		return
	}

	// Exempt whitelisted client IPs or CIDR subnets from blocking
	if rw.ipFilter.IsAllowed(req) {
		rw.logDebug("client IP %s is whitelisted, allowing request", req.RemoteAddr)
		rw.next.ServeHTTP(w, req)
		return
	}

	// Canonicalize and inspect paths with anti-evasion protections
	candidatePaths := ExtractCandidatePaths(req.URL.RawPath, req.URL.Path, req.RequestURI)
	rw.logDebug("inspecting request %s %s with %d candidate paths: %v", req.Method, req.URL.Path, len(candidatePaths), candidatePaths)

	// 1. Check AllowPatterns first (Allowlist override)
	for _, p := range candidatePaths {
		if re := rw.findMatchingAllow(p); re != nil {
			rw.logDebug("path %q allowed by pattern %q", p, re.String())
			rw.next.ServeHTTP(w, req)
			return
		}
	}

	// 2. Check BlockPatterns against URL paths
	for _, p := range candidatePaths {
		if re := rw.findMatchingBlock(p); re != nil {
			rw.logDebug("path %q blocked by pattern %q (mode: %s)", p, re.String(), rw.responseHandler.config.Mode)
			rw.responseHandler.ServeBlockedRequest(w, req)
			return
		}
	}

	// 3. Optional: Check Query String if enabled
	if rw.checkQuery && req.URL.RawQuery != "" {
		unescapedQuery, err := url.QueryUnescape(req.URL.RawQuery)
		if err != nil {
			unescapedQuery = req.URL.RawQuery
		}

		if re := rw.findMatchingBlock(unescapedQuery); re != nil {
			rw.logDebug("unescaped query %q blocked by pattern %q", unescapedQuery, re.String())
			rw.responseHandler.ServeBlockedRequest(w, req)
			return
		}
		if re := rw.findMatchingBlock(req.URL.RawQuery); re != nil {
			rw.logDebug("raw query %q blocked by pattern %q", req.URL.RawQuery, re.String())
			rw.responseHandler.ServeBlockedRequest(w, req)
			return
		}

		queryParams := req.URL.Query()
		for key, values := range queryParams {
			for _, val := range values {
				if re := rw.findMatchingBlock(val); re != nil {
					rw.logDebug("query param %q with value %q blocked by pattern %q", key, val, re.String())
					rw.responseHandler.ServeBlockedRequest(w, req)
					return
				}
			}
		}
	}

	rw.logDebug("request %s %s passed inspection", req.Method, req.URL.Path)
	rw.next.ServeHTTP(w, req)
}

func (rw *RouteWarden) findMatchingAllow(target string) *regexp.Regexp {
	for _, re := range rw.allowRegexes {
		if re.MatchString(target) {
			return re
		}
	}
	return nil
}

func (rw *RouteWarden) findMatchingBlock(target string) *regexp.Regexp {
	for _, re := range rw.blockRegexes {
		if re.MatchString(target) {
			return re
		}
	}
	return nil
}
