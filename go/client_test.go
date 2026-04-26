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

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New(ClientOptions{
		AccessToken: "TKN",
		BaseURL:     srv.URL,
		Timeout:     5 * time.Second,
	})
	t.Cleanup(func() { _ = c.Close() })
	return c, srv
}

func TestNew_DefaultsHTTPClient(t *testing.T) {
	c := New(ClientOptions{AccessToken: "x"})
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

func TestGetChannel_RequiresFilter(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.GetChannel(context.Background(), GetChannelParams{}); err == nil {
		t.Fatal("expected error when no filter is set")
	}
}

func TestGetChannel_Mine(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("mine") != "true" {
			t.Errorf("mine: %s", r.URL.Query().Get("mine"))
		}
		if r.Header.Get("Authorization") != "Bearer TKN" {
			t.Errorf("auth header: %s", r.Header.Get("Authorization"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"UC123","snippet":{"title":"Demo"},"statistics":{"viewCount":"42"}}],"pageInfo":{"totalResults":1,"resultsPerPage":1}}`)
	})
	ch, err := c.GetChannel(context.Background(), GetChannelParams{Mine: true})
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if ch.ID != "UC123" || ch.Snippet.Title != "Demo" || ch.Statistics.ViewCount != "42" {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}

func TestGetChannel_NotFound(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"items":[],"pageInfo":{"totalResults":0,"resultsPerPage":0}}`)
	})
	_, err := c.GetChannel(context.Background(), GetChannelParams{ID: "missing"})
	apiErr, ok := AsError(err)
	if !ok || !apiErr.IsNotFound() {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestListChannelVideos_PassesPagination(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/playlistItems" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("playlistId") != "UU123" {
			t.Errorf("playlistId: %s", r.URL.Query().Get("playlistId"))
		}
		if r.URL.Query().Get("maxResults") != "10" {
			t.Errorf("maxResults: %s", r.URL.Query().Get("maxResults"))
		}
		if r.URL.Query().Get("pageToken") != "P" {
			t.Errorf("pageToken: %s", r.URL.Query().Get("pageToken"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"PI1","snippet":{"resourceId":{"videoId":"V1"}}}],"nextPageToken":"NEXT"}`)
	})
	page, err := c.ListChannelVideos(context.Background(), ListChannelVideosParams{
		UploadsPlaylistID: "UU123", MaxResults: 10, PageToken: "P",
	})
	if err != nil {
		t.Fatalf("ListChannelVideos: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].Snippet.ResourceID.VideoID != "V1" {
		t.Fatalf("unexpected items: %+v", page.Items)
	}
	if page.NextPageToken != "NEXT" {
		t.Fatalf("expected nextPageToken")
	}
}

func TestGetVideo_AuthError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"error":{"code":403,"message":"forbidden","status":"PERMISSION_DENIED","errors":[{"reason":"forbidden","message":"forbidden"}]}}`)
	})
	_, err := c.GetVideo(context.Background(), "v1", nil)
	apiErr, ok := AsError(err)
	if !ok || !apiErr.IsAuthError() {
		t.Fatalf("expected IsAuthError, got %v", err)
	}
	if apiErr.Status != "PERMISSION_DENIED" {
		t.Fatalf("expected PERMISSION_DENIED, got %q", apiErr.Status)
	}
}

func TestGetVideo_QuotaExceeded(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"error":{"code":403,"message":"quota","status":"PERMISSION_DENIED","errors":[{"reason":"quotaExceeded"}]}}`)
	})
	_, err := c.GetVideo(context.Background(), "v1", nil)
	apiErr, _ := AsError(err)
	if !apiErr.IsRateLimited() {
		t.Fatalf("expected IsRateLimited, got %+v", apiErr)
	}
}

func TestAPIKeyFallback(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Query().Get("key") != "MYKEY" {
			t.Errorf("expected key=MYKEY, got %s", r.URL.Query().Get("key"))
		}
		if r.Header.Get("Authorization") != "" {
			t.Errorf("must not send Authorization when only API key is set: %s", r.Header.Get("Authorization"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"v1"}]}`)
	}))
	defer srv.Close()
	c := New(ClientOptions{APIKey: "MYKEY", BaseURL: srv.URL, Timeout: 5 * time.Second})
	defer c.Close()
	if _, err := c.GetVideo(context.Background(), "v1", nil); err != nil {
		t.Fatalf("GetVideo with API key: %v", err)
	}
	if !called {
		t.Fatal("server was not called")
	}
}

func TestRequiresCredentials(t *testing.T) {
	c := New(ClientOptions{})
	defer c.Close()
	_, err := c.GetVideo(context.Background(), "v1", nil)
	if err == nil || !strings.Contains(err.Error(), "AccessToken") {
		t.Fatalf("expected credential error, got %v", err)
	}
}
