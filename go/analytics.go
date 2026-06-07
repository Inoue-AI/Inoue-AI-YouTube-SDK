package youtube

import (
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

// DefaultAnalyticsBaseURL is the production YouTube Analytics API v2 root.
const DefaultAnalyticsBaseURL = "https://youtubeanalytics.googleapis.com/v2"

// AnalyticsColumnHeader describes one column in an analytics report.
type AnalyticsColumnHeader struct {
	Name       string `json:"name"`
	ColumnType string `json:"columnType"`
	DataType   string `json:"dataType"`
}

// AnalyticsReport is the response from an analytics query. Rows is a matrix of
// untyped JSON values whose order matches ColumnHeaders; callers interpret each
// cell per its column's DataType.
type AnalyticsReport struct {
	Kind          string                  `json:"kind"`
	ColumnHeaders []AnalyticsColumnHeader `json:"columnHeaders"`
	Rows          [][]json.RawMessage     `json:"rows"`
}

// AnalyticsQueryParams configures AnalyticsQuery.
type AnalyticsQueryParams struct {
	// IDs identifies the channel or content owner. Defaults to
	// "channel==MINE".
	IDs string
	// StartDate / EndDate are inclusive YYYY-MM-DD bounds. Required.
	StartDate string
	EndDate   string
	// Metrics is the list of metrics to report (e.g. ["views","likes"]).
	// Required.
	Metrics []string
	// Dimensions is the optional list of dimensions (e.g. ["day"]).
	Dimensions []string
	// Filters is an optional semicolon-delimited filter expression.
	Filters string
	// Sort is the optional list of sort fields ("-" prefix for descending).
	Sort []string
	// MaxResults limits the number of rows (0 = API default).
	MaxResults int
	// StartIndex is the 1-based index of the first row (0 = API default).
	StartIndex int
	// Currency is the ISO 4217 code for revenue metrics.
	Currency string
}

// AnalyticsQuery runs a YouTube Analytics report query. Requires an OAuth
// access token (the Analytics API does not accept API keys).
func (c *Client) AnalyticsQuery(ctx context.Context, p AnalyticsQueryParams) (*AnalyticsReport, error) {
	if err := c.requireOAuth(); err != nil {
		return nil, err
	}
	if p.StartDate == "" || p.EndDate == "" {
		return nil, errors.New("youtube: AnalyticsQuery requires StartDate and EndDate")
	}
	if len(p.Metrics) == 0 {
		return nil, errors.New("youtube: AnalyticsQuery requires at least one metric")
	}

	ids := p.IDs
	if ids == "" {
		ids = "channel==MINE"
	}

	q := url.Values{}
	q.Set("ids", ids)
	q.Set("startDate", p.StartDate)
	q.Set("endDate", p.EndDate)
	q.Set("metrics", strings.Join(p.Metrics, ","))
	if len(p.Dimensions) > 0 {
		q.Set("dimensions", strings.Join(p.Dimensions, ","))
	}
	if p.Filters != "" {
		q.Set("filters", p.Filters)
	}
	if len(p.Sort) > 0 {
		q.Set("sort", strings.Join(p.Sort, ","))
	}
	if p.MaxResults > 0 {
		q.Set("maxResults", strconv.Itoa(p.MaxResults))
	}
	if p.StartIndex > 0 {
		q.Set("startIndex", strconv.Itoa(p.StartIndex))
	}
	if p.Currency != "" {
		q.Set("currency", p.Currency)
	}

	out := &AnalyticsReport{}
	if err := c.getAnalytics(ctx, "reports", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// getAnalytics issues a GET against the Analytics API host. The Analytics base
// URL is fixed (it is not overridable via ClientOptions today); tests exercise
// AnalyticsQuery via the request-building path with a stubbed transport.
func (c *Client) getAnalytics(ctx context.Context, path string, query url.Values, out any) error {
	base := c.analyticsBaseURL()
	fullURL := base + "/" + strings.TrimLeft(path, "/") + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("youtube: build analytics request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("youtube: analytics request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("youtube: read analytics response: %w", err)
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
			Message:    fmt.Sprintf("analytics returned non-JSON: %v", err),
			Body:       raw,
		}
	}
	return nil
}

// analyticsBaseURL returns the Analytics host, honoring an override stored on
// the client for tests.
func (c *Client) analyticsBaseURL() string {
	if c.analyticsBase != "" {
		return c.analyticsBase
	}
	return DefaultAnalyticsBaseURL
}
