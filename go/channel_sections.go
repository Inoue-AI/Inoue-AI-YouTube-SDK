package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ChannelSection parts accepted by the part parameter.
const (
	ChannelSectionPartID             = "id"
	ChannelSectionPartSnippet        = "snippet"
	ChannelSectionPartContentDetails = "contentDetails"
)

// ChannelSection type constants (snippet.type).
const (
	ChannelSectionTypeSinglePlaylist    = "singlePlaylist"
	ChannelSectionTypeMultiplePlaylists = "multiplePlaylists"
	ChannelSectionTypeAllPlaylists      = "allPlaylists"
	ChannelSectionTypeRecentUploads     = "recentUploads"
	ChannelSectionTypePopularUploads    = "popularUploads"
	ChannelSectionTypeMultipleChannels  = "multipleChannels"
	ChannelSectionTypeSubscriptions     = "subscriptions"
	ChannelSectionTypeLiveEvents        = "liveEvents"
	ChannelSectionTypeUpcomingEvents    = "upcomingEvents"
	ChannelSectionTypeCompletedEvents   = "completedEvents"
)

// ChannelSectionSnippet is the snippet part of a ChannelSection resource.
type ChannelSectionSnippet struct {
	Type            string        `json:"type,omitempty"`
	ChannelID       string        `json:"channelId,omitempty"`
	Title           string        `json:"title,omitempty"`
	Position        *int          `json:"position,omitempty"`
	DefaultLanguage string        `json:"defaultLanguage,omitempty"`
	Localized       *Localization `json:"localized,omitempty"`
}

// ChannelSectionContentDetails is the contentDetails part of a ChannelSection.
type ChannelSectionContentDetails struct {
	Playlists []string `json:"playlists,omitempty"`
	Channels  []string `json:"channels,omitempty"`
}

// ChannelSection is a YouTube ChannelSection resource (request body and
// response).
type ChannelSection struct {
	Kind           string                        `json:"kind,omitempty"`
	Etag           string                        `json:"etag,omitempty"`
	ID             string                        `json:"id,omitempty"`
	Snippet        *ChannelSectionSnippet        `json:"snippet,omitempty"`
	ContentDetails *ChannelSectionContentDetails `json:"contentDetails,omitempty"`
}

// ChannelSectionListResponse is the response from ListChannelSections.
type ChannelSectionListResponse struct {
	Kind  string           `json:"kind,omitempty"`
	Etag  string           `json:"etag,omitempty"`
	Items []ChannelSection `json:"items"`
}

// ListChannelSectionsParams configures ListChannelSections.
type ListChannelSectionsParams struct {
	// Parts is the list of resource parts to include. Required.
	Parts []string
	// ChannelID restricts the result to a single channel.
	ChannelID string
	// IDs restricts the result to specific channel-section IDs.
	IDs []string
	// Mine, when non-nil, restricts the result to the authenticated user's
	// channel sections.
	Mine *bool
}

// ListChannelSections returns channel sections matching the given filter.
// Mirrors channelSections.list.
func (c *Client) ListChannelSections(ctx context.Context, p ListChannelSectionsParams) (*ChannelSectionListResponse, error) {
	if len(p.Parts) == 0 {
		return nil, errors.New("youtube: ListChannelSections requires at least one part")
	}
	q := url.Values{}
	q.Set("part", strings.Join(p.Parts, ","))
	if p.ChannelID != "" {
		q.Set("channelId", p.ChannelID)
	}
	if len(p.IDs) > 0 {
		q.Set("id", strings.Join(p.IDs, ","))
	}
	if p.Mine != nil {
		q.Set("mine", strconv.FormatBool(*p.Mine))
	}

	out := &ChannelSectionListResponse{}
	if err := c.doGet(ctx, "channelSections", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// InsertChannelSection creates a new channel section (max 10 per channel).
// When parts is empty, the parts are inferred from the populated fields of the
// body. Mirrors channelSections.insert. Requires OAuth.
func (c *Client) InsertChannelSection(ctx context.Context, body ChannelSection, parts []string) (*ChannelSection, error) {
	if len(parts) == 0 {
		parts = inferChannelSectionParts(body)
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &ChannelSection{}
	if err := c.doJSON(ctx, http.MethodPost, "channelSections", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateChannelSection updates an existing channel section. When parts is
// empty, the parts are inferred from the populated fields of the body. Mirrors
// channelSections.update. Requires OAuth.
func (c *Client) UpdateChannelSection(ctx context.Context, body ChannelSection, parts []string) (*ChannelSection, error) {
	if len(parts) == 0 {
		parts = inferChannelSectionParts(body)
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &ChannelSection{}
	if err := c.doJSON(ctx, http.MethodPut, "channelSections", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteChannelSection deletes a channel section by ID. Mirrors
// channelSections.delete. Requires OAuth.
func (c *Client) DeleteChannelSection(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("youtube: DeleteChannelSection requires an ID")
	}
	q := url.Values{}
	q.Set("id", id)
	return c.doDelete(ctx, "channelSections", q)
}

// inferChannelSectionParts derives the resource parts from the populated fields
// of a ChannelSection body, mirroring the Python _infer_parts helper.
func inferChannelSectionParts(body ChannelSection) []string {
	parts := make([]string, 0, 2)
	if body.Snippet != nil {
		parts = append(parts, ChannelSectionPartSnippet)
	}
	if body.ContentDetails != nil {
		parts = append(parts, ChannelSectionPartContentDetails)
	}
	if len(parts) == 0 {
		return []string{ChannelSectionPartSnippet}
	}
	return parts
}
