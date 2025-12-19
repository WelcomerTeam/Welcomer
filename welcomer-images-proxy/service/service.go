package service

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
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
}

type proxyResponse struct {
	contentType string
	body        []byte
	statusCode  int
}
type cacheEntry struct {
	proxyResponse

	expiresAt time.Time
	permanent bool
}

type ProxyServiceOptions struct {
	Debug bool
	Host  string
}

func NewProxyService(options ProxyServiceOptions) (ps *ProxyService, err error) {
	ps = &ProxyService{
		Options: options,

		client: &http.Client{Timeout: 10 * time.Second},

		cacheMu: sync.RWMutex{},
		cache:   make(map[string]*cacheEntry),

		sf: singleflight.Group{},
	}

	return ps, nil
}

func (ps *ProxyService) Open() {
	ps.StartTime = time.Now()

	welcomer.Logger.Info().Msgf("Starting image proxy service. Version %s", VERSION)

	ps.server = &http.Server{
		WriteTimeout:      time.Second * 10,
		ReadTimeout:       time.Second * 10,
		ReadHeaderTimeout: time.Second * 10,
		IdleTimeout:       time.Second * 10,
		Addr:              ps.Options.Host,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

				return
			}

			// read host from url query parameter
			host := r.URL.Query().Get("url")

			requestURL, err := url.Parse(host)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msgf("invalid url attempted: %s", r.Host)
				http.Error(w, "Bad request", http.StatusBadRequest)

				return
			}

			// Disallow local or IP connections
			if !welcomer.IsValidHostname(requestURL.Host) {
				welcomer.Logger.Error().Msgf("invalid hostname attempted: %s", r.Host)
				http.Error(w, "Forbidden", http.StatusForbidden)

				return
			}

			key := requestURL.String()

			// Try cache
			if resp := ps.cachedResponse(r, key); resp != nil {
				welcomer.Logger.Debug().Msgf("cache hit for %s", key)

				w.Header().Set("Content-Type", resp.contentType)
				w.WriteHeader(resp.statusCode)
				w.Write(resp.body)

				return
			}

			// Coalesce concurrent fetches for the same URL
			res, err, _ := ps.sf.Do(key, func() (any, error) {
				// Double-check cache inside the singleflight
				if resp := ps.cachedResponse(r, key); resp != nil {
					return resp, nil
				}

				return ps.fetchAndStore(r.Context(), *requestURL)
			})

			if err != nil {
				// Fallback to normal proxying if fetch failed
				if ps.Options.Debug {
					welcomer.Logger.Warn().Err(err).Msgf("proxy fetch failed for %s, falling back", key)
				}

				http.Error(w, "Bad Gateway", http.StatusBadGateway)

				return
			}

			proxyResponse, _ := res.(proxyResponse)

			w.Header().Set("Content-Type", proxyResponse.contentType)
			w.WriteHeader(proxyResponse.statusCode)
			w.Write(proxyResponse.body)
		}),
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
func (ps *ProxyService) cachedResponse(req *http.Request, key string) *cacheEntry {
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

	return entry
}

// fetch fetches the URL and returns a response.
func (ps *ProxyService) fetch(ctx context.Context, requestURL url.URL) (*http.Response, error) {
	// Build request and forward common headers so origin can respond correctly
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := ps.client.Do(req)
	if err != nil {
		welcomer.Logger.Warn().Err(err).Msgf("fetch failed for %s", requestURL.String())

		return nil, err
	}

	return res, nil
}

// fetchAndStore fetches the URL, stores it in cache with appropriate TTL, and returns a response.
func (ps *ProxyService) fetchAndStore(ctx context.Context, requestURL url.URL) (proxyResponse, error) {
	res, err := ps.fetch(ctx, requestURL)
	if err != nil {
		return proxyResponse{}, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return proxyResponse{}, err
	}

	contentType := res.Header.Get("Content-Type")

	status := res.StatusCode

	// Only cache successful 2xx responses
	if status < 200 || status >= 300 {
		welcomer.Logger.Warn().Msgf("Not caching non-2xx response for %s: status=%d", requestURL, status)

		return proxyResponse{contentType, body, status}, nil
	}

	entry := &cacheEntry{
		proxyResponse: proxyResponse{
			contentType: contentType,
			body:        body,
			statusCode:  status,
		},
		permanent: isPermanentResource(requestURL, contentType),
	}
	if !entry.permanent {
		entry.expiresAt = time.Now().Add(resourceTTL)
		welcomer.Logger.Debug().Msgf("Caching temporary resource %s for %s", requestURL.String(), resourceTTL)
	} else {
		welcomer.Logger.Info().Msgf("Caching permanent resource %s", requestURL)
	}

	ps.cacheMu.Lock()
	ps.cache[requestURL.String()] = entry
	ps.cacheMu.Unlock()

	// Return a fresh response built from the cached entry
	return proxyResponse{entry.contentType, entry.body, entry.statusCode}, nil
}

func isPermanentResource(u url.URL, contentType string) bool {
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
