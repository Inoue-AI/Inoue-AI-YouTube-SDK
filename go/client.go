package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the production YouTube Data API v3 root.
const DefaultBaseURL = "https://www.googleapis.com/youtube/v3"

// DefaultUploadBaseURL is the production YouTube Data API v3 upload root used
// for resumable + direct media uploads (videos, thumbnails, ...).
const DefaultUploadBaseURL = "https://www.googleapis.com/upload/youtube/v3"

// DefaultTimeout is the per-request timeout used when ClientOptions.Timeout is zero.
const DefaultTimeout = 30 * time.Second

// DefaultUploadChunkSize is the chunk size used for resumable uploads.
// Google requires chunk sizes to be a multiple of 256 KiB; 8 MiB is a safe,
// reliable default.
const DefaultUploadChunkSize = 8 * 1024 * 1024

// ClientOptions configures a *Client created via New.
//
// At least one of AccessToken or APIKey must be supplied. AccessToken takes
// precedence when both are set.
type ClientOptions struct {
	// AccessToken is the OAuth2 access token (required for write operations
	// and for "mine=true" queries).
	AccessToken string

	// APIKey is a Google API key suitable for read-only public endpoints.
	APIKey string

	// BaseURL overrides the production API host (useful for testing).
	BaseURL string

	// UploadBaseURL overrides the upload API host (useful for testing).
	// Defaults to DefaultUploadBaseURL.
	UploadBaseURL string

	// AnalyticsBaseURL overrides the Analytics API host (useful for testing).
	// Defaults to DefaultAnalyticsBaseURL.
	AnalyticsBaseURL string

	// Timeout sets the per-request total timeout. Defaults to DefaultTimeout.
	// Note: resumable uploads issue multiple requests, each bounded by this
	// timeout, so a single large upload may exceed it in aggregate.
	Timeout time.Duration

	// HTTPClient lets callers inject a fully configured *http.Client. When
	// nil, New constructs a client with bounded idle connections and an
	// explicit Timeout. The package never falls back to http.DefaultClient.
	HTTPClient *http.Client

	// UserAgent overrides the User-Agent header.
	UserAgent string
}

// Client is the YouTube Data API client. Safe for concurrent use. Callers
// must invoke Close (or use defer) so idle TCP connections are released.
type Client struct {
	httpClient    *http.Client
	baseURL       string
	uploadBaseURL string
	analyticsBase string
	accessToken   string
	apiKey        string
	userAgent     string
	ownsHTTP      bool
}

// New constructs a *Client.
func New(opts ClientOptions) *Client {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	baseURL := strings.TrimRight(opts.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	uploadBaseURL := strings.TrimRight(opts.UploadBaseURL, "/")
	if uploadBaseURL == "" {
		uploadBaseURL = DefaultUploadBaseURL
	}
	analyticsBase := strings.TrimRight(opts.AnalyticsBaseURL, "/")
	ua := opts.UserAgent
	if ua == "" {
		ua = "inoue-youtube-sdk-go/1"
	}
	c := &Client{
		baseURL:       baseURL,
		uploadBaseURL: uploadBaseURL,
		analyticsBase: analyticsBase,
		accessToken:   opts.AccessToken,
		apiKey:        opts.APIKey,
		userAgent:     ua,
	}
	if opts.HTTPClient != nil {
		c.httpClient = opts.HTTPClient
		c.ownsHTTP = false
	} else {
		c.httpClient = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   20,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ResponseHeaderTimeout: timeout,
			},
		}
		c.ownsHTTP = true
	}
	return c
}

// Close releases idle TCP connections.
func (c *Client) Close() error {
	if c == nil || c.httpClient == nil {
		return nil
	}
	if c.ownsHTTP {
		if t, ok := c.httpClient.Transport.(*http.Transport); ok {
			t.CloseIdleConnections()
		}
	}
	return nil
}

// HTTPClient returns the underlying *http.Client.
func (c *Client) HTTPClient() *http.Client { return c.httpClient }

// errorPayload mirrors the standard googleapis error envelope.
type errorPayload struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
		Errors  []struct {
			Reason  string `json:"reason"`
			Message string `json:"message"`
		} `json:"errors"`
	} `json:"error"`
}

// requireCredentials returns an error when neither an OAuth token nor an API
// key is configured.
func (c *Client) requireCredentials() error {
	if c.accessToken == "" && c.apiKey == "" {
		return errors.New("youtube: AccessToken or APIKey is required for this endpoint")
	}
	return nil
}

// requireOAuth returns an error unless an OAuth access token is configured.
// Write/upload endpoints reject API-key-only auth.
func (c *Client) requireOAuth() error {
	if c.accessToken == "" {
		return errors.New("youtube: this endpoint requires an OAuth AccessToken")
	}
	return nil
}

// applyAuth attaches the bearer token (or appends the API key to the query
// when only a key is configured) and the common request headers.
func (c *Client) applyAuth(req *http.Request, query url.Values) {
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	} else if c.apiKey != "" {
		query.Set("key", c.apiKey)
		req.URL.RawQuery = query.Encode()
	}
	req.Header.Set("User-Agent", c.userAgent)
}

// parseErrorBody converts a googleapis error envelope into a *Error.
func parseErrorBody(statusCode int, raw []byte) *Error {
	var ep errorPayload
	_ = json.Unmarshal(raw, &ep)
	reasons := make([]string, 0, len(ep.Error.Errors))
	for _, item := range ep.Error.Errors {
		reasons = append(reasons, item.Reason)
	}
	return &Error{
		StatusCode: statusCode,
		Code:       ep.Error.Code,
		Status:     ep.Error.Status,
		Message:    ep.Error.Message,
		Reasons:    reasons,
		Body:       raw,
	}
}

// doGet executes a GET request and decodes JSON into out. Either an OAuth
// bearer token or an API key (as the "key" query param) is required.
func (c *Client) doGet(ctx context.Context, path string, query url.Values, out any) error {
	if err := c.requireCredentials(); err != nil {
		return err
	}
	if query == nil {
		query = url.Values{}
	}
	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("youtube: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	c.applyAuth(req, query)

	return c.do(req, http.MethodGet, path, out)
}

// doJSON executes a request with a JSON-encoded body (POST/PUT) against the
// data API and decodes the JSON response into out. Requires OAuth.
func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body, out any) error {
	if err := c.requireOAuth(); err != nil {
		return err
	}
	if query == nil {
		query = url.Values{}
	}
	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("youtube: encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}
	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("youtube: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	c.applyAuth(req, query)

	return c.do(req, method, path, out)
}

// doDelete executes a DELETE request. Requires OAuth. A 2xx (including 204) is
// treated as success.
func (c *Client) doDelete(ctx context.Context, path string, query url.Values) error {
	if err := c.requireOAuth(); err != nil {
		return err
	}
	if query == nil {
		query = url.Values{}
	}
	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		return fmt.Errorf("youtube: build request: %w", err)
	}
	c.applyAuth(req, query)

	return c.do(req, http.MethodDelete, path, nil)
}

// doPostNoContent issues a POST with no request body (query-only) against the
// data API, expecting a 204 No Content. Used by videos.rate,
// comments.setModerationStatus, and watermarks.unset. Requires OAuth.
func (c *Client) doPostNoContent(ctx context.Context, path string, query url.Values) error {
	if err := c.requireOAuth(); err != nil {
		return err
	}
	if query == nil {
		query = url.Values{}
	}
	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, nil)
	if err != nil {
		return fmt.Errorf("youtube: build request: %w", err)
	}
	c.applyAuth(req, query)

	return c.do(req, http.MethodPost, path, nil)
}

// doGetBinary issues a GET against the data API and returns the raw response
// body. Used by captions.download. Either an OAuth token or an API key is
// required.
func (c *Client) doGetBinary(ctx context.Context, path string, query url.Values) ([]byte, error) {
	if err := c.requireCredentials(); err != nil {
		return nil, err
	}
	if query == nil {
		query = url.Values{}
	}
	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("youtube: build request: %w", err)
	}
	c.applyAuth(req, query)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube: %s %s: %w", http.MethodGet, path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("youtube: read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, parseErrorBody(resp.StatusCode, raw)
	}
	return raw, nil
}

// do issues req, reads the body, maps errors, and decodes JSON into out when
// out is non-nil and the body is non-empty. The request context is already
// attached via http.NewRequestWithContext at every call site.
func (c *Client) do(req *http.Request, method, path string, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("youtube: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("youtube: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseErrorBody(resp.StatusCode, raw)
	}

	if out == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return &Error{
			StatusCode: resp.StatusCode,
			Status:     "INVALID_RESPONSE",
			Message:    fmt.Sprintf("YouTube returned non-JSON: %v", err),
			Body:       raw,
		}
	}
	return nil
}
