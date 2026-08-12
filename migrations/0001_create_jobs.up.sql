CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           TEXT NOT NULL,
    company_name    TEXT NOT NULL,
    description     TEXT NOT NULL,
    url             TEXT NOT NULL,
    source          TEXT NOT NULL,
    source_job_id   TEXT NOT NULL,
    country         TEXT NOT NULL DEFAULT '',
    workplace       TEXT NOT NULL DEFAULT 'unknown',
    employment      TEXT NOT NULL DEFAULT 'unknown',
    seniority       TEXT NOT NULL DEFAULT 'unknown',
    tags            TEXT[] NOT NULL DEFAULT '{}',
    published_at    TIMESTAMPTZ NOT NULL,
    ingested_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Dedup guard: same URL should never be stored twice, regardless of
    -- source. Combined with the source_job_id uniqueness below, this
    -- covers both dedup strategies mentioned in the project scope
    -- (apply URL, and title+company as a fallback handled at the
    -- application layer for sources with unstable URLs).
    CONSTRAINT jobs_url_unique UNIQUE (url),
    CONSTRAINT jobs_source_job_unique UNIQUE (source, source_job_id)
);

CREATE INDEX IF NOT EXISTS idx_jobs_country ON jobs (country);
CREATE INDEX IF NOT EXISTS idx_jobs_workplace ON jobs (workplace);
CREATE INDEX IF NOT EXISTS idx_jobs_seniority ON jobs (seniority);
CREATE INDEX IF NOT EXISTS idx_jobs_published_at ON jobs (published_at DESC);
CREATE INDEX IF NOT EXISTS idx_jobs_tags ON jobs USING GIN (tags);
