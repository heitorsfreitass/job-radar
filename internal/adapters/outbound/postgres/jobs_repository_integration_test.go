//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("job_radar"),
		tcpostgres.WithUsername("job_radar"),
		tcpostgres.WithPassword("job_radar"),
		tcpostgres.WithInitScripts("../../../../migrations/0001_create_jobs.up.sql"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		panic("start postgres container: " + err.Error())
	}
	defer func() { _ = container.Terminate(ctx) }()

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("get connection string: " + err.Error())
	}

	pool, err := Connect(ctx, connStr)
	if err != nil {
		panic("connect: " + err.Error())
	}
	defer pool.Close()

	testPool = pool
	os.Exit(m.Run())
}

func newTestRepo(t *testing.T) *JobsRepository {
	t.Helper()
	t.Cleanup(func() {
		if _, err := testPool.Exec(context.Background(), "TRUNCATE TABLE jobs"); err != nil {
			t.Fatalf("truncate jobs: %v", err)
		}
	})
	return NewJobsRepository(testPool)
}

func sampleJob(sourceJobID string) *domain.Job {
	return &domain.Job{
		Title:       "Backend Engineer",
		CompanyName: "Acme",
		Description: "desc",
		URL:         "https://example.com/jobs/" + sourceJobID,
		Source:      domain.SourceArbeitnow,
		SourceJobID: sourceJobID,
		Country:     "Germany",
		Workplace:   domain.WorkplaceRemote,
		Employment:  domain.EmploymentTypeFullTime,
		Seniority:   domain.SenioritySenior,
		Tags:        []string{"go", "backend"},
		PublishedAt: time.Now().UTC().Truncate(time.Second),
	}
}

func TestUpsert_InsertsNewJob(t *testing.T) {
	repo := newTestRepo(t)
	job := sampleJob("job-1")

	inserted, err := repo.Upsert(context.Background(), job)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if !inserted {
		t.Error("inserted = false, want true for a brand new job")
	}
	if job.ID == "" {
		t.Error("job.ID not populated after insert")
	}
}

func TestUpsert_SameURLUpdatesInPlace(t *testing.T) {
	repo := newTestRepo(t)
	job := sampleJob("job-1")

	if _, err := repo.Upsert(context.Background(), job); err != nil {
		t.Fatalf("first Upsert() error = %v", err)
	}
	firstID := job.ID

	updated := sampleJob("job-1")
	updated.Title = "Senior Backend Engineer"
	inserted, err := repo.Upsert(context.Background(), updated)
	if err != nil {
		t.Fatalf("second Upsert() error = %v", err)
	}
	if inserted {
		t.Error("inserted = true on re-upsert of the same URL, want false")
	}
	if updated.ID != firstID {
		t.Errorf("ID changed on re-upsert: got %q, want %q", updated.ID, firstID)
	}

	got, err := repo.GetByID(context.Background(), firstID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Title != "Senior Backend Engineer" {
		t.Errorf("Title = %q, want updated title", got.Title)
	}
}

// TestUpsert_SameSourceJobIDDifferentURL exercises the fallback dedup path:
// when a job is re-posted under a new URL but keeps the same
// (source, source_job_id), the unique constraint on url collides with a
// *different* row than the one matching source_job_id, so Upsert must fall
// back to updating by (source, source_job_id) instead of inserting a
// duplicate.
func TestUpsert_SameSourceJobIDDifferentURL(t *testing.T) {
	repo := newTestRepo(t)
	job := sampleJob("job-1")

	if _, err := repo.Upsert(context.Background(), job); err != nil {
		t.Fatalf("first Upsert() error = %v", err)
	}
	firstID := job.ID

	moved := sampleJob("job-1")
	moved.URL = "https://example.com/jobs/job-1-moved"
	moved.Title = "Relocated Posting"
	inserted, err := repo.Upsert(context.Background(), moved)
	if err != nil {
		t.Fatalf("second Upsert() error = %v", err)
	}
	if inserted {
		t.Error("inserted = true when source_job_id already existed, want false")
	}
	if moved.ID != firstID {
		t.Errorf("ID = %q, want the original row's ID %q (updated in place)", moved.ID, firstID)
	}

	_, total, err := repo.Search(context.Background(), domain.JobFilter{})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1 (no duplicate row created)", total)
	}
}

func TestSearch_FiltersAndPagination(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	remote := sampleJob("remote-1")
	remote.Workplace = domain.WorkplaceRemote
	remote.Seniority = domain.SenioritySenior
	remote.Country = "Germany"
	remote.Tags = []string{"go"}

	onsite := sampleJob("onsite-1")
	onsite.URL = "https://example.com/jobs/onsite-1"
	onsite.Workplace = domain.WorkplaceOnsite
	onsite.Seniority = domain.SeniorityJunior
	onsite.Country = "France"
	onsite.Tags = []string{"python"}

	for _, j := range []*domain.Job{remote, onsite} {
		if _, err := repo.Upsert(ctx, j); err != nil {
			t.Fatalf("Upsert(%s) error = %v", j.SourceJobID, err)
		}
	}

	jobs, total, err := repo.Search(ctx, domain.JobFilter{Workplace: domain.WorkplaceRemote})
	if err != nil {
		t.Fatalf("Search(workplace=remote) error = %v", err)
	}
	if total != 1 || len(jobs) != 1 || jobs[0].SourceJobID != "remote-1" {
		t.Errorf("Search(workplace=remote) = %+v (total=%d), want only remote-1", jobs, total)
	}

	jobs, total, err = repo.Search(ctx, domain.JobFilter{Tag: "python"})
	if err != nil {
		t.Fatalf("Search(tag=python) error = %v", err)
	}
	if total != 1 || len(jobs) != 1 || jobs[0].SourceJobID != "onsite-1" {
		t.Errorf("Search(tag=python) = %+v (total=%d), want only onsite-1", jobs, total)
	}

	jobs, total, err = repo.Search(ctx, domain.JobFilter{Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("Search(page_size=1) error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2 (pagination shouldn't affect the count)", total)
	}
	if len(jobs) != 1 {
		t.Errorf("len(jobs) = %d, want 1 (page_size clamp)", len(jobs))
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newTestRepo(t)

	job, err := repo.GetByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("GetByID() error = %v, want nil error for a missing row", err)
	}
	if job != nil {
		t.Errorf("GetByID() = %+v, want nil", job)
	}
}
