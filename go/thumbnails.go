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

// ThumbnailSetResponse is the response from setting a custom thumbnail.
type ThumbnailSetResponse struct {
	Kind  string `json:"kind"`
	Etag  string `json:"etag"`
	Items []struct {
		Default *ThumbnailInfo `json:"default,omitempty"`
		Medium  *ThumbnailInfo `json:"medium,omitempty"`
		High    *ThumbnailInfo `json:"high,omitempty"`
	} `json:"items"`
}

// ThumbnailInfo describes one rendition of a thumbnail image.
type ThumbnailInfo struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// SetThumbnailParams configures SetThumbnail.
type SetThumbnailParams struct {
	// VideoID is the target video. Required.
	VideoID string
	// Image is the raw image bytes (JPEG or PNG, max 2 MB).
	Image []byte
	// ContentType is the image MIME type. Defaults to "image/jpeg".
	ContentType string
}

// SetThumbnail uploads a custom thumbnail for a video via a direct media
// upload. Requires OAuth.
func (c *Client) SetThumbnail(ctx context.Context, p SetThumbnailParams) (*ThumbnailSetResponse, error) {
	if err := c.requireOAuth(); err != nil {
		return nil, err
	}
	if p.VideoID == "" {
		return nil, errors.New("youtube: SetThumbnail requires a VideoID")
	}
	if len(p.Image) == 0 {
		return nil, errors.New("youtube: SetThumbnail requires image bytes")
	}
	contentType := p.ContentType
	if contentType == "" {
		contentType = "image/jpeg"
	}

	query := url.Values{}
	query.Set("videoId", p.VideoID)
	query.Set("uploadType", "media")

	out := &ThumbnailSetResponse{}
	if err := c.uploadMedia(ctx, "thumbnails/set", query, p.Image, contentType, out); err != nil {
		return nil, err
	}
	return out, nil
}

// uploadMedia performs a single-shot ("media") binary upload against the upload
// API and decodes the JSON response into out.
func (c *Client) uploadMedia(
	ctx context.Context,
	path string,
	query url.Values,
	data []byte,
	contentType string,
	out any,
) error {
	fullURL := c.uploadBaseURL + "/" + strings.TrimLeft(path, "/") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("youtube: build media-upload request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", strconv.Itoa(len(data)))
	req.Header.Set("Accept", "application/json")
	c.applyAuth(req, query)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("youtube: media-upload request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("youtube: read media-upload response: %w", err)
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
			Message:    fmt.Sprintf("media upload returned non-JSON: %v", err),
			Body:       raw,
		}
	}
	return nil
}
