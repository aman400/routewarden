package routewarden

import (
	"context"
	"fmt"
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
		blockRegexes:    compiledBlockRegexes,
		allowRegexes:    compiledAllowRegexes,
		ipFilter:        ipFilter,
		checkQuery:      config.CheckQuery,
		responseHandler: respHandler,
	}, nil
}

func (rw *RouteWarden) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !rw.enabled {
		rw.next.ServeHTTP(w, req)
		return
	}

	// Exempt whitelisted client IPs or CIDR subnets from blocking
	if rw.ipFilter.IsAllowed(req) {
		rw.next.ServeHTTP(w, req)
		return
	}

	// Canonicalize and inspect paths with anti-evasion protections
	candidatePaths := ExtractCandidatePaths(req.URL.RawPath, req.URL.Path, req.RequestURI)

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

		if rw.isBlocked(unescapedQuery) || rw.isBlocked(req.URL.RawQuery) {
			rw.responseHandler.ServeBlockedRequest(w, req)
			return
		}

		queryParams := req.URL.Query()
		for _, values := range queryParams {
			for _, val := range values {
				if rw.isBlocked(val) {
					rw.responseHandler.ServeBlockedRequest(w, req)
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
