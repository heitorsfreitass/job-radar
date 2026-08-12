// Package application holds the use cases that orchestrate the domain
// through its ports. It has no dependency on HTTP, Postgres, Redis, or
// any specific upstream API — only on the domain package's interfaces.
package application

import (
	"context"
	"fmt"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

// IngestResult summarizes the outcome of one ingestion run against a
// single domain.JobSource.
type IngestResult struct {
	Source   domain.Source
	Fetched  int
	Inserted int
	Updated  int
}

// IngestJobs fetches every job currently offered by source and upserts
// each one into repo, deduplicating as defined by JobRepository.Upsert.
// A failure to upsert one job is recorded but does not abort the run.
func IngestJobs(ctx context.Context, source domain.JobSource, repo domain.JobRepository) (IngestResult, error) {
	result := IngestResult{Source: source.Name()}

	jobs, err := source.Fetch(ctx)
	if err != nil {
		return result, fmt.Errorf("fetch from %s: %w", source.Name(), err)
	}
	result.Fetched = len(jobs)

	var firstErr error
	for _, job := range jobs {
		inserted, err := repo.Upsert(ctx, job)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("upsert job %q from %s: %w", job.SourceJobID, source.Name(), err)
			}
			continue
		}
		if inserted {
			result.Inserted++
		} else {
			result.Updated++
		}
	}

	return result, firstErr
}
