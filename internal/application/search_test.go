package application

import (
	"context"
	"testing"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

type searchFakeRepo struct {
	gotFilter domain.JobFilter
	jobs      []*domain.Job
	total     int
}

func (f *searchFakeRepo) Upsert(ctx context.Context, job *domain.Job) (bool, error) {
	return false, nil
}

func (f *searchFakeRepo) Search(ctx context.Context, filter domain.JobFilter) ([]*domain.Job, int, error) {
	f.gotFilter = filter
	return f.jobs, f.total, nil
}

func (f *searchFakeRepo) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	if id == "found" {
		return &domain.Job{ID: "found"}, nil
	}
	return nil, nil
}

func TestSearchJobs_NormalizesPagination(t *testing.T) {
	repo := &searchFakeRepo{jobs: []*domain.Job{{ID: "1"}}, total: 1}

	result, err := SearchJobs(context.Background(), repo, domain.JobFilter{Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("SearchJobs() error = %v", err)
	}
	if result.Page != 1 || result.PageSize != 20 {
		t.Errorf("Page=%d PageSize=%d, want defaults 1/20", result.Page, result.PageSize)
	}
	if result.Total != 1 || len(result.Jobs) != 1 {
		t.Errorf("result = %+v, want Total=1 and 1 job", result)
	}
}

func TestSearchJobs_ClampsOversizedPageSize(t *testing.T) {
	repo := &searchFakeRepo{}

	result, err := SearchJobs(context.Background(), repo, domain.JobFilter{Page: 2, PageSize: 500})
	if err != nil {
		t.Fatalf("SearchJobs() error = %v", err)
	}
	if result.PageSize != 20 {
		t.Errorf("PageSize = %d, want clamped to 20", result.PageSize)
	}
	if result.Page != 2 {
		t.Errorf("Page = %d, want 2", result.Page)
	}
}

func TestGetJob_Found(t *testing.T) {
	repo := &searchFakeRepo{}

	job, err := GetJob(context.Background(), repo, "found")
	if err != nil {
		t.Fatalf("GetJob() error = %v", err)
	}
	if job.ID != "found" {
		t.Errorf("job.ID = %q, want \"found\"", job.ID)
	}
}

func TestGetJob_NotFound(t *testing.T) {
	repo := &searchFakeRepo{}

	_, err := GetJob(context.Background(), repo, "missing")
	if err != ErrJobNotFound {
		t.Errorf("GetJob() error = %v, want ErrJobNotFound", err)
	}
}
