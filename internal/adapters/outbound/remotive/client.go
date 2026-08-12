// Package remotive implements the JobSource port against Remotive's
// public job board API (https://remotive.com/api/remote-jobs).
//
// Remotive's terms of use (see the API's own "0-legal-notice" response
// field) ask API consumers to call this endpoint at most a few times a
// day and to keep attribution back to Remotive on every listing shown.
// Both rules are enforced here, not left implicit:
//   - Rate limit: Fetch itself makes exactly one HTTP call per invocation;
//     internal/scheduler is responsible for calling it at most
//     config.RemotiveMaxRequestsPerDay times per day.
//   - Attribution: mapper.go tags every job with "via Remotive" and keeps
//     the original apply URL untouched.
package remotive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

const defaultBaseURL = "https://remotive.com/api/remote-jobs"

type response struct {
	Jobs []rawJob `json:"jobs"`
}

// Client fetches and normalizes job postings from Remotive.
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
	return domain.SourceRemotive
}

// Fetch makes a single call to the Remotive API and returns its jobs
// normalized into domain.Job values. Callers (internal/scheduler) are
// responsible for respecting Remotive's daily request cap.
func (c *Client) Fetch(ctx context.Context) ([]*domain.Job, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build remotive request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call remotive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remotive returned status %d", resp.StatusCode)
	}

	var out response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode remotive response: %w", err)
	}

	jobs := make([]*domain.Job, 0, len(out.Jobs))
	for _, raw := range out.Jobs {
		jobs = append(jobs, raw.toDomain())
	}

	return jobs, nil
}
