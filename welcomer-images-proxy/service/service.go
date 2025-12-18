package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/elazarl/goproxy"
	"golang.org/x/sync/singleflight"
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
}

func NewProxyService(options ProxyServiceOptions) (ps *ProxyService, err error) {
	ps = &ProxyService{
		Options: options,

		cacheMu: sync.RWMutex{},
		cache:   make(map[string]*cacheEntry),

		client: &http.Client{
			Timeout: 10 * time.Second,
		},

		sf:     singleflight.Group{},
		server: &http.Server{},
	}

	return ps, nil
}

func (ps *ProxyService) Open() {
	ps.StartTime = time.Now()
	welcomer.Logger.Info().Msgf("Starting image proxy service. Version %s", VERSION)

	ps.proxy = goproxy.NewProxyHttpServer()
	ps.proxy.Verbose = ps.Options.Debug

	// MITM HTTPS so we can see the real request URL and serve from cache.
	ps.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// Serve from cache or coalesce fetches per URL.
	ps.proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		// Only cache GETs
		if r.Method != http.MethodGet {
			return r, nil
		}

		// Disallow local or IP connections
		if !welcomer.IsValidHostname(r.Host) {
			return r, nil
		}

		key := r.URL.String()

		// Try cache
		if resp := ps.cachedResponse(r, key); resp != nil {
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

		proxyResponse := res.(ProxyResponse)

		return nil, &http.Response{
			StatusCode: http.StatusOK,
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
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{entry.contentType}},
		Body:       io.NopCloser(io.LimitReader(io.NewSectionReader(bytes.NewReader(entry.body), 0, int64(len(entry.body))), int64(len(entry.body)))),
		Request:    req,
	}
}

// fetch fetches the URL and returns a response.
func (ps *ProxyService) fetch(ctx context.Context, orig *http.Request) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, orig.URL.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := ps.client.Do(req)
	if err != nil {
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

	entry := &cacheEntry{
		contentType: ct,
		body:        body,
		permanent:   isPermanentResource(orig.URL, ct),
	}
	if !entry.permanent {
		entry.expiresAt = time.Now().Add(resourceTTL)
	}

	ps.cacheMu.Lock()
	ps.cache[orig.URL.String()] = entry
	ps.cacheMu.Unlock()

	// Return a fresh response built from the cached entry
	return ProxyResponse{entry.contentType, entry.body}, nil
}

func isPermanentResource(url *url.URL, contentType string) bool {
	println(url.Host)

	if url.Host == "fonts.googleapis.com" || url.Host == "fonts.gstatic.com" {
		return true
	}

	return false
}
