package welcomer

import (
	"net/http"
)

type UserAgentSetterTransport struct {
	UserAgent    string
	roundTripper http.RoundTripper
}

func (t *UserAgentSetterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", t.UserAgent)

	return t.roundTripper.RoundTrip(req)
}

// NewUserAgentSetterTransport creates a new UserAgentSetterTransport.
func NewUserAgentSetterTransport(roundTripper http.RoundTripper, userAgent string) *UserAgentSetterTransport {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}

	return &UserAgentSetterTransport{
		UserAgent:    userAgent,
		roundTripper: roundTripper,
	}
}
