// Package arbeitnow implements the JobSource port against Arbeitnow's
// public job-board API (https://www.arbeitnow.com/api/job-board-api).
package arbeitnow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

const (
	defaultBaseURL = "https://arbeitnow.com/api/job-board-api"
	// maxPages caps how many pages Fetch will follow via the API's own
	// `links.next` pagination, so a single ingestion run can't grow
	// unbounded if the upstream catalog grows very large.
	maxPages = 5
)

type response struct {
	Data  []rawJob `json:"data"`
	Links struct {
		Next *string `json:"next"`
	} `json:"links"`
}

// Client fetches and normalizes job postings from Arbeitnow.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{httpClient: httpClient, baseURL: defaultBaseURL}
}

func (c *Client) Name() domain.Source {
	return domain.SourceArbeitnow
}

// Fetch retrieves job postings across up to maxPages of results, following
// the API's `links.next` pagination cursor, and returns them normalized
// into domain.Job values.
func (c *Client) Fetch(ctx context.Context) ([]*domain.Job, error) {
	var jobs []*domain.Job

	url := c.baseURL
	for page := 0; page < maxPages && url != ""; page++ {
		resp, err := c.fetchPage(ctx, url)
		if err != nil {
			return nil, err
		}
		for _, raw := range resp.Data {
			jobs = append(jobs, raw.toDomain())
		}
		if resp.Links.Next == nil {
			break
		}
		url = *resp.Links.Next
	}

	return jobs, nil
}

func (c *Client) fetchPage(ctx context.Context, url string) (*response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build arbeitnow request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call arbeitnow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arbeitnow returned status %d", resp.StatusCode)
	}

	var out response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode arbeitnow response: %w", err)
	}

	return &out, nil
}
