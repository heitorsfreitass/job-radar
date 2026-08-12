package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

const uniqueViolation = "23505"

// JobsRepository implements domain.JobRepository on top of a Postgres pool.
type JobsRepository struct {
	pool *pgxpool.Pool
}

func NewJobsRepository(pool *pgxpool.Pool) *JobsRepository {
	return &JobsRepository{pool: pool}
}

// Upsert stores a job. The jobs table enforces two independent dedup
// guards (see migrations/0001_create_jobs.up.sql): a unique URL, and a
// unique (source, source_job_id) pair. The URL is the primary dedup key
// used by ON CONFLICT below; if a row instead collides on the
// (source, source_job_id) guard (same upstream job re-posted under a new
// URL), the fallback path updates that existing row instead.
func (r *JobsRepository) Upsert(ctx context.Context, job *domain.Job) (bool, error) {
	inserted, err := r.upsertByURL(ctx, job)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "jobs_source_job_unique" {
		return r.updateBySourceJobID(ctx, job)
	}
	return inserted, err
}

func (r *JobsRepository) upsertByURL(ctx context.Context, job *domain.Job) (bool, error) {
	const q = `
		INSERT INTO jobs (
			title, company_name, description, url, source, source_job_id,
			country, workplace, employment, seniority, tags, published_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (url) DO UPDATE SET
			title = EXCLUDED.title,
			company_name = EXCLUDED.company_name,
			description = EXCLUDED.description,
			source = EXCLUDED.source,
			source_job_id = EXCLUDED.source_job_id,
			country = EXCLUDED.country,
			workplace = EXCLUDED.workplace,
			employment = EXCLUDED.employment,
			seniority = EXCLUDED.seniority,
			tags = EXCLUDED.tags,
			published_at = EXCLUDED.published_at
		RETURNING id, (xmax = 0) AS inserted
	`
	var inserted bool
	err := r.pool.QueryRow(ctx, q,
		job.Title, job.CompanyName, job.Description, job.URL, job.Source, job.SourceJobID,
		job.Country, job.Workplace, job.Employment, job.Seniority, job.Tags, job.PublishedAt,
	).Scan(&job.ID, &inserted)
	if err != nil {
		return false, fmt.Errorf("upsert job by url: %w", err)
	}
	return inserted, nil
}

func (r *JobsRepository) updateBySourceJobID(ctx context.Context, job *domain.Job) (bool, error) {
	const q = `
		UPDATE jobs SET
			title = $1, company_name = $2, description = $3, url = $4,
			country = $5, workplace = $6, employment = $7, seniority = $8,
			tags = $9, published_at = $10
		WHERE source = $11 AND source_job_id = $12
		RETURNING id
	`
	err := r.pool.QueryRow(ctx, q,
		job.Title, job.CompanyName, job.Description, job.URL,
		job.Country, job.Workplace, job.Employment, job.Seniority,
		job.Tags, job.PublishedAt, job.Source, job.SourceJobID,
	).Scan(&job.ID)
	if err != nil {
		return false, fmt.Errorf("update job by source_job_id: %w", err)
	}
	return false, nil
}

func (r *JobsRepository) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	const q = `
		SELECT id, title, company_name, description, url, source, source_job_id,
		       country, workplace, employment, seniority, tags, published_at, ingested_at
		FROM jobs WHERE id = $1
	`
	job, err := scanJob(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get job by id: %w", err)
	}
	return job, nil
}

// Search returns jobs matching filter, ordered newest-first, along with
// the total count of matching rows (ignoring pagination) for the caller
// to build pagination metadata.
func (r *JobsRepository) Search(ctx context.Context, filter domain.JobFilter) ([]*domain.Job, int, error) {
	var where []string
	var args []any

	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if filter.Country != "" {
		where = append(where, "country = "+arg(filter.Country))
	}
	if filter.Workplace != "" {
		where = append(where, "workplace = "+arg(filter.Workplace))
	}
	if filter.Seniority != "" {
		where = append(where, "seniority = "+arg(filter.Seniority))
	}
	if filter.Tag != "" {
		where = append(where, arg(filter.Tag)+" = ANY(tags)")
	}
	if filter.Keyword != "" {
		p := arg("%" + filter.Keyword + "%")
		where = append(where, "(title ILIKE "+p+" OR description ILIKE "+p+")")
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	var total int
	countQ := "SELECT count(*) FROM jobs " + whereClause
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count jobs: %w", err)
	}

	page, pageSize := filter.Page, filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	q := fmt.Sprintf(`
		SELECT id, title, company_name, description, url, source, source_job_id,
		       country, workplace, employment, seniority, tags, published_at, ingested_at
		FROM jobs %s
		ORDER BY published_at DESC
		LIMIT %s OFFSET %s
	`, whereClause, arg(pageSize), arg(offset))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*domain.Job
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("search jobs: %w", err)
	}

	return jobs, total, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(row rowScanner) (*domain.Job, error) {
	var job domain.Job
	err := row.Scan(
		&job.ID, &job.Title, &job.CompanyName, &job.Description, &job.URL,
		&job.Source, &job.SourceJobID, &job.Country, &job.Workplace,
		&job.Employment, &job.Seniority, &job.Tags, &job.PublishedAt, &job.IngestedAt,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}
