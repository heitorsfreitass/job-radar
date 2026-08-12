package application

import (
	"context"
	"errors"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

// ErrJobNotFound is returned by GetJob when no job matches the given id.
var ErrJobNotFound = errors.New("job not found")

// SearchResult carries a page of jobs plus enough metadata for the
// caller to build pagination info.
type SearchResult struct {
	Jobs     []*domain.Job
	Total    int
	Page     int
	PageSize int
}

// SearchJobs runs filter against repo and normalizes the page/page size
// that were actually applied (JobRepository.Search clamps invalid
// values) into the returned SearchResult.
func SearchJobs(ctx context.Context, repo domain.JobRepository, filter domain.JobFilter) (SearchResult, error) {
	jobs, total, err := repo.Search(ctx, filter)
	if err != nil {
		return SearchResult{}, err
	}

	page, pageSize := filter.Page, filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return SearchResult{Jobs: jobs, Total: total, Page: page, PageSize: pageSize}, nil
}

// GetJob fetches a single job by id, returning ErrJobNotFound when no
// such job exists.
func GetJob(ctx context.Context, repo domain.JobRepository, id string) (*domain.Job, error) {
	job, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, ErrJobNotFound
	}
	return job, nil
}
