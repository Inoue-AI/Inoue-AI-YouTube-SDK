package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Playlist parts accepted by the part parameter.
const (
	PlaylistPartID             = "id"
	PlaylistPartSnippet        = "snippet"
	PlaylistPartStatus         = "status"
	PlaylistPartContentDetails = "contentDetails"
	PlaylistPartPlayer         = "player"
	PlaylistPartLocalizations  = "localizations"
)

// PlaylistSnippet is the snippet part of a Playlist resource.
type PlaylistSnippet struct {
	PublishedAt     string            `json:"publishedAt,omitempty"`
	ChannelID       string            `json:"channelId,omitempty"`
	Title           string            `json:"title,omitempty"`
	Description     string            `json:"description,omitempty"`
	Thumbnails      *ThumbnailDetails `json:"thumbnails,omitempty"`
	ChannelTitle    string            `json:"channelTitle,omitempty"`
	DefaultLanguage string            `json:"defaultLanguage,omitempty"`
	Localized       *Localization     `json:"localized,omitempty"`
}

// PlaylistStatus is the status part of a Playlist resource.
type PlaylistStatus struct {
	PrivacyStatus string `json:"privacyStatus,omitempty"`
}

// PlaylistContentDetails is the contentDetails part of a Playlist resource.
type PlaylistContentDetails struct {
	ItemCount *int `json:"itemCount,omitempty"`
}

// PlaylistPlayer is the player part of a Playlist resource.
type PlaylistPlayer struct {
	EmbedHTML string `json:"embedHtml,omitempty"`
}

// Playlist is a YouTube Playlist resource (request body and response).
type Playlist struct {
	Kind           string                  `json:"kind,omitempty"`
	Etag           string                  `json:"etag,omitempty"`
	ID             string                  `json:"id,omitempty"`
	Snippet        *PlaylistSnippet        `json:"snippet,omitempty"`
	Status         *PlaylistStatus         `json:"status,omitempty"`
	ContentDetails *PlaylistContentDetails `json:"contentDetails,omitempty"`
	Player         *PlaylistPlayer         `json:"player,omitempty"`
	Localizations  map[string]Localization `json:"localizations,omitempty"`
}

// PlaylistListResponse is the response from ListPlaylists.
type PlaylistListResponse struct {
	Kind          string     `json:"kind,omitempty"`
	Etag          string     `json:"etag,omitempty"`
	NextPageToken string     `json:"nextPageToken,omitempty"`
	PrevPageToken string     `json:"prevPageToken,omitempty"`
	PageInfo      *PageInfo  `json:"pageInfo,omitempty"`
	Items         []Playlist `json:"items"`
}

// ListPlaylistsParams configures ListPlaylists.
type ListPlaylistsParams struct {
	// Parts is the list of resource parts to include. Required.
	Parts []string
	// ChannelID returns playlists owned by the channel.
	ChannelID string
	// IDs restricts the result to specific playlist IDs.
	IDs []string
	// Mine, when non-nil, returns the authenticated user's playlists.
	Mine *bool
	// MaxResults is the page size. Defaults to 5.
	MaxResults int
	// PageToken is the pagination cursor.
	PageToken string
	// HL is the language for localized metadata.
	HL string
}

// ListPlaylists returns playlists matching the given filter. Mirrors
// playlists.list.
func (c *Client) ListPlaylists(ctx context.Context, p ListPlaylistsParams) (*PlaylistListResponse, error) {
	if len(p.Parts) == 0 {
		return nil, errors.New("youtube: ListPlaylists requires at least one part")
	}
	maxResults := p.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}
	q := url.Values{}
	q.Set("part", strings.Join(p.Parts, ","))
	q.Set("maxResults", strconv.Itoa(maxResults))
	if p.ChannelID != "" {
		q.Set("channelId", p.ChannelID)
	}
	if len(p.IDs) > 0 {
		q.Set("id", strings.Join(p.IDs, ","))
	}
	if p.Mine != nil {
		q.Set("mine", strconv.FormatBool(*p.Mine))
	}
	if p.PageToken != "" {
		q.Set("pageToken", p.PageToken)
	}
	if p.HL != "" {
		q.Set("hl", p.HL)
	}

	out := &PlaylistListResponse{}
	if err := c.doGet(ctx, "playlists", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IterMyPlaylists walks every playlist owned by the authenticated user,
// invoking fn for each. It pages through the result set until exhausted or fn
// returns an error. This is the Go-idiomatic equivalent of the Python iter_mine
// async generator.
func (c *Client) IterMyPlaylists(
	ctx context.Context,
	parts []string,
	maxResults int,
	fn func(Playlist) error,
) error {
	if maxResults <= 0 {
		maxResults = 25
	}
	mine := true
	pageToken := ""
	for {
		page, err := c.ListPlaylists(ctx, ListPlaylistsParams{
			Parts:      parts,
			Mine:       &mine,
			MaxResults: maxResults,
			PageToken:  pageToken,
		})
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			if err := fn(item); err != nil {
				return err
			}
		}
		if page.NextPageToken == "" {
			return nil
		}
		pageToken = page.NextPageToken
	}
}

// InsertPlaylist creates a new playlist. When parts is empty, the parts are
// inferred from the populated fields of the body. Mirrors playlists.insert.
// Requires OAuth.
func (c *Client) InsertPlaylist(ctx context.Context, body Playlist, parts []string) (*Playlist, error) {
	if len(parts) == 0 {
		parts = inferPlaylistParts(body)
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &Playlist{}
	if err := c.doJSON(ctx, http.MethodPost, "playlists", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdatePlaylist updates an existing playlist. The body must contain id and
// snippet.title. When parts is empty, the parts are inferred from the populated
// fields of the body. Mirrors playlists.update. Requires OAuth.
func (c *Client) UpdatePlaylist(ctx context.Context, body Playlist, parts []string) (*Playlist, error) {
	if len(parts) == 0 {
		parts = inferPlaylistParts(body)
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &Playlist{}
	if err := c.doJSON(ctx, http.MethodPut, "playlists", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeletePlaylist deletes a playlist by ID. Mirrors playlists.delete. Requires
// OAuth.
func (c *Client) DeletePlaylist(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("youtube: DeletePlaylist requires an ID")
	}
	q := url.Values{}
	q.Set("id", id)
	return c.doDelete(ctx, "playlists", q)
}

// inferPlaylistParts derives the resource parts from the populated fields of a
// Playlist body, mirroring the Python _infer_parts helper.
func inferPlaylistParts(body Playlist) []string {
	parts := make([]string, 0, 3)
	if body.Snippet != nil {
		parts = append(parts, PlaylistPartSnippet)
	}
	if body.Status != nil {
		parts = append(parts, PlaylistPartStatus)
	}
	if body.Localizations != nil {
		parts = append(parts, PlaylistPartLocalizations)
	}
	if len(parts) == 0 {
		return []string{PlaylistPartSnippet}
	}
	return parts
}
