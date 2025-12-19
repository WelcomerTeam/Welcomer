package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/elazarl/goproxy"
	"golang.org/x/sync/singleflight"

	_ "embed"
)

// VERSION follows semantic versioning.
const VERSION = "0.0.1"

const (
	resourceTTL = 30 * time.Second
)

type ProxyService struct {
	StartTime time.Time

	Options ProxyServiceOptions

	// HTTP client and in-memory cache
	client *http.Client

	cacheMu sync.RWMutex
	cache   map[string]*cacheEntry

	sf singleflight.Group

	server *http.Server
	proxy  *goproxy.ProxyHttpServer
}

type cacheEntry struct {
	contentType string
	body        []byte
	statusCode  int
	expiresAt   time.Time
	permanent   bool
}

type ProxyServiceOptions struct {
	Debug bool
	Host  string
}

type ProxyResponse struct {
	contentType string
	body        []byte
	statusCode  int
}

func NewProxyService(options ProxyServiceOptions) (ps *ProxyService, err error) {
	ps = &ProxyService{
		Options: options,

		cacheMu: sync.RWMutex{},
		cache:   make(map[string]*cacheEntry),

		sf:     singleflight.Group{},
		server: &http.Server{},
	}

	// Allow optional custom CA cert+key for goproxy (useful when distributing a fixed CA to clients)
	if certPath := os.Getenv("PROXY_CA_CERT"); certPath != "" {
		keyPath := os.Getenv("PROXY_CA_KEY")
		if keyPath == "" {
			welcomer.Logger.Warn().Msg("PROXY_CA_CERT set but PROXY_CA_KEY is not set; skipping custom CA load")
		} else {
			cert, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				welcomer.Logger.Error().Err(err).Msgf("failed to load PROXY_CA_CERT/PROXY_CA_KEY from %s/%s", certPath, keyPath)
			} else {
				if cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0]); err != nil {
					welcomer.Logger.Warn().Err(err).Msg("failed to parse certificate leaf for loaded CA")
				}
				// Set goproxy's CA so generated MITM certs are signed by this CA
				goproxy.GoproxyCa = cert

				welcomer.Logger.Info().Msgf("Loaded goproxy CA from %s and %s", certPath, keyPath)
			}
		}
	} else {
		welcomer.Logger.Debug().Msg("No PROXY_CA_CERT provided, using built-in goproxy CA")
	}

	// Client transport: keep default unless PROXY_ROOT_CA is set (left intact for TLS-to-upstream trust)
	ps.client = &http.Client{Timeout: 10 * time.Second}

	return ps, nil
}

func (ps *ProxyService) Open() {
	ps.StartTime = time.Now()
	welcomer.Logger.Info().Msgf("Starting image proxy service. Version %s", VERSION)

	ps.proxy = goproxy.NewProxyHttpServer()
	ps.proxy.Verbose = ps.Options.Debug

	ps.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// Expose the goproxy CA so clients can download and trust it if needed.
	ps.proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		welcomer.Logger.Debug().Msgf("proxy request: %s", r.URL.String())

		// Only cache GETs
		if r.Method != http.MethodGet {
			welcomer.Logger.Debug().Msgf("not a GET: %s", r.Method)

			return r, nil
		}

		// Disallow local or IP connections
		if !welcomer.IsValidHostname(r.Host) {
			welcomer.Logger.Warn().Msgf("invalid hostname attempted: %s", r.Host)

			return r, nil
		}

		key := r.URL.String()

		// Try cache
		if resp := ps.cachedResponse(r, key); resp != nil {
			welcomer.Logger.Debug().Msgf("cache hit for %s", key)

			return nil, resp
		}

		// Coalesce concurrent fetches for the same URL
		res, err, _ := ps.sf.Do(key, func() (any, error) {
			// Double-check cache inside the singleflight
			if resp := ps.cachedResponse(r, key); resp != nil {
				return resp, nil
			}

			return ps.fetchAndStore(r.Context(), r)
		})

		if err != nil {
			// Fallback to normal proxying if fetch failed
			if ps.Options.Debug {
				welcomer.Logger.Warn().Err(err).Msgf("proxy fetch failed for %s, falling back", key)
			}

			return r, nil
		}

		proxyResponse, _ := res.(ProxyResponse)

		return nil, &http.Response{
			StatusCode: proxyResponse.statusCode,
			Header:     http.Header{"Content-Type": []string{proxyResponse.contentType}},
			Body:       io.NopCloser(io.LimitReader(io.NewSectionReader(bytes.NewReader(proxyResponse.body), 0, int64(len(proxyResponse.body))), int64(len(proxyResponse.body)))),
			Request:    r,
		}
	})

	ps.server = &http.Server{
		Addr:    ps.Options.Host,
		Handler: ps.proxy,
	}

	go func() {
		if err := ps.server.ListenAndServe(); err != nil {
			welcomer.Logger.Panic().Str("host", ps.Options.Host).Err(err).Msg("Faied to serve proxy")
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			ps.evictCache()
		}
	}()
}

func (ps *ProxyService) Close() {
	ps.server.Shutdown(context.Background())
}

// evictCache clears expired entries from the cache.
func (ps *ProxyService) evictCache() {
	ps.cacheMu.Lock()
	defer ps.cacheMu.Unlock()

	now := time.Now()
	for key, entry := range ps.cache {
		if !entry.permanent && now.After(entry.expiresAt) {
			delete(ps.cache, key)
		}
	}
}

// cachedResponse returns a cached http.Response if present and valid.
func (ps *ProxyService) cachedResponse(req *http.Request, key string) *http.Response {
	ps.cacheMu.RLock()
	entry, ok := ps.cache[key]
	ps.cacheMu.RUnlock()

	if !ok {
		return nil
	}

	if !entry.permanent && time.Now().After(entry.expiresAt) {
		// expired, evict lazily
		ps.cacheMu.Lock()
		delete(ps.cache, key)
		ps.cacheMu.Unlock()

		return nil
	}

	return &http.Response{
		StatusCode: entry.statusCode,
		Header:     http.Header{"Content-Type": []string{entry.contentType}},
		Body:       io.NopCloser(io.LimitReader(io.NewSectionReader(bytes.NewReader(entry.body), 0, int64(len(entry.body))), int64(len(entry.body)))),
		Request:    req,
	}
}

// fetch fetches the URL and returns a response.
func (ps *ProxyService) fetch(ctx context.Context, orig *http.Request) (*http.Response, error) {
	// Build request and forward common headers so origin can respond correctly
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, orig.URL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Forward a few headers from the original request
	if ua := orig.Header.Get("User-Agent"); ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	if accept := orig.Header.Get("Accept"); accept != "" {
		req.Header.Set("Accept", accept)
	}
	if ae := orig.Header.Get("Accept-Encoding"); ae != "" {
		req.Header.Set("Accept-Encoding", ae)
	}
	if ref := orig.Header.Get("Referer"); ref != "" {
		req.Header.Set("Referer", ref)
	}

	res, err := ps.client.Do(req)
	if err != nil {
		welcomer.Logger.Warn().Err(err).Msgf("fetch failed for %s", orig.URL.String())
		return nil, err
	}

	return res, nil
}

// fetchAndStore fetches the URL, stores it in cache with appropriate TTL, and returns a response.
func (ps *ProxyService) fetchAndStore(ctx context.Context, orig *http.Request) (ProxyResponse, error) {
	res, err := ps.fetch(ctx, orig)
	if err != nil {
		return ProxyResponse{}, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ProxyResponse{}, err
	}

	ct := res.Header.Get("Content-Type")
	status := res.StatusCode

	// Only cache successful 2xx responses
	if status < 200 || status >= 300 {
		welcomer.Logger.Warn().Msgf("Not caching non-2xx response for %s: status=%d", orig.URL.String(), status)
		return ProxyResponse{ct, body, status}, nil
	}

	entry := &cacheEntry{
		contentType: ct,
		body:        body,
		statusCode:  status,
		permanent:   isPermanentResource(orig.URL, ct),
	}
	if !entry.permanent {
		entry.expiresAt = time.Now().Add(resourceTTL)
		welcomer.Logger.Debug().Msgf("Caching temporary resource %s for %s", orig.URL.String(), resourceTTL)
	} else {
		welcomer.Logger.Info().Msgf("Caching permanent resource %s", orig.URL.String())
	}

	ps.cacheMu.Lock()
	ps.cache[orig.URL.String()] = entry
	ps.cacheMu.Unlock()

	// Return a fresh response built from the cached entry
	return ProxyResponse{entry.contentType, entry.body, entry.statusCode}, nil
}

func isPermanentResource(u *url.URL, contentType string) bool {
	host := strings.ToLower(u.Host)
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	if host == "fonts.googleapis.com" || host == "fonts.gstatic.com" {
		welcomer.Logger.Debug().Msgf("%s is permanent resource", host)
		return true
	}

	welcomer.Logger.Debug().Msgf("%s is not permanent resource", host)
	return false
}
