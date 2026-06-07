package youtube

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newUploadTestClient wires both the data API and the upload API at the test
// server, so a single handler can serve init + chunk + write requests.
func newUploadTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New(ClientOptions{
		AccessToken:      "TKN",
		BaseURL:          srv.URL,
		UploadBaseURL:    srv.URL + "/upload",
		AnalyticsBaseURL: srv.URL + "/analytics",
		Timeout:          5 * time.Second,
	})
	t.Cleanup(func() { _ = c.Close() })
	return c, srv
}

func TestInsertVideo_ResumableMultiChunk(t *testing.T) {
	const total = 10 // 10 bytes, 4-byte chunks => 4,4,2
	uploadPath := "/upload/videos"
	sessionPath := "/session/upload-here"

	var initCalled bool
	var ranges []string

	c, srv := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == uploadPath:
			initCalled = true
			if r.URL.Query().Get("uploadType") != "resumable" {
				t.Errorf("uploadType: %s", r.URL.Query().Get("uploadType"))
			}
			if r.Header.Get("X-Upload-Content-Length") != fmt.Sprint(total) {
				t.Errorf("X-Upload-Content-Length: %s", r.Header.Get("X-Upload-Content-Length"))
			}
			if r.Header.Get("Authorization") != "Bearer TKN" {
				t.Errorf("auth: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Location", srvURL(r)+sessionPath)
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPut && r.URL.Path == sessionPath:
			cr := r.Header.Get("Content-Range")
			ranges = append(ranges, cr)
			body, _ := io.ReadAll(r.Body)
			// Final chunk ends at total-1.
			if strings.HasSuffix(cr, fmt.Sprintf("/%d", total)) && strings.Contains(cr, fmt.Sprintf("-%d/", total-1)) {
				_, _ = io.WriteString(w, `{"id":"VID123","snippet":{"title":"Uploaded"}}`)
				return
			}
			_ = body
			w.WriteHeader(308)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	_ = srv

	media := bytes.NewReader([]byte("0123456789"))
	vid, err := c.InsertVideo(context.Background(), UploadVideoParams{
		Metadata: VideoResource{
			Snippet: &VideoSnippetInput{Title: "Uploaded", CategoryID: "22"},
			Status:  &VideoStatusInput{PrivacyStatus: "private"},
		},
		Media:       media,
		MediaSize:   total,
		ContentType: "video/mp4",
		ChunkSize:   4,
	})
	if err != nil {
		t.Fatalf("InsertVideo: %v", err)
	}
	if !initCalled {
		t.Fatal("init session was not called")
	}
	if vid.ID != "VID123" {
		t.Fatalf("unexpected video: %+v", vid)
	}
	if len(ranges) != 3 {
		t.Fatalf("expected 3 chunks, got %d (%v)", len(ranges), ranges)
	}
	if ranges[0] != "bytes 0-3/10" || ranges[1] != "bytes 4-7/10" || ranges[2] != "bytes 8-9/10" {
		t.Fatalf("unexpected content ranges: %v", ranges)
	}
}

func TestInsertVideo_RequiresOAuth(t *testing.T) {
	c := New(ClientOptions{APIKey: "KEY", Timeout: time.Second})
	defer c.Close()
	_, err := c.InsertVideo(context.Background(), UploadVideoParams{
		Media:     bytes.NewReader([]byte("x")),
		MediaSize: 1,
	})
	if err == nil || !strings.Contains(err.Error(), "OAuth") {
		t.Fatalf("expected OAuth requirement error, got %v", err)
	}
}

func TestInsertVideo_Validation(t *testing.T) {
	c := New(ClientOptions{AccessToken: "TKN", Timeout: time.Second})
	defer c.Close()
	if _, err := c.InsertVideo(context.Background(), UploadVideoParams{MediaSize: 1}); err == nil {
		t.Fatal("expected error for nil Media")
	}
	if _, err := c.InsertVideo(context.Background(), UploadVideoParams{Media: bytes.NewReader([]byte("x"))}); err == nil {
		t.Fatal("expected error for zero MediaSize")
	}
}

func TestUpdateVideo(t *testing.T) {
	c, _ := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/videos" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("part") == "" {
			t.Error("missing part param")
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"id":"VID123"`) {
			t.Errorf("body missing id: %s", body)
		}
		_, _ = io.WriteString(w, `{"id":"VID123","snippet":{"title":"New Title"}}`)
	})
	vid, err := c.UpdateVideo(context.Background(), VideoResource{
		ID:      "VID123",
		Snippet: &VideoSnippetInput{Title: "New Title"},
	}, nil)
	if err != nil {
		t.Fatalf("UpdateVideo: %v", err)
	}
	if vid.Snippet == nil || vid.Snippet.Title != "New Title" {
		t.Fatalf("unexpected video: %+v", vid)
	}
}

func TestUpdateVideo_RequiresID(t *testing.T) {
	c := New(ClientOptions{AccessToken: "TKN", Timeout: time.Second})
	defer c.Close()
	if _, err := c.UpdateVideo(context.Background(), VideoResource{}, nil); err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestDeleteVideo(t *testing.T) {
	c, _ := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/videos" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("id") != "VID123" {
			t.Errorf("id: %s", r.URL.Query().Get("id"))
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteVideo(context.Background(), "VID123"); err != nil {
		t.Fatalf("DeleteVideo: %v", err)
	}
}

func TestSetThumbnail(t *testing.T) {
	c, _ := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/upload/thumbnails/set" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("videoId") != "VID123" {
			t.Errorf("videoId: %s", r.URL.Query().Get("videoId"))
		}
		if r.Header.Get("Content-Type") != "image/png" {
			t.Errorf("content-type: %s", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "PNGDATA" {
			t.Errorf("body: %s", body)
		}
		_, _ = io.WriteString(w, `{"kind":"youtube#thumbnailSetResponse","items":[{"default":{"url":"https://x/t.png","width":120,"height":90}}]}`)
	})
	resp, err := c.SetThumbnail(context.Background(), SetThumbnailParams{
		VideoID:     "VID123",
		Image:       []byte("PNGDATA"),
		ContentType: "image/png",
	})
	if err != nil {
		t.Fatalf("SetThumbnail: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Default == nil || resp.Items[0].Default.Width != 120 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestSetThumbnail_RequiresOAuth(t *testing.T) {
	c := New(ClientOptions{APIKey: "KEY", Timeout: time.Second})
	defer c.Close()
	_, err := c.SetThumbnail(context.Background(), SetThumbnailParams{VideoID: "v", Image: []byte("x")})
	if err == nil || !strings.Contains(err.Error(), "OAuth") {
		t.Fatalf("expected OAuth requirement, got %v", err)
	}
}

func TestSearch(t *testing.T) {
	c, _ := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("q") != "go tutorial" {
			t.Errorf("q: %s", q.Get("q"))
		}
		if q.Get("type") != "video,channel" {
			t.Errorf("type: %s", q.Get("type"))
		}
		if q.Get("maxResults") != "10" {
			t.Errorf("maxResults: %s", q.Get("maxResults"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":{"kind":"youtube#video","videoId":"V1"},"snippet":{"title":"Go"}}],"nextPageToken":"NEXT"}`)
	})
	page, err := c.Search(context.Background(), SearchParams{
		Query:      "go tutorial",
		Types:      []string{"video", "channel"},
		MaxResults: 10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID.VideoID != "V1" {
		t.Fatalf("unexpected items: %+v", page.Items)
	}
	if page.NextPageToken != "NEXT" {
		t.Fatal("expected nextPageToken")
	}
}

func TestSearch_WithAPIKey(t *testing.T) {
	c, _ := func() (*Client, *httptest.Server) {
		var srv *httptest.Server
		srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("key") != "MYKEY" {
				t.Errorf("expected key=MYKEY, got %s", r.URL.Query().Get("key"))
			}
			_, _ = io.WriteString(w, `{"items":[]}`)
		}))
		client := New(ClientOptions{APIKey: "MYKEY", BaseURL: srv.URL, Timeout: 5 * time.Second})
		t.Cleanup(srv.Close)
		t.Cleanup(func() { _ = client.Close() })
		return client, srv
	}()
	if _, err := c.Search(context.Background(), SearchParams{Query: "x"}); err != nil {
		t.Fatalf("Search with API key: %v", err)
	}
}

func TestAnalyticsQuery(t *testing.T) {
	c, _ := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/analytics/reports" {
			t.Errorf("path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("ids") != "channel==MINE" {
			t.Errorf("ids: %s", q.Get("ids"))
		}
		if q.Get("metrics") != "views,likes" {
			t.Errorf("metrics: %s", q.Get("metrics"))
		}
		if q.Get("dimensions") != "day" {
			t.Errorf("dimensions: %s", q.Get("dimensions"))
		}
		if r.Header.Get("Authorization") != "Bearer TKN" {
			t.Errorf("auth: %s", r.Header.Get("Authorization"))
		}
		_, _ = io.WriteString(w, `{"kind":"youtubeAnalytics#resultTable","columnHeaders":[{"name":"day","columnType":"DIMENSION","dataType":"STRING"},{"name":"views","columnType":"METRIC","dataType":"INTEGER"}],"rows":[["2025-01-01",42]]}`)
	})
	rep, err := c.AnalyticsQuery(context.Background(), AnalyticsQueryParams{
		StartDate:  "2025-01-01",
		EndDate:    "2025-01-31",
		Metrics:    []string{"views", "likes"},
		Dimensions: []string{"day"},
	})
	if err != nil {
		t.Fatalf("AnalyticsQuery: %v", err)
	}
	if len(rep.ColumnHeaders) != 2 || len(rep.Rows) != 1 || len(rep.Rows[0]) != 2 {
		t.Fatalf("unexpected report: %+v", rep)
	}
}

func TestAnalyticsQuery_Validation(t *testing.T) {
	c := New(ClientOptions{AccessToken: "TKN", Timeout: time.Second})
	defer c.Close()
	if _, err := c.AnalyticsQuery(context.Background(), AnalyticsQueryParams{Metrics: []string{"views"}}); err == nil {
		t.Fatal("expected error for missing dates")
	}
	if _, err := c.AnalyticsQuery(context.Background(), AnalyticsQueryParams{StartDate: "a", EndDate: "b"}); err == nil {
		t.Fatal("expected error for missing metrics")
	}
}

func TestAnalyticsQuery_RequiresOAuth(t *testing.T) {
	c := New(ClientOptions{APIKey: "KEY", Timeout: time.Second})
	defer c.Close()
	_, err := c.AnalyticsQuery(context.Background(), AnalyticsQueryParams{
		StartDate: "2025-01-01", EndDate: "2025-01-31", Metrics: []string{"views"},
	})
	if err == nil || !strings.Contains(err.Error(), "OAuth") {
		t.Fatalf("expected OAuth requirement, got %v", err)
	}
}

// srvURL reconstructs the absolute base URL of the test server from a request.
func srvURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
