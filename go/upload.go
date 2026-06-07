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
	"strconv"
	"strings"
)

// VideoSnippetInput is the writable snippet for a video insert/update.
type VideoSnippetInput struct {
	Title                string   `json:"title,omitempty"`
	Description          string   `json:"description,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	CategoryID           string   `json:"categoryId,omitempty"`
	DefaultLanguage      string   `json:"defaultLanguage,omitempty"`
	DefaultAudioLanguage string   `json:"defaultAudioLanguage,omitempty"`
}

// VideoStatusInput is the writable status for a video insert/update.
type VideoStatusInput struct {
	PrivacyStatus           string `json:"privacyStatus,omitempty"`
	License                 string `json:"license,omitempty"`
	Embeddable              *bool  `json:"embeddable,omitempty"`
	PublicStatsViewable     *bool  `json:"publicStatsViewable,omitempty"`
	SelfDeclaredMadeForKids *bool  `json:"selfDeclaredMadeForKids,omitempty"`
	PublishAt               string `json:"publishAt,omitempty"`
}

// VideoResource is the request/response body for video insert and update. On
// update, ID must be set.
type VideoResource struct {
	ID      string             `json:"id,omitempty"`
	Snippet *VideoSnippetInput `json:"snippet,omitempty"`
	Status  *VideoStatusInput  `json:"status,omitempty"`
}

// UploadVideoParams configures InsertVideo.
type UploadVideoParams struct {
	// Metadata is the video resource (snippet/status) to attach.
	Metadata VideoResource
	// Media is the source of the video bytes. Required.
	Media io.Reader
	// MediaSize is the total byte length of Media. Required (the resumable
	// protocol must advertise the total length up front).
	MediaSize int64
	// ContentType is the MIME type of the media (e.g. "video/mp4"). Defaults
	// to "video/*".
	ContentType string
	// Parts overrides the resource parts. Defaults are inferred from Metadata.
	Parts []string
	// NotifySubscribers controls whether subscribers are notified. Defaults to
	// true; set the pointer to &false to suppress.
	NotifySubscribers *bool
	// ChunkSize overrides the upload chunk size. Defaults to
	// DefaultUploadChunkSize. Must be a multiple of 256 KiB.
	ChunkSize int
}

// InsertVideo uploads a new video using the resumable upload protocol.
//
// The upload is performed in two phases: (1) a session-initiation POST that
// returns an upload URL, then (2) one or more chunked PUTs streaming the media.
// Requires an OAuth access token.
func (c *Client) InsertVideo(ctx context.Context, p UploadVideoParams) (*Video, error) {
	if err := c.requireOAuth(); err != nil {
		return nil, err
	}
	if p.Media == nil {
		return nil, errors.New("youtube: InsertVideo requires a Media reader")
	}
	if p.MediaSize <= 0 {
		return nil, errors.New("youtube: InsertVideo requires a positive MediaSize")
	}

	contentType := p.ContentType
	if contentType == "" {
		contentType = "video/*"
	}
	chunkSize := p.ChunkSize
	if chunkSize <= 0 {
		chunkSize = DefaultUploadChunkSize
	}

	parts := p.Parts
	if len(parts) == 0 {
		parts = inferVideoParts(p.Metadata)
	}
	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))
	query.Set("uploadType", "resumable")
	if p.NotifySubscribers != nil && !*p.NotifySubscribers {
		query.Set("notifySubscribers", "false")
	}

	uploadURL, err := c.initResumableSession(ctx, "videos", query, p.Metadata, contentType, p.MediaSize)
	if err != nil {
		return nil, err
	}

	out := &Video{}
	if err := c.uploadResumable(ctx, uploadURL, p.Media, p.MediaSize, contentType, chunkSize, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateVideo updates a video's metadata. Metadata.ID must be set. Properties
// omitted from the body are cleared by YouTube, so callers should send the full
// desired state of each included part. Requires OAuth.
func (c *Client) UpdateVideo(ctx context.Context, metadata VideoResource, parts []string) (*Video, error) {
	if metadata.ID == "" {
		return nil, errors.New("youtube: UpdateVideo requires Metadata.ID")
	}
	if len(parts) == 0 {
		parts = inferVideoParts(metadata)
	}
	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	out := &Video{}
	if err := c.doJSON(ctx, http.MethodPut, "videos", query, metadata, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteVideo permanently deletes a video by ID. Requires OAuth.
func (c *Client) DeleteVideo(ctx context.Context, videoID string) error {
	if videoID == "" {
		return errors.New("youtube: DeleteVideo requires a video ID")
	}
	query := url.Values{}
	query.Set("id", videoID)
	return c.doDelete(ctx, "videos", query)
}

// initResumableSession performs the resumable-upload initiation POST and
// returns the session upload URL from the Location header.
func (c *Client) initResumableSession(
	ctx context.Context,
	path string,
	query url.Values,
	metadata any,
	contentType string,
	mediaSize int64,
) (string, error) {
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("youtube: encode upload metadata: %w", err)
	}
	fullURL := c.uploadBaseURL + "/" + strings.TrimLeft(path, "/") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(encoded))
	if err != nil {
		return "", fmt.Errorf("youtube: build upload-init request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Upload-Content-Type", contentType)
	req.Header.Set("X-Upload-Content-Length", strconv.FormatInt(mediaSize, 10))
	c.applyAuth(req, query)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("youtube: upload-init request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", parseErrorBody(resp.StatusCode, raw)
	}
	uploadURL := resp.Header.Get("Location")
	if uploadURL == "" {
		return "", &Error{
			StatusCode: resp.StatusCode,
			Status:     "UPLOAD_INIT_FAILED",
			Message:    "resumable upload init returned no Location header",
			Body:       raw,
		}
	}
	return uploadURL, nil
}

// uploadResumable streams media to the session upload URL in chunks, honoring
// the 308 "Resume Incomplete" protocol, and decodes the final resource into
// out.
func (c *Client) uploadResumable(
	ctx context.Context,
	uploadURL string,
	media io.Reader,
	mediaSize int64,
	contentType string,
	chunkSize int,
	out any,
) error {
	buf := make([]byte, chunkSize)
	var offset int64

	for offset < mediaSize {
		n, readErr := io.ReadFull(media, buf)
		if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
			return fmt.Errorf("youtube: read media chunk: %w", readErr)
		}
		if n == 0 {
			break
		}
		chunk := buf[:n]
		end := offset + int64(n) - 1

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(chunk))
		if err != nil {
			return fmt.Errorf("youtube: build chunk request: %w", err)
		}
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Content-Length", strconv.Itoa(n))
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, end, mediaSize))

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("youtube: chunk upload at offset %d: %w", offset, err)
		}
		raw, _ := io.ReadAll(resp.Body)
		status := resp.StatusCode
		resp.Body.Close()

		switch {
		case status == http.StatusOK || status == http.StatusCreated:
			// Upload complete — decode the created resource.
			if out == nil || len(raw) == 0 {
				return nil
			}
			if err := json.Unmarshal(raw, out); err != nil {
				return &Error{
					StatusCode: status,
					Status:     "INVALID_RESPONSE",
					Message:    fmt.Sprintf("upload returned non-JSON: %v", err),
					Body:       raw,
				}
			}
			return nil
		case status == 308:
			// Resume Incomplete — continue with the next chunk.
			offset = end + 1
		default:
			return parseErrorBody(status, raw)
		}
	}

	return &Error{
		Status:  "UPLOAD_INCOMPLETE",
		Message: "upload finished without a final 200/201 response",
	}
}

// inferVideoParts derives the resource parts to request from the populated
// fields of a VideoResource.
func inferVideoParts(v VideoResource) []string {
	parts := make([]string, 0, 2)
	if v.Snippet != nil {
		parts = append(parts, "snippet")
	}
	if v.Status != nil {
		parts = append(parts, "status")
	}
	if len(parts) == 0 {
		return []string{"snippet"}
	}
	return parts
}
