# job-radar

An aggregator for international job postings (focused on the European
market). It ingests jobs from public job-board APIs, normalizes and
deduplicates them into a single Postgres-backed model, and exposes them
through a documented REST API with filtering and pagination.

This is a portfolio project. It is a work in progress — see
[Project status](#project-status) below.

## Architecture

The project follows a hexagonal (ports & adapters) architecture:

- **`internal/domain`** — the core `Job` model and the ports (interfaces)
  every adapter must satisfy. No dependency on HTTP, Postgres, Redis, or
  any external API.
- **`internal/application`** — use cases (search jobs, ingest jobs) that
  orchestrate the domain through its ports.
- **`internal/adapters/inbound/http`** — the REST API (chi router,
  handlers, rate-limiting middleware).
- **`internal/adapters/outbound/{arbeitnow,remotive}`** — one client +
  mapper per upstream job source, translating each source's raw payload
  into the normalized `domain.Job` model.
- **`internal/adapters/outbound/postgres`** — job persistence.
- **`internal/adapters/outbound/redis`** — response caching and rate
  limiting.
- **`internal/scheduler`** — the worker's cron loop, responsible for
  enforcing per-source request limits.
- **`cmd/api`** and **`cmd/worker`** — the two binaries: the HTTP API and
  the background ingestion worker.

This separation keeps the domain testable in isolation and makes each
upstream source or storage backend swappable without touching business
logic.

### Why chi and sqlc

- **chi** is a thin router built directly on `net/http`, not a full
  framework — it stays close to the Go standard library instead of
  imposing its own request/response abstractions.
- **sqlc** generates type-safe Go code from plain SQL, instead of hiding
  queries behind an ORM's query builder. Queries stay explicit,
  reviewable, and fast, and errors surface at compile time.

## Data sources and terms-of-use compliance

job-radar consumes two public, unauthenticated job-board APIs. Respecting
each source's usage terms is treated as a first-class requirement, not an
afterthought:

### Arbeitnow

- `GET https://arbeitnow.com/api/job-board-api`
- No authentication required; paginated (~100 results per page).
- Broad coverage of European jobs and aggregated ATS listings.

### Remotive

- `GET https://remotive.com/api/remote-jobs`
- No authentication required; supports `category` and `limit` query
  params.
- **Usage rules enforced explicitly in code, not left implicit:**
  1. **Rate limit** — the worker calls the Remotive endpoint **at most 4
     times per day**. This cap is not a side effect of a generic
     scheduler interval; it is an explicit, named limit
     (`RemotiveMaxRequestsPerDay` in `internal/config`, defaulting to
     `4`) enforced by `internal/scheduler` before every Remotive fetch.
  2. **Attribution** — every job returned by job-radar's API that
     originated from Remotive keeps its original `apply` URL untouched
     and is tagged with its source (`"via Remotive"`) so API consumers
     can see and follow the original listing.

  Both rules are also called out with a code comment at the call site in
  `internal/adapters/outbound/remotive/client.go`, so the intent is
  visible to anyone reading the source, not just this README.

## Tech stack

| Concern         | Choice                          |
|------------------|----------------------------------|
| Language         | Go                               |
| HTTP router      | chi                              |
| Database         | PostgreSQL (via `pgx`, queries via `sqlc`) |
| Cache / rate limit | Redis                          |
| Containerization | Docker + docker-compose          |
| CI               | GitHub Actions (lint, test, build) |
| API docs         | OpenAPI / Swagger                |

## Getting started

Requirements: Go 1.26+, Docker.

```bash
cp .env.example .env
make up          # starts postgres, redis, api, worker via docker-compose
```

Run migrations (requires the [golang-migrate](https://github.com/golang-migrate/migrate) CLI):

```bash
make migrate-up
```

Run locally without Docker:

```bash
go run ./cmd/api      # REST API, defaults to :8080
go run ./cmd/worker    # ingestion worker
```

Run tests:

```bash
make test
```

## API

Full spec: [`docs/openapi.yaml`](docs/openapi.yaml). Summary:

| Endpoint | Description |
|----------|--------------|
| `GET /healthz` | Liveness check |
| `GET /jobs` | Search jobs. Query params: `country`, `workplace`, `seniority`, `tag`, `keyword`, `page`, `page_size` (max 100) |
| `GET /jobs/{id}` | Get a single job by id |

All endpoints are rate-limited to 60 requests/minute per client IP
(Redis-backed, returns `429` with a `Retry-After` header when exceeded).

## Project status

Built incrementally, in public stages:

- [x] Stage 1 — project skeleton, domain model, config, database
      connection, docker-compose, initial docs.
- [x] Stage 2 — ingestion worker (Arbeitnow + Remotive clients,
      normalization, deduplication, scheduler with the Remotive rate
      limit).
- [x] Stage 3 — REST API (filters, pagination, get-by-id, Redis rate
      limiting, OpenAPI/Swagger docs).
- [ ] Stage 4 — integration tests against a live Postgres/Redis, and CI.
      (Unit tests for mappers, use cases, and the scheduler's rate cap
      already exist and run via `make test`.)

**Out of scope for now:** frontend, user authentication, production
deployment, additional job sources beyond Arbeitnow and Remotive.

## License

MIT
