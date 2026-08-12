package scheduler

import (
	"context"
	"sync"
	"testing"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

// fakeCache implements domain.Cache in-memory, with no TTL expiry: good
// enough to simulate "stays within the same rate-limit window" across
// repeated calls in a single test.
type fakeCache struct {
	mu     sync.Mutex
	counts map[string]int64
}

func newFakeCache() *fakeCache {
	return &fakeCache{counts: map[string]int64{}}
}

func (c *fakeCache) Get(ctx context.Context, key string) (string, error)              { return "", nil }
func (c *fakeCache) Set(ctx context.Context, key, value string, ttlSeconds int) error { return nil }

func (c *fakeCache) Increment(ctx context.Context, key string, ttlSeconds int) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[key]++
	return c.counts[key], nil
}

type countingSource struct {
	mu    sync.Mutex
	name  domain.Source
	calls int
}

func (s *countingSource) Name() domain.Source { return s.name }

func (s *countingSource) Fetch(ctx context.Context) ([]*domain.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return nil, nil
}

func (s *countingSource) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

type noopRepo struct{}

func (noopRepo) Upsert(ctx context.Context, job *domain.Job) (bool, error) { return false, nil }
func (noopRepo) Search(ctx context.Context, filter domain.JobFilter) ([]*domain.Job, int, error) {
	return nil, 0, nil
}
func (noopRepo) GetByID(ctx context.Context, id string) (*domain.Job, error) { return nil, nil }

func TestIngestRemotive_StopsAtDailyCap(t *testing.T) {
	remotive := &countingSource{name: domain.SourceRemotive}
	arbeitnow := &countingSource{name: domain.SourceArbeitnow}
	cache := newFakeCache()

	const maxPerDay = 4
	s := New(noopRepo{}, cache, arbeitnow, remotive, maxPerDay)

	ctx := context.Background()
	for i := 0; i < maxPerDay+3; i++ {
		s.ingestRemotive(ctx)
	}

	if got := remotive.callCount(); got != maxPerDay {
		t.Errorf("remotive.Fetch called %d times, want exactly %d (the daily cap)", got, maxPerDay)
	}
}

func TestNew_FallsBackToDefaultCapWhenInvalid(t *testing.T) {
	s := New(noopRepo{}, newFakeCache(), &countingSource{}, &countingSource{}, 0)
	if s.remotiveMaxPerDay != fallbackRemotiveMaxDay {
		t.Errorf("remotiveMaxPerDay = %d, want fallback %d", s.remotiveMaxPerDay, fallbackRemotiveMaxDay)
	}
}
