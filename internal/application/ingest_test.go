package application

import (
	"context"
	"errors"
	"testing"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

type fakeSource struct {
	name domain.Source
	jobs []*domain.Job
	err  error
}

func (f *fakeSource) Name() domain.Source { return f.name }

func (f *fakeSource) Fetch(ctx context.Context) ([]*domain.Job, error) {
	return f.jobs, f.err
}

type fakeRepo struct {
	upserted  []*domain.Job
	insertIDs map[string]bool // SourceJobID -> already seen (so repeat Upsert reports "updated")
	upsertErr map[string]error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{insertIDs: map[string]bool{}, upsertErr: map[string]error{}}
}

func (f *fakeRepo) Upsert(ctx context.Context, job *domain.Job) (bool, error) {
	if err, ok := f.upsertErr[job.SourceJobID]; ok {
		return false, err
	}
	f.upserted = append(f.upserted, job)
	if f.insertIDs[job.SourceJobID] {
		return false, nil
	}
	f.insertIDs[job.SourceJobID] = true
	return true, nil
}

func (f *fakeRepo) Search(ctx context.Context, filter domain.JobFilter) ([]*domain.Job, int, error) {
	return nil, 0, nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	return nil, nil
}

func TestIngestJobs_InsertsNewJobs(t *testing.T) {
	source := &fakeSource{
		name: domain.SourceArbeitnow,
		jobs: []*domain.Job{
			{SourceJobID: "1", Title: "A"},
			{SourceJobID: "2", Title: "B"},
		},
	}
	repo := newFakeRepo()

	result, err := IngestJobs(context.Background(), source, repo)
	if err != nil {
		t.Fatalf("IngestJobs() error = %v", err)
	}
	if result.Fetched != 2 || result.Inserted != 2 || result.Updated != 0 {
		t.Errorf("result = %+v, want Fetched=2 Inserted=2 Updated=0", result)
	}
}

func TestIngestJobs_ReingestingSameJobCountsAsUpdated(t *testing.T) {
	source := &fakeSource{
		name: domain.SourceArbeitnow,
		jobs: []*domain.Job{{SourceJobID: "1", Title: "A"}},
	}
	repo := newFakeRepo()

	if _, err := IngestJobs(context.Background(), source, repo); err != nil {
		t.Fatalf("first IngestJobs() error = %v", err)
	}
	result, err := IngestJobs(context.Background(), source, repo)
	if err != nil {
		t.Fatalf("second IngestJobs() error = %v", err)
	}
	if result.Inserted != 0 || result.Updated != 1 {
		t.Errorf("result = %+v, want Inserted=0 Updated=1 on re-ingest", result)
	}
}

func TestIngestJobs_PartialUpsertFailureDoesNotAbortRun(t *testing.T) {
	source := &fakeSource{
		name: domain.SourceArbeitnow,
		jobs: []*domain.Job{
			{SourceJobID: "bad", Title: "A"},
			{SourceJobID: "good", Title: "B"},
		},
	}
	repo := newFakeRepo()
	repo.upsertErr["bad"] = errors.New("boom")

	result, err := IngestJobs(context.Background(), source, repo)
	if err == nil {
		t.Fatal("IngestJobs() error = nil, want error from failed upsert")
	}
	if result.Inserted != 1 {
		t.Errorf("Inserted = %d, want 1 (the job that didn't fail)", result.Inserted)
	}
}

func TestIngestJobs_FetchErrorPropagates(t *testing.T) {
	source := &fakeSource{name: domain.SourceRemotive, err: errors.New("upstream down")}
	repo := newFakeRepo()

	_, err := IngestJobs(context.Background(), source, repo)
	if err == nil {
		t.Fatal("IngestJobs() error = nil, want fetch error")
	}
}
