package youtube

import (
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

// DefaultTimeout is the per-request timeout used when ClientOptions.Timeout is zero.
const DefaultTimeout = 30 * time.Second

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

	// Timeout sets the per-request total timeout. Defaults to DefaultTimeout.
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
	httpClient  *http.Client
	baseURL     string
	accessToken string
	apiKey      string
	userAgent   string
	ownsHTTP    bool
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
	ua := opts.UserAgent
	if ua == "" {
		ua = "inoue-youtube-sdk-go/1"
	}
	c := &Client{
		baseURL:     baseURL,
		accessToken: opts.AccessToken,
		apiKey:      opts.APIKey,
		userAgent:   ua,
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

// doGet executes a GET request and decodes JSON into out. Either an OAuth
// bearer token or an API key (as the "key" query param) is required.
func (c *Client) doGet(ctx context.Context, path string, query url.Values, out any) error {
	if c.accessToken == "" && c.apiKey == "" {
		return errors.New("youtube: AccessToken or APIKey is required for this endpoint")
	}
	if query == nil {
		query = url.Values{}
	}
	if c.accessToken == "" && c.apiKey != "" {
		query.Set("key", c.apiKey)
	}
	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("youtube: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("youtube: GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("youtube: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var ep errorPayload
		_ = json.Unmarshal(raw, &ep)
		reasons := make([]string, 0, len(ep.Error.Errors))
		for _, item := range ep.Error.Errors {
			reasons = append(reasons, item.Reason)
		}
		return &Error{
			StatusCode: resp.StatusCode,
			Code:       ep.Error.Code,
			Status:     ep.Error.Status,
			Message:    ep.Error.Message,
			Reasons:    reasons,
			Body:       raw,
		}
	}

	if out == nil {
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
