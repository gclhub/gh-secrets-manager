package api

import (
	"net/http"
	"net/url"
)

// testTransport is a custom http.RoundTripper that ensures all requests go through our test server
type testTransport struct {
	baseURL    *url.URL
	underlying http.RoundTripper
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Make a copy of the request so we don't modify the original
	reqCopy := *req
	
	// If the request is to api.github.com, rewrite it to our test server
	if req.URL.Host == "api.github.com" || req.URL.Hostname() == "api.github.com" {
		reqCopy.URL.Scheme = t.baseURL.Scheme
		reqCopy.URL.Host = t.baseURL.Host
		reqCopy.Host = t.baseURL.Host
	}
	
	// Handle URLs that are already pointing to our test server but might be using absolute paths
	if req.URL.Path[0] == '/' && !req.URL.IsAbs() {
		// Ensure the path is properly handled when it's an absolute path
		reqCopy.URL.Path = req.URL.Path
	}

	return t.underlying.RoundTrip(&reqCopy)
}

// newTestClient creates an HTTP client that redirects all GitHub API requests to our test server
func newTestClient(baseURL *url.URL) *http.Client {
	return &http.Client{
		Transport: &testTransport{
			baseURL:    baseURL,
			underlying: http.DefaultTransport,
		},
	}
}
