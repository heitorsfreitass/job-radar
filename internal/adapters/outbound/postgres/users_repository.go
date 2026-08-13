package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

// UsersRepository implements domain.UserRepository on top of a Postgres pool.
type UsersRepository struct {
	pool *pgxpool.Pool
}

func NewUsersRepository(pool *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{pool: pool}
}

func (r *UsersRepository) Create(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	const q = `
		INSERT INTO users (email, password_hash) VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at
	`
	user, err := scanUser(r.pool.QueryRow(ctx, q, email, passwordHash))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return nil, domain.ErrEmailTaken
	}
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *UsersRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`
	user, err := scanUser(r.pool.QueryRow(ctx, q, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UsersRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`
	user, err := scanUser(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *UsersRepository) GetPreferences(ctx context.Context, userID string) (domain.Preferences, error) {
	const q = `SELECT country, workplace, seniority, tag, keyword FROM user_preferences WHERE user_id = $1`

	var p domain.Preferences
	err := r.pool.QueryRow(ctx, q, userID).Scan(&p.Country, &p.Workplace, &p.Seniority, &p.Tag, &p.Keyword)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Preferences{}, nil
	}
	if err != nil {
		return domain.Preferences{}, fmt.Errorf("get preferences: %w", err)
	}
	return p, nil
}

func (r *UsersRepository) SavePreferences(ctx context.Context, userID string, prefs domain.Preferences) error {
	const q = `
		INSERT INTO user_preferences (user_id, country, workplace, seniority, tag, keyword, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (user_id) DO UPDATE SET
			country = EXCLUDED.country,
			workplace = EXCLUDED.workplace,
			seniority = EXCLUDED.seniority,
			tag = EXCLUDED.tag,
			keyword = EXCLUDED.keyword,
			updated_at = now()
	`
	_, err := r.pool.Exec(ctx, q, userID, prefs.Country, prefs.Workplace, prefs.Seniority, prefs.Tag, prefs.Keyword)
	if err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}
	return nil
}

func scanUser(row rowScanner) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}
