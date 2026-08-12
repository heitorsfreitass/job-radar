// Package scheduler runs the worker's periodic ingestion loops. It is the
// single place responsible for enforcing per-source request limits (in
// particular Remotive's daily cap), so ingestion adapters themselves stay
// free of scheduling concerns.
package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/heitorsfreitass/job-radar/internal/application"
	"github.com/heitorsfreitass/job-radar/internal/domain"
)

const (
	// arbeitnowInterval matches Arbeitnow's own update cadence ("Jobs are
	// updated every hour", per the API's `meta.info` field), so polling
	// faster would only ever return unchanged data.
	arbeitnowInterval = 1 * time.Hour

	remotiveWindow         = 24 * time.Hour
	remotiveRateLimitKey   = "remotive:requests:window"
	minRemotiveMaxPerDay   = 1
	fallbackRemotiveMaxDay = 4
)

// Scheduler owns the two ingestion loops (Arbeitnow, Remotive) and the
// Remotive daily rate cap.
type Scheduler struct {
	repo      domain.JobRepository
	cache     domain.Cache
	arbeitnow domain.JobSource
	remotive  domain.JobSource

	remotiveMaxPerDay int
	remotiveInterval  time.Duration
}

func New(repo domain.JobRepository, cache domain.Cache, arbeitnow, remotive domain.JobSource, remotiveMaxPerDay int) *Scheduler {
	if remotiveMaxPerDay < minRemotiveMaxPerDay {
		remotiveMaxPerDay = fallbackRemotiveMaxDay
	}

	return &Scheduler{
		repo:              repo,
		cache:             cache,
		arbeitnow:         arbeitnow,
		remotive:          remotive,
		remotiveMaxPerDay: remotiveMaxPerDay,
		// Spread the daily cap evenly across the window (e.g. 4/day -> one
		// attempt every 6h) instead of front-loading every call.
		remotiveInterval: remotiveWindow / time.Duration(remotiveMaxPerDay),
	}
}

// Run blocks, driving both ingestion loops until ctx is canceled.
func (s *Scheduler) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		s.loop(ctx, arbeitnowInterval, s.ingestArbeitnow)
	}()
	go func() {
		defer wg.Done()
		s.loop(ctx, s.remotiveInterval, s.ingestRemotive)
	}()

	wg.Wait()
}

func (s *Scheduler) loop(ctx context.Context, interval time.Duration, run func(context.Context)) {
	run(ctx) // ingest once immediately on startup, then on every tick

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run(ctx)
		}
	}
}

func (s *Scheduler) ingestArbeitnow(ctx context.Context) {
	result, err := application.IngestJobs(ctx, s.arbeitnow, s.repo)
	if err != nil {
		log.Printf("arbeitnow ingest error: %v", err)
	}
	log.Printf("arbeitnow ingest: fetched=%d inserted=%d updated=%d", result.Fetched, result.Inserted, result.Updated)
}

// ingestRemotive enforces the daily request cap before calling Remotive:
// it atomically increments a Redis counter bucketed to a fixed 24h window
// and only proceeds if the resulting count is still within
// remotiveMaxPerDay. This is the enforcement point named in the README's
// "Rate limit" rule (config.RemotiveMaxRequestsPerDay).
func (s *Scheduler) ingestRemotive(ctx context.Context) {
	count, err := s.cache.Increment(ctx, remotiveRateLimitKey, int(remotiveWindow.Seconds()))
	if err != nil {
		log.Printf("remotive rate-limit check failed, skipping ingest: %v", err)
		return
	}
	if count > int64(s.remotiveMaxPerDay) {
		log.Printf("remotive daily cap reached (%d/%d), skipping ingest", count, s.remotiveMaxPerDay)
		return
	}

	result, err := application.IngestJobs(ctx, s.remotive, s.repo)
	if err != nil {
		log.Printf("remotive ingest error: %v", err)
	}
	log.Printf("remotive ingest (%d/%d today): fetched=%d inserted=%d updated=%d", count, s.remotiveMaxPerDay, result.Fetched, result.Inserted, result.Updated)
}
