package youtube

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Captions
// ---------------------------------------------------------------------------

func TestListCaptions(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/captions" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("part") != "snippet" {
			t.Errorf("part: %s", q.Get("part"))
		}
		if q.Get("videoId") != "VID1" {
			t.Errorf("videoId: %s", q.Get("videoId"))
		}
		if q.Get("id") != "C1,C2" {
			t.Errorf("id: %s", q.Get("id"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"C1","snippet":{"language":"en","name":"English"}}]}`)
	})
	resp, err := c.ListCaptions(context.Background(), ListCaptionsParams{
		Parts: []string{CaptionPartSnippet}, VideoID: "VID1", IDs: []string{"C1", "C2"},
	})
	if err != nil {
		t.Fatalf("ListCaptions: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Snippet.Language != "en" {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
}

func TestInsertCaption_Multipart(t *testing.T) {
	var gotMeta Caption
	var gotFile string
	c, _ := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/upload/captions" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("uploadType") != "multipart" {
			t.Errorf("uploadType: %s", r.URL.Query().Get("uploadType"))
		}
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
			t.Fatalf("content-type: %s err=%v", r.Header.Get("Content-Type"), err)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("read part: %v", err)
			}
			data, _ := io.ReadAll(part)
			switch part.FormName() {
			case "metadata":
				if part.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
					t.Errorf("metadata content-type: %s", part.Header.Get("Content-Type"))
				}
				if strings.Contains(string(data), "VIDX") {
					gotMeta.Snippet = &CaptionSnippet{VideoID: "VIDX"}
				}
			case "file":
				gotFile = string(data)
				if part.Header.Get("Content-Type") != "text/vtt" {
					t.Errorf("file content-type: %s", part.Header.Get("Content-Type"))
				}
			}
		}
		_, _ = io.WriteString(w, `{"id":"NEWCAP","snippet":{"videoId":"VIDX"}}`)
	})
	out, err := c.InsertCaption(context.Background(), InsertCaptionParams{
		Body:        Caption{Snippet: &CaptionSnippet{VideoID: "VIDX", Language: "en", Name: "t"}},
		CaptionData: []byte("WEBVTT\n"),
		ContentType: "text/vtt",
	})
	if err != nil {
		t.Fatalf("InsertCaption: %v", err)
	}
	if out.ID != "NEWCAP" {
		t.Fatalf("unexpected out: %+v", out)
	}
	if gotMeta.Snippet == nil || gotMeta.Snippet.VideoID != "VIDX" {
		t.Fatalf("metadata part not received")
	}
	if gotFile != "WEBVTT\n" {
		t.Fatalf("file part mismatch: %q", gotFile)
	}
}

func TestUpdateCaption_MetadataOnly(t *testing.T) {
	c, _ := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/captions" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"id":"C1"`) {
			t.Errorf("body missing id: %s", body)
		}
		_, _ = io.WriteString(w, `{"id":"C1","snippet":{"isDraft":false}}`)
	})
	out, err := c.UpdateCaption(context.Background(), UpdateCaptionParams{
		Body: Caption{ID: "C1", Snippet: &CaptionSnippet{IsDraft: boolPtr(false)}},
	})
	if err != nil {
		t.Fatalf("UpdateCaption: %v", err)
	}
	if out.ID != "C1" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestUpdateCaption_WithData_Multipart(t *testing.T) {
	c, _ := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/upload/captions" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("uploadType") != "multipart" {
			t.Errorf("uploadType: %s", r.URL.Query().Get("uploadType"))
		}
		_, _ = io.WriteString(w, `{"id":"C1"}`)
	})
	_, err := c.UpdateCaption(context.Background(), UpdateCaptionParams{
		Body:        Caption{ID: "C1", Snippet: &CaptionSnippet{}},
		CaptionData: []byte("WEBVTT\n"),
		ContentType: "text/vtt",
	})
	if err != nil {
		t.Fatalf("UpdateCaption: %v", err)
	}
}

func TestDownloadCaption(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/captions/C1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("tfmt") != "srt" {
			t.Errorf("tfmt: %s", r.URL.Query().Get("tfmt"))
		}
		if r.URL.Query().Get("tlang") != "es" {
			t.Errorf("tlang: %s", r.URL.Query().Get("tlang"))
		}
		_, _ = io.WriteString(w, "1\n00:00:00,000 --> 00:00:01,000\nHola\n")
	})
	data, err := c.DownloadCaption(context.Background(), DownloadCaptionParams{ID: "C1", Tfmt: "srt", Tlang: "es"})
	if err != nil {
		t.Fatalf("DownloadCaption: %v", err)
	}
	if !strings.Contains(string(data), "Hola") {
		t.Fatalf("unexpected caption: %q", data)
	}
}

func TestDeleteCaption(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/captions" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("id") != "C1" {
			t.Errorf("id: %s", r.URL.Query().Get("id"))
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteCaption(context.Background(), "C1"); err != nil {
		t.Fatalf("DeleteCaption: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Channel banners
// ---------------------------------------------------------------------------

func TestInsertChannelBanner(t *testing.T) {
	c, _ := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/upload/channelBanners/insert" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("uploadType") != "media" {
			t.Errorf("uploadType: %s", r.URL.Query().Get("uploadType"))
		}
		if r.Header.Get("Content-Type") != "image/png" {
			t.Errorf("content-type: %s", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "BANNER" {
			t.Errorf("body: %s", body)
		}
		_, _ = io.WriteString(w, `{"url":"https://x/banner"}`)
	})
	out, err := c.InsertChannelBanner(context.Background(), InsertChannelBannerParams{
		ImageData: []byte("BANNER"), ContentType: "image/png",
	})
	if err != nil {
		t.Fatalf("InsertChannelBanner: %v", err)
	}
	if out.URL != "https://x/banner" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestInsertChannelBanner_RequiresOAuth(t *testing.T) {
	c := New(ClientOptions{APIKey: "K"})
	defer c.Close()
	_, err := c.InsertChannelBanner(context.Background(), InsertChannelBannerParams{ImageData: []byte("x")})
	if err == nil || !strings.Contains(err.Error(), "OAuth") {
		t.Fatalf("expected OAuth requirement, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Watermarks
// ---------------------------------------------------------------------------

func TestSetWatermark(t *testing.T) {
	c, _ := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/upload/watermarks/set" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("channelId") != "UC1" {
			t.Errorf("channelId: %s", r.URL.Query().Get("channelId"))
		}
		if r.URL.Query().Get("uploadType") != "media" {
			t.Errorf("uploadType: %s", r.URL.Query().Get("uploadType"))
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "WM" {
			t.Errorf("body: %s", body)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.SetWatermark(context.Background(), SetWatermarkParams{
		ChannelID: "UC1",
		Watermark: Watermark{Timing: &WatermarkTiming{Type: WatermarkTimingOffsetFromStart}},
		ImageData: []byte("WM"),
		// ContentType defaults to image/png.
	})
	if err != nil {
		t.Fatalf("SetWatermark: %v", err)
	}
}

func TestUnsetWatermark(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/watermarks/unset" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("channelId") != "UC1" {
			t.Errorf("channelId: %s", r.URL.Query().Get("channelId"))
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.UnsetWatermark(context.Background(), "UC1"); err != nil {
		t.Fatalf("UnsetWatermark: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Channel sections
// ---------------------------------------------------------------------------

func TestListChannelSections(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channelSections" {
			t.Errorf("path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("channelId") != "UC1" {
			t.Errorf("channelId: %s", q.Get("channelId"))
		}
		if q.Get("mine") != "true" {
			t.Errorf("mine: %s", q.Get("mine"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"S1","snippet":{"type":"singlePlaylist"}}]}`)
	})
	mine := true
	resp, err := c.ListChannelSections(context.Background(), ListChannelSectionsParams{
		Parts: []string{ChannelSectionPartSnippet}, ChannelID: "UC1", Mine: &mine,
	})
	if err != nil {
		t.Fatalf("ListChannelSections: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Snippet.Type != ChannelSectionTypeSinglePlaylist {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
}

func TestInsertChannelSection_InfersParts(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/channelSections" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("part") != "snippet,contentDetails" {
			t.Errorf("part: %s", r.URL.Query().Get("part"))
		}
		_, _ = io.WriteString(w, `{"id":"S1"}`)
	})
	out, err := c.InsertChannelSection(context.Background(), ChannelSection{
		Snippet:        &ChannelSectionSnippet{Type: ChannelSectionTypeSinglePlaylist},
		ContentDetails: &ChannelSectionContentDetails{Playlists: []string{"PL1"}},
	}, nil)
	if err != nil {
		t.Fatalf("InsertChannelSection: %v", err)
	}
	if out.ID != "S1" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestUpdateChannelSection(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/channelSections" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":"S1"}`)
	})
	if _, err := c.UpdateChannelSection(context.Background(), ChannelSection{
		ID: "S1", Snippet: &ChannelSectionSnippet{Title: "x"},
	}, nil); err != nil {
		t.Fatalf("UpdateChannelSection: %v", err)
	}
}

func TestDeleteChannelSection(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Query().Get("id") != "S1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteChannelSection(context.Background(), "S1"); err != nil {
		t.Fatalf("DeleteChannelSection: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Channels list + update
// ---------------------------------------------------------------------------

func TestListChannels(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/channels" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if q.Get("forHandle") != "@demo" {
			t.Errorf("forHandle: %s", q.Get("forHandle"))
		}
		if q.Get("maxResults") != "5" {
			t.Errorf("maxResults: %s", q.Get("maxResults"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"UC1","snippet":{"title":"Demo"}}],"pageInfo":{"totalResults":1}}`)
	})
	resp, err := c.ListChannels(context.Background(), ListChannelsParams{
		Parts: []string{ChannelPartSnippet}, ForHandle: "@demo",
	})
	if err != nil {
		t.Fatalf("ListChannels: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Snippet.Title != "Demo" {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
}

func TestUpdateChannel_InfersParts(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/channels" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("part") != "brandingSettings" {
			t.Errorf("part: %s", r.URL.Query().Get("part"))
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"bannerExternalUrl":"https://x/banner"`) {
			t.Errorf("body: %s", body)
		}
		_, _ = io.WriteString(w, `{"id":"UC1"}`)
	})
	out, err := c.UpdateChannel(context.Background(), Channel{
		ID: "UC1",
		BrandingSettings: &ChannelBrandingSettings{
			Image: &ChannelBrandingImage{BannerExternalURL: "https://x/banner"},
		},
	}, nil)
	if err != nil {
		t.Fatalf("UpdateChannel: %v", err)
	}
	if out.ID != "UC1" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestUpdateChannel_RequiresID(t *testing.T) {
	c := New(ClientOptions{AccessToken: "T"})
	defer c.Close()
	if _, err := c.UpdateChannel(context.Background(), Channel{}, nil); err == nil {
		t.Fatal("expected error for missing ID")
	}
}

// ---------------------------------------------------------------------------
// Comments
// ---------------------------------------------------------------------------

func TestListComments(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/comments" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if q.Get("parentId") != "P1" {
			t.Errorf("parentId: %s", q.Get("parentId"))
		}
		if q.Get("maxResults") != "20" {
			t.Errorf("maxResults: %s", q.Get("maxResults"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"X1","snippet":{"textOriginal":"hi"}}]}`)
	})
	resp, err := c.ListComments(context.Background(), ListCommentsParams{
		Parts: []string{CommentPartSnippet}, ParentID: "P1",
	})
	if err != nil {
		t.Fatalf("ListComments: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Snippet.TextOriginal != "hi" {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
}

func TestInsertComment(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/comments" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"parentId":"P1"`) {
			t.Errorf("body: %s", body)
		}
		_, _ = io.WriteString(w, `{"id":"R1"}`)
	})
	out, err := c.InsertComment(context.Background(), Comment{
		Snippet: &CommentSnippet{ParentID: "P1", TextOriginal: "reply"},
	}, nil)
	if err != nil {
		t.Fatalf("InsertComment: %v", err)
	}
	if out.ID != "R1" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestUpdateComment(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/comments" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":"R1"}`)
	})
	if _, err := c.UpdateComment(context.Background(), Comment{ID: "R1", Snippet: &CommentSnippet{TextOriginal: "edit"}}, nil); err != nil {
		t.Fatalf("UpdateComment: %v", err)
	}
}

func TestDeleteComment(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Query().Get("id") != "R1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteComment(context.Background(), "R1"); err != nil {
		t.Fatalf("DeleteComment: %v", err)
	}
}

func TestSetCommentModerationStatus(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/comments/setModerationStatus" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("id") != "A,B" {
			t.Errorf("id: %s", q.Get("id"))
		}
		if q.Get("moderationStatus") != "rejected" {
			t.Errorf("moderationStatus: %s", q.Get("moderationStatus"))
		}
		if q.Get("banAuthor") != "true" {
			t.Errorf("banAuthor: %s", q.Get("banAuthor"))
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.SetCommentModerationStatus(context.Background(), SetCommentModerationStatusParams{
		IDs: []string{"A", "B"}, ModerationStatus: ModerationStatusRejected, BanAuthor: true,
	})
	if err != nil {
		t.Fatalf("SetCommentModerationStatus: %v", err)
	}
}

func TestIterReplies_Pagination(t *testing.T) {
	calls := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("pageToken") == "" {
			_, _ = io.WriteString(w, `{"items":[{"id":"R1"}],"nextPageToken":"N"}`)
			return
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"R2"}]}`)
	})
	var ids []string
	err := c.IterReplies(context.Background(), "P1", []string{CommentPartSnippet}, 50, func(cm Comment) error {
		ids = append(ids, cm.ID)
		return nil
	})
	if err != nil {
		t.Fatalf("IterReplies: %v", err)
	}
	if calls != 2 || len(ids) != 2 || ids[0] != "R1" || ids[1] != "R2" {
		t.Fatalf("unexpected calls=%d ids=%v", calls, ids)
	}
}

// ---------------------------------------------------------------------------
// Comment threads
// ---------------------------------------------------------------------------

func TestListCommentThreads(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/commentThreads" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if q.Get("videoId") != "V1" {
			t.Errorf("videoId: %s", q.Get("videoId"))
		}
		if q.Get("order") != "relevance" {
			t.Errorf("order: %s", q.Get("order"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"T1","snippet":{"videoId":"V1"}}]}`)
	})
	resp, err := c.ListCommentThreads(context.Background(), ListCommentThreadsParams{
		Parts: []string{CommentThreadPartSnippet}, VideoID: "V1", Order: CommentThreadOrderRelevance,
	})
	if err != nil {
		t.Fatalf("ListCommentThreads: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Snippet.VideoID != "V1" {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
}

func TestInsertCommentThread(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/commentThreads" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":"T1"}`)
	})
	out, err := c.InsertCommentThread(context.Background(), CommentThread{
		Snippet: &CommentThreadSnippet{
			ChannelID:       "UC1",
			VideoID:         "V1",
			TopLevelComment: &Comment{Snippet: &CommentSnippet{TextOriginal: "hi"}},
		},
	}, nil)
	if err != nil {
		t.Fatalf("InsertCommentThread: %v", err)
	}
	if out.ID != "T1" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestIterVideoThreads_Pagination(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("pageToken") == "" {
			_, _ = io.WriteString(w, `{"items":[{"id":"T1"}],"nextPageToken":"N"}`)
			return
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"T2"}]}`)
	})
	var ids []string
	err := c.IterVideoThreads(context.Background(), "V1", []string{CommentThreadPartSnippet}, 100, "", func(t CommentThread) error {
		ids = append(ids, t.ID)
		return nil
	})
	if err != nil {
		t.Fatalf("IterVideoThreads: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("unexpected ids: %v", ids)
	}
}

// ---------------------------------------------------------------------------
// Playlists
// ---------------------------------------------------------------------------

func TestListPlaylists(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/playlists" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("channelId") != "UC1" {
			t.Errorf("channelId: %s", r.URL.Query().Get("channelId"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"PL1","snippet":{"title":"My PL"}}]}`)
	})
	resp, err := c.ListPlaylists(context.Background(), ListPlaylistsParams{
		Parts: []string{PlaylistPartSnippet}, ChannelID: "UC1",
	})
	if err != nil {
		t.Fatalf("ListPlaylists: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Snippet.Title != "My PL" {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
}

func TestInsertPlaylist_InfersParts(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/playlists" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("part") != "snippet,status" {
			t.Errorf("part: %s", r.URL.Query().Get("part"))
		}
		_, _ = io.WriteString(w, `{"id":"PL1"}`)
	})
	out, err := c.InsertPlaylist(context.Background(), Playlist{
		Snippet: &PlaylistSnippet{Title: "x"},
		Status:  &PlaylistStatus{PrivacyStatus: PrivacyPrivate},
	}, nil)
	if err != nil {
		t.Fatalf("InsertPlaylist: %v", err)
	}
	if out.ID != "PL1" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestUpdatePlaylist(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/playlists" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":"PL1"}`)
	})
	if _, err := c.UpdatePlaylist(context.Background(), Playlist{ID: "PL1", Snippet: &PlaylistSnippet{Title: "x"}}, nil); err != nil {
		t.Fatalf("UpdatePlaylist: %v", err)
	}
}

func TestDeletePlaylist(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Query().Get("id") != "PL1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeletePlaylist(context.Background(), "PL1"); err != nil {
		t.Fatalf("DeletePlaylist: %v", err)
	}
}

func TestIterMyPlaylists(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mine") != "true" {
			t.Errorf("mine: %s", r.URL.Query().Get("mine"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"PL1"}]}`)
	})
	var ids []string
	err := c.IterMyPlaylists(context.Background(), []string{PlaylistPartSnippet}, 25, func(p Playlist) error {
		ids = append(ids, p.ID)
		return nil
	})
	if err != nil {
		t.Fatalf("IterMyPlaylists: %v", err)
	}
	if len(ids) != 1 || ids[0] != "PL1" {
		t.Fatalf("unexpected: %v", ids)
	}
}

// ---------------------------------------------------------------------------
// Playlist items
// ---------------------------------------------------------------------------

func TestListPlaylistItems(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/playlistItems" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("playlistId") != "PL1" {
			t.Errorf("playlistId: %s", r.URL.Query().Get("playlistId"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"PI1","contentDetails":{"videoId":"V1"}}]}`)
	})
	resp, err := c.ListPlaylistItems(context.Background(), ListPlaylistItemsParams{
		Parts: []string{PlaylistItemPartSnippet, PlaylistItemPartContentDetails}, PlaylistID: "PL1",
	})
	if err != nil {
		t.Fatalf("ListPlaylistItems: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].ContentDetails.VideoID != "V1" {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
}

func TestInsertPlaylistItem(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/playlistItems" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"videoId":"V1"`) {
			t.Errorf("body: %s", body)
		}
		_, _ = io.WriteString(w, `{"id":"PI1"}`)
	})
	out, err := c.InsertPlaylistItem(context.Background(), PlaylistItem{
		Snippet: &PlaylistItemSnippet{
			PlaylistID: "PL1",
			ResourceID: &ResourceID{Kind: "youtube#video", VideoID: "V1"},
		},
	}, nil)
	if err != nil {
		t.Fatalf("InsertPlaylistItem: %v", err)
	}
	if out.ID != "PI1" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestUpdatePlaylistItem(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/playlistItems" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":"PI1"}`)
	})
	pos := 2
	if _, err := c.UpdatePlaylistItem(context.Background(), PlaylistItem{
		ID:      "PI1",
		Snippet: &PlaylistItemSnippet{PlaylistID: "PL1", Position: &pos, ResourceID: &ResourceID{VideoID: "V1"}},
	}, nil); err != nil {
		t.Fatalf("UpdatePlaylistItem: %v", err)
	}
}

func TestDeletePlaylistItem(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Query().Get("id") != "PI1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeletePlaylistItem(context.Background(), "PI1"); err != nil {
		t.Fatalf("DeletePlaylistItem: %v", err)
	}
}

func TestIterPlaylistItems_Pagination(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("pageToken") == "" {
			_, _ = io.WriteString(w, `{"items":[{"id":"PI1"}],"nextPageToken":"N"}`)
			return
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"PI2"}]}`)
	})
	var ids []string
	err := c.IterPlaylistItems(context.Background(), "PL1", []string{PlaylistItemPartSnippet}, 50, func(it PlaylistItem) error {
		ids = append(ids, it.ID)
		return nil
	})
	if err != nil {
		t.Fatalf("IterPlaylistItems: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("unexpected: %v", ids)
	}
}

// ---------------------------------------------------------------------------
// Subscriptions
// ---------------------------------------------------------------------------

func TestListSubscriptions(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/subscriptions" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if q.Get("mine") != "true" {
			t.Errorf("mine: %s", q.Get("mine"))
		}
		if q.Get("order") != "alphabetical" {
			t.Errorf("order: %s", q.Get("order"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"SB1","snippet":{"title":"Chan"}}]}`)
	})
	mine := true
	resp, err := c.ListSubscriptions(context.Background(), ListSubscriptionsParams{
		Parts: []string{SubscriptionPartSnippet}, Mine: &mine, Order: SubscriptionOrderAlphabetical,
	})
	if err != nil {
		t.Fatalf("ListSubscriptions: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Snippet.Title != "Chan" {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
}

func TestInsertSubscription(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/subscriptions" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"channelId":"UC2"`) {
			t.Errorf("body: %s", body)
		}
		_, _ = io.WriteString(w, `{"id":"SB1"}`)
	})
	out, err := c.InsertSubscription(context.Background(), Subscription{
		Snippet: &SubscriptionSnippet{ResourceID: &ResourceID{Kind: "youtube#channel", ChannelID: "UC2"}},
	}, nil)
	if err != nil {
		t.Fatalf("InsertSubscription: %v", err)
	}
	if out.ID != "SB1" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestDeleteSubscription(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Query().Get("id") != "SB1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteSubscription(context.Background(), "SB1"); err != nil {
		t.Fatalf("DeleteSubscription: %v", err)
	}
}

func TestIterMySubscriptions(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mine") != "true" {
			t.Errorf("mine: %s", r.URL.Query().Get("mine"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"SB1"}]}`)
	})
	var ids []string
	err := c.IterMySubscriptions(context.Background(), []string{SubscriptionPartSnippet}, 50, "", func(s Subscription) error {
		ids = append(ids, s.ID)
		return nil
	})
	if err != nil {
		t.Fatalf("IterMySubscriptions: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("unexpected: %v", ids)
	}
}

// ---------------------------------------------------------------------------
// Videos list / rate / getRating / reportAbuse
// ---------------------------------------------------------------------------

func TestListVideos(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/videos" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if q.Get("chart") != "mostPopular" {
			t.Errorf("chart: %s", q.Get("chart"))
		}
		if q.Get("regionCode") != "US" {
			t.Errorf("regionCode: %s", q.Get("regionCode"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"V1","snippet":{"title":"Top"}}]}`)
	})
	resp, err := c.ListVideos(context.Background(), ListVideosParams{
		Parts: []string{VideoPartSnippet}, Chart: "mostPopular", RegionCode: "US",
	})
	if err != nil {
		t.Fatalf("ListVideos: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Snippet.Title != "Top" {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
}

func TestRateVideo(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/videos/rate" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("id") != "V1" || q.Get("rating") != "like" {
			t.Errorf("query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.RateVideo(context.Background(), "V1", RatingLike); err != nil {
		t.Fatalf("RateVideo: %v", err)
	}
}

func TestGetVideoRating(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/videos/getRating" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("id") != "V1,V2" {
			t.Errorf("id: %s", r.URL.Query().Get("id"))
		}
		_, _ = io.WriteString(w, `{"items":[{"videoId":"V1","rating":"like"},{"videoId":"V2","rating":"none"}]}`)
	})
	resp, err := c.GetVideoRating(context.Background(), []string{"V1", "V2"})
	if err != nil {
		t.Fatalf("GetVideoRating: %v", err)
	}
	if len(resp.Items) != 2 || resp.Items[0].Rating != RatingLike {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
}

func TestReportVideoAbuse(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/videos/reportAbuse" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"reasonId":"spam"`) {
			t.Errorf("body: %s", body)
		}
		if strings.Contains(string(body), "secondaryReasonId") {
			t.Errorf("secondaryReasonId should be omitted: %s", body)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.ReportVideoAbuse(context.Background(), ReportVideoAbuseParams{
		VideoID: "V1", ReasonID: "spam",
	})
	if err != nil {
		t.Fatalf("ReportVideoAbuse: %v", err)
	}
}

func TestIterVideosByChart_Pagination(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("pageToken") == "" {
			_, _ = io.WriteString(w, `{"items":[{"id":"V1"}],"nextPageToken":"N"}`)
			return
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"V2"}]}`)
	})
	var ids []string
	err := c.IterVideosByChart(context.Background(), []string{VideoPartSnippet}, "", 5, "", "", func(v Video) error {
		ids = append(ids, v.ID)
		return nil
	})
	if err != nil {
		t.Fatalf("IterVideosByChart: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("unexpected: %v", ids)
	}
}

// ---------------------------------------------------------------------------
// Search (full parity)
// ---------------------------------------------------------------------------

func TestListSearch(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/search" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if q.Get("q") != "go" {
			t.Errorf("q: %s", q.Get("q"))
		}
		if q.Get("type") != "video,channel" {
			t.Errorf("type: %s", q.Get("type"))
		}
		if q.Get("safeSearch") != "strict" {
			t.Errorf("safeSearch: %s", q.Get("safeSearch"))
		}
		if q.Get("videoDuration") != "short" {
			t.Errorf("videoDuration: %s", q.Get("videoDuration"))
		}
		_, _ = io.WriteString(w, `{"items":[{"id":{"kind":"youtube#video","videoId":"V1"},"snippet":{"title":"Go"}}],"regionCode":"US"}`)
	})
	resp, err := c.ListSearch(context.Background(), ListSearchParams{
		Q:             "go",
		Types:         []string{SearchResultTypeVideo, SearchResultTypeChannel},
		SafeSearch:    SafeSearchStrict,
		VideoDuration: VideoDurationShort,
	})
	if err != nil {
		t.Fatalf("ListSearch: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].ID.VideoID != "V1" {
		t.Fatalf("unexpected: %+v", resp.Items)
	}
	if resp.RegionCode != "US" {
		t.Fatalf("expected regionCode")
	}
}

func TestIterSearchResults_MaxPages(t *testing.T) {
	calls := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		// Always returns a nextPageToken so only maxPages bounds the loop.
		_, _ = io.WriteString(w, `{"items":[{"id":{"videoId":"V1"}}],"nextPageToken":"N"}`)
	})
	count := 0
	err := c.IterSearchResults(context.Background(), ListSearchParams{Q: "x"}, 3, func(SearchItem) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("IterSearchResults: %v", err)
	}
	if calls != 3 || count != 3 {
		t.Fatalf("expected 3 pages, got calls=%d count=%d", calls, count)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func boolPtr(b bool) *bool { return &b }
