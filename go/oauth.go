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

// GoogleTokenEndpoint is Google's OAuth 2.0 token endpoint.
const GoogleTokenEndpoint = "https://oauth2.googleapis.com/token"

// TokenResponse is a successful response from the OAuth 2.0 token endpoint.
//
// RefreshToken is only populated on flows where Google returns a new one (it is
// usually absent on a plain refresh, in which case the original refresh token
// remains valid).
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// tokenErrorResponse mirrors the OAuth error envelope.
type tokenErrorResponse struct {
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// OAuthClient performs the server-side OAuth 2.0 token-refresh exchange.
//
// At minimum, ClientID and ClientSecret are required. Source them from the
// environment — never hardcode the secret. Callers must invoke Close (or use
// defer) so idle TCP connections are released.
type OAuthClient struct {
	clientID      string
	clientSecret  string
	tokenEndpoint string
	httpClient    *http.Client
	ownsHTTP      bool
}

// OAuthOptions configures NewOAuthClient.
type OAuthOptions struct {
	// ClientID is the OAuth 2.0 client identifier issued by Google Cloud.
	ClientID string
	// ClientSecret is the OAuth 2.0 client secret. Pass from an env var.
	ClientSecret string
	// TokenEndpoint overrides the token URL (used by tests). Defaults to
	// GoogleTokenEndpoint.
	TokenEndpoint string
	// Timeout sets the per-request total timeout. Defaults to DefaultTimeout.
	Timeout time.Duration
	// HTTPClient lets callers inject a fully configured *http.Client. When nil,
	// NewOAuthClient builds one with bounded idle connections and an explicit
	// timeout. The package never falls back to http.DefaultClient.
	HTTPClient *http.Client
}

// NewOAuthClient constructs an *OAuthClient. It returns an error if either
// credential is missing.
func NewOAuthClient(opts OAuthOptions) (*OAuthClient, error) {
	if opts.ClientID == "" || opts.ClientSecret == "" {
		return nil, errors.New("youtube: OAuthClient requires both ClientID and ClientSecret")
	}
	endpoint := opts.TokenEndpoint
	if endpoint == "" {
		endpoint = GoogleTokenEndpoint
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	c := &OAuthClient{
		clientID:      opts.ClientID,
		clientSecret:  opts.ClientSecret,
		tokenEndpoint: endpoint,
	}
	if opts.HTTPClient != nil {
		c.httpClient = opts.HTTPClient
		c.ownsHTTP = false
	} else {
		c.httpClient = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:          10,
				MaxIdleConnsPerHost:   5,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
		c.ownsHTTP = true
	}
	return c, nil
}

// Close releases idle TCP connections owned by the client.
func (c *OAuthClient) Close() error {
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
func (c *OAuthClient) HTTPClient() *http.Client { return c.httpClient }

// RefreshAccessToken exchanges a refresh token for a fresh access token.
func (c *OAuthClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, errors.New("youtube: RefreshAccessToken requires a refresh token")
	}
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", "refresh_token")
	return c.postToken(ctx, form)
}

// ExchangeAuthorizationCode exchanges an authorization code for access +
// refresh tokens. The interactive consent step that produces code is the
// frontend's responsibility; this completes the exchange server-side.
//
// codeVerifier may be empty when PKCE was not used.
func (c *OAuthClient) ExchangeAuthorizationCode(ctx context.Context, code, redirectURI, codeVerifier string) (*TokenResponse, error) {
	if code == "" {
		return nil, errors.New("youtube: ExchangeAuthorizationCode requires a code")
	}
	if redirectURI == "" {
		return nil, errors.New("youtube: ExchangeAuthorizationCode requires a redirect URI")
	}
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")
	if codeVerifier != "" {
		form.Set("code_verifier", codeVerifier)
	}
	return c.postToken(ctx, form)
}

// postToken sends a form-encoded token request and parses the response.
func (c *OAuthClient) postToken(ctx context.Context, form url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("youtube: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube: token request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("youtube: read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var te tokenErrorResponse
		_ = json.Unmarshal(raw, &te)
		reason := te.ErrorCode
		if reason == "" {
			reason = "invalid_request"
		}
		message := te.ErrorDescription
		if message == "" {
			message = reason
		}
		return nil, &Error{
			StatusCode: resp.StatusCode,
			Code:       resp.StatusCode,
			Status:     "OAUTH_ERROR",
			Message:    message,
			Reasons:    []string{reason},
			Body:       raw,
		}
	}

	var token TokenResponse
	if err := json.Unmarshal(raw, &token); err != nil {
		return nil, &Error{
			StatusCode: resp.StatusCode,
			Status:     "INVALID_RESPONSE",
			Message:    fmt.Sprintf("token endpoint returned non-JSON: %v", err),
			Body:       raw,
		}
	}
	return &token, nil
}
