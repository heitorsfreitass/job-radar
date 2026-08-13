package domain

import "context"

// JobFilter carries the optional criteria used to search jobs.
type JobFilter struct {
	Country   string
	Workplace WorkplaceType
	Seniority SeniorityLevel
	Tag       string
	Keyword   string
	Page      int
	PageSize  int
}

// JobRepository is the outbound port for persisting and querying jobs.
// Implemented by internal/adapters/outbound/postgres.
type JobRepository interface {
	// Upsert stores a job, deduplicating on URL. It returns true when a
	// new row was inserted and false when an existing job was matched.
	Upsert(ctx context.Context, job *Job) (inserted bool, err error)
	Search(ctx context.Context, filter JobFilter) ([]*Job, int, error)
	GetByID(ctx context.Context, id string) (*Job, error)
}

// JobSource is the outbound port implemented by each upstream job board
// adapter (Arbeitnow, Remotive, ...). Fetch returns already-normalized
// domain.Job values.
type JobSource interface {
	Name() Source
	Fetch(ctx context.Context) ([]*Job, error)
}

// UserRepository is the outbound port for user accounts and their saved
// search preferences. Implemented by internal/adapters/outbound/postgres.
type UserRepository interface {
	// Create inserts a new user. It returns an error satisfying
	// errors.Is(err, ErrEmailTaken) when the email is already registered.
	Create(ctx context.Context, email, passwordHash string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetPreferences(ctx context.Context, userID int64) (Preferences, error)
	SavePreferences(ctx context.Context, userID int64, prefs Preferences) error
}

// Cache is the outbound port used for response caching and distributed
// rate limiting, backed by Redis.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttlSeconds int) error
	// Increment atomically increments key and returns the new value,
	// setting ttlSeconds on the key's first increment. Used for rate
	// limiting both inbound API traffic and outbound calls to Remotive.
	Increment(ctx context.Context, key string, ttlSeconds int) (int64, error)
}
