package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Subscription parts accepted by the part parameter.
const (
	SubscriptionPartID                = "id"
	SubscriptionPartSnippet           = "snippet"
	SubscriptionPartContentDetails    = "contentDetails"
	SubscriptionPartSubscriberSnippet = "subscriberSnippet"
)

// Subscription sort orders.
const (
	SubscriptionOrderAlphabetical = "alphabetical"
	SubscriptionOrderRelevance    = "relevance"
	SubscriptionOrderUnread       = "unread"
)

// SubscriptionSnippet is the snippet part of a Subscription resource.
type SubscriptionSnippet struct {
	PublishedAt string            `json:"publishedAt,omitempty"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	ResourceID  *ResourceID       `json:"resourceId,omitempty"`
	ChannelID   string            `json:"channelId,omitempty"`
	Thumbnails  *ThumbnailDetails `json:"thumbnails,omitempty"`
}

// SubscriptionContentDetails is the contentDetails part of a Subscription.
type SubscriptionContentDetails struct {
	TotalItemCount *int   `json:"totalItemCount,omitempty"`
	NewItemCount   *int   `json:"newItemCount,omitempty"`
	ActivityType   string `json:"activityType,omitempty"`
}

// SubscriberSnippet is the subscriberSnippet part of a Subscription resource.
type SubscriberSnippet struct {
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	ChannelID   string            `json:"channelId,omitempty"`
	Thumbnails  *ThumbnailDetails `json:"thumbnails,omitempty"`
}

// Subscription is a YouTube Subscription resource (request body and response).
type Subscription struct {
	Kind              string                      `json:"kind,omitempty"`
	Etag              string                      `json:"etag,omitempty"`
	ID                string                      `json:"id,omitempty"`
	Snippet           *SubscriptionSnippet        `json:"snippet,omitempty"`
	ContentDetails    *SubscriptionContentDetails `json:"contentDetails,omitempty"`
	SubscriberSnippet *SubscriberSnippet          `json:"subscriberSnippet,omitempty"`
}

// SubscriptionListResponse is the response from ListSubscriptions.
type SubscriptionListResponse struct {
	Kind          string         `json:"kind,omitempty"`
	Etag          string         `json:"etag,omitempty"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
	PrevPageToken string         `json:"prevPageToken,omitempty"`
	PageInfo      *PageInfo      `json:"pageInfo,omitempty"`
	Items         []Subscription `json:"items"`
}

// ListSubscriptionsParams configures ListSubscriptions.
type ListSubscriptionsParams struct {
	// Parts is the list of resource parts to include. Required.
	Parts []string
	// ChannelID returns subscriptions of the given channel.
	ChannelID string
	// IDs restricts the result to specific subscription IDs.
	IDs []string
	// Mine, when non-nil, returns the authenticated user's subscriptions.
	Mine *bool
	// MyRecentSubscribers, when non-nil, returns recent subscribers.
	MyRecentSubscribers *bool
	// MySubscribers, when non-nil, returns the authenticated user's subscribers.
	MySubscribers *bool
	// ForChannelID restricts the result to subscriptions of the given channels.
	ForChannelID string
	// MaxResults is the page size. Defaults to 5.
	MaxResults int
	// PageToken is the pagination cursor.
	PageToken string
	// Order selects the sort order.
	Order string
}

// ListSubscriptions returns subscriptions matching the given filter. Mirrors
// subscriptions.list.
func (c *Client) ListSubscriptions(ctx context.Context, p ListSubscriptionsParams) (*SubscriptionListResponse, error) {
	if len(p.Parts) == 0 {
		return nil, errors.New("youtube: ListSubscriptions requires at least one part")
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
	if p.MyRecentSubscribers != nil {
		q.Set("myRecentSubscribers", strconv.FormatBool(*p.MyRecentSubscribers))
	}
	if p.MySubscribers != nil {
		q.Set("mySubscribers", strconv.FormatBool(*p.MySubscribers))
	}
	if p.ForChannelID != "" {
		q.Set("forChannelId", p.ForChannelID)
	}
	if p.PageToken != "" {
		q.Set("pageToken", p.PageToken)
	}
	if p.Order != "" {
		q.Set("order", p.Order)
	}

	out := &SubscriptionListResponse{}
	if err := c.doGet(ctx, "subscriptions", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IterMySubscriptions walks every subscription of the authenticated user,
// invoking fn for each. It pages through the result set until exhausted or fn
// returns an error. This is the Go-idiomatic equivalent of the Python iter_mine
// async generator.
func (c *Client) IterMySubscriptions(
	ctx context.Context,
	parts []string,
	maxResults int,
	order string,
	fn func(Subscription) error,
) error {
	if maxResults <= 0 {
		maxResults = 50
	}
	mine := true
	pageToken := ""
	for {
		page, err := c.ListSubscriptions(ctx, ListSubscriptionsParams{
			Parts:      parts,
			Mine:       &mine,
			MaxResults: maxResults,
			PageToken:  pageToken,
			Order:      order,
		})
		if err != nil {
			return err
		}
		for _, sub := range page.Items {
			if err := fn(sub); err != nil {
				return err
			}
		}
		if page.NextPageToken == "" {
			return nil
		}
		pageToken = page.NextPageToken
	}
}

// InsertSubscription subscribes to a channel. The body must include
// snippet.resourceId with kind "youtube#channel" and the target channelId. When
// parts is empty it defaults to ["snippet"]. Mirrors subscriptions.insert.
// Requires OAuth.
func (c *Client) InsertSubscription(ctx context.Context, body Subscription, parts []string) (*Subscription, error) {
	if len(parts) == 0 {
		parts = []string{SubscriptionPartSnippet}
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &Subscription{}
	if err := c.doJSON(ctx, http.MethodPost, "subscriptions", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteSubscription unsubscribes by deleting a subscription by its ID. Mirrors
// subscriptions.delete. Requires OAuth.
func (c *Client) DeleteSubscription(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("youtube: DeleteSubscription requires an ID")
	}
	q := url.Values{}
	q.Set("id", id)
	return c.doDelete(ctx, "subscriptions", q)
}
