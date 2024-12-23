package welcomer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/WelcomerTeam/Discord/discord"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TwilightProxy is a proxy that requests are sent through, instead of directly to discord that will handle
// distributed requests and ratelimits automatically. See more at: https://github.com/twilight-rs/http-proxy
type TwilightProxy struct {
	HTTP       *http.Client
	APIVersion string
	URLHost    string
	URLScheme  string
	UserAgent  string

	Debug bool
}

func NewTwilightProxy(url url.URL) discord.RESTInterface {
	return &TwilightProxy{
		HTTP: &http.Client{
			Timeout: 20 * time.Second,
		},
		APIVersion: discord.APIVersion,
		URLHost:    url.Host,
		URLScheme:  url.Scheme,
		UserAgent:  "Sandwich (github.com/WelcomerTeam/Discord)",
	}
}

func (tl *TwilightProxy) Fetch(session *discord.Session, method, endpoint, contentType string, body []byte, headers http.Header) ([]byte, error) {
	if bytes.Contains(body, []byte("nigger")) {
		return nil, fmt.Errorf("very bad request")
	}

	req, err := http.NewRequestWithContext(session.Context, method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	req.URL.Host = tl.URLHost
	req.URL.Scheme = tl.URLScheme

	if strings.Contains(endpoint, "?") {
		req.URL.RawQuery = strings.SplitN(endpoint, "?", 2)[1]
		endpoint = strings.SplitN(endpoint, "?", 2)[0]
	}

	if tl.APIVersion != "" && !strings.HasPrefix(req.URL.Path, "/api") {
		req.URL.Path = "/api/" + tl.APIVersion + endpoint
	}

	for name, values := range headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	if body != nil && len(req.Header.Get("Content-Type")) == 0 {
		req.Header.Set("Content-Type", contentType)
	}

	if session.Token != "" {
		req.Header.Set("Authorization", session.Token)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := tl.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	if tl.Debug {
		if contentType != "application/json" || headers.Get("Content-Disposition") != "" {
			println(method, req.URL.String(), resp.StatusCode, contentType, hex.EncodeToString(body), string(response))
		} else {
			println(method, req.URL.String(), resp.StatusCode, contentType, string(body), string(response))
		}
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
	case http.StatusUnauthorized:
		return response, discord.ErrUnauthorized
	default:
		return response, discord.NewRestError(req, resp, body)
	}

	return response, nil
}

func (tl *TwilightProxy) FetchBJ(session *discord.Session, method, endpoint, contentType string, body []byte, headers http.Header, response interface{}) error {
	resp, err := tl.Fetch(session, method, endpoint, contentType, body, headers)
	if err != nil {
		return err
	}

	if response != nil {
		err = json.Unmarshal(resp, response)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (tl *TwilightProxy) FetchJJ(session *discord.Session, method, endpoint string, payload interface{}, headers http.Header, response interface{}) error {
	var body []byte
	var err error

	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	} else {
		body = make([]byte, 0)
	}

	return tl.FetchBJ(session, method, endpoint, "application/json", body, headers, response)
}

func (tl *TwilightProxy) SetDebug(value bool) {
	tl.Debug = value
}
