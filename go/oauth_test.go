package youtube

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestOAuthClient(t *testing.T, handler http.HandlerFunc) *OAuthClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewOAuthClient(OAuthOptions{
		ClientID:      "cid",
		ClientSecret:  "csecret",
		TokenEndpoint: srv.URL,
		Timeout:       5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewOAuthClient: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestNewOAuthClient_RequiresCredentials(t *testing.T) {
	if _, err := NewOAuthClient(OAuthOptions{ClientID: "", ClientSecret: "s"}); err == nil {
		t.Fatal("expected error for missing ClientID")
	}
	if _, err := NewOAuthClient(OAuthOptions{ClientID: "i", ClientSecret: ""}); err == nil {
		t.Fatal("expected error for missing ClientSecret")
	}
}

func TestNewOAuthClient_DefaultsHTTPClient(t *testing.T) {
	c, err := NewOAuthClient(OAuthOptions{ClientID: "i", ClientSecret: "s"})
	if err != nil {
		t.Fatalf("NewOAuthClient: %v", err)
	}
	defer c.Close()
	if c.HTTPClient() == http.DefaultClient {
		t.Fatal("must NOT use http.DefaultClient")
	}
	if c.HTTPClient().Timeout == 0 {
		t.Fatal("must set explicit Timeout")
	}
	tr, ok := c.HTTPClient().Transport.(*http.Transport)
	if !ok || tr.MaxIdleConnsPerHost == 0 {
		t.Fatal("Transport must configure idle-conn limits")
	}
}

func TestRefreshAccessToken_Success(t *testing.T) {
	var gotForm string
	c := newTestOAuthClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			t.Errorf("content-type: %s", ct)
		}
		body, _ := io.ReadAll(r.Body)
		gotForm = string(body)
		_, _ = io.WriteString(w, `{"access_token":"ya29.NEW","expires_in":3599,"token_type":"Bearer","scope":"s"}`)
	})
	tok, err := c.RefreshAccessToken(context.Background(), "rt-123")
	if err != nil {
		t.Fatalf("RefreshAccessToken: %v", err)
	}
	if tok.AccessToken != "ya29.NEW" || tok.ExpiresIn != 3599 {
		t.Fatalf("unexpected token: %+v", tok)
	}
	for _, want := range []string{"grant_type=refresh_token", "refresh_token=rt-123", "client_id=cid", "client_secret=csecret"} {
		if !strings.Contains(gotForm, want) {
			t.Errorf("form missing %q in %q", want, gotForm)
		}
	}
}

func TestRefreshAccessToken_RequiresToken(t *testing.T) {
	c := newTestOAuthClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.RefreshAccessToken(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty refresh token")
	}
}

func TestRefreshAccessToken_InvalidGrant(t *testing.T) {
	c := newTestOAuthClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":"invalid_grant","error_description":"Token revoked."}`)
	})
	_, err := c.RefreshAccessToken(context.Background(), "revoked")
	apiErr, ok := AsError(err)
	if !ok {
		t.Fatalf("expected *Error, got %v", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("status: %d", apiErr.StatusCode)
	}
	if len(apiErr.Reasons) == 0 || apiErr.Reasons[0] != "invalid_grant" {
		t.Errorf("reasons: %v", apiErr.Reasons)
	}
}

func TestExchangeAuthorizationCode_WithPKCE(t *testing.T) {
	var gotForm string
	c := newTestOAuthClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotForm = string(body)
		_, _ = io.WriteString(w, `{"access_token":"ya29.X","expires_in":3600,"refresh_token":"rt-new"}`)
	})
	tok, err := c.ExchangeAuthorizationCode(context.Background(), "auth-code", "https://app/cb", "verifier")
	if err != nil {
		t.Fatalf("ExchangeAuthorizationCode: %v", err)
	}
	if tok.RefreshToken != "rt-new" {
		t.Fatalf("expected new refresh token, got %+v", tok)
	}
	for _, want := range []string{"grant_type=authorization_code", "code=auth-code", "code_verifier=verifier"} {
		if !strings.Contains(gotForm, want) {
			t.Errorf("form missing %q in %q", want, gotForm)
		}
	}
}

func TestExchangeAuthorizationCode_Validation(t *testing.T) {
	c := newTestOAuthClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.ExchangeAuthorizationCode(context.Background(), "", "https://cb", ""); err == nil {
		t.Fatal("expected error for empty code")
	}
	if _, err := c.ExchangeAuthorizationCode(context.Background(), "code", "", ""); err == nil {
		t.Fatal("expected error for empty redirect URI")
	}
}
