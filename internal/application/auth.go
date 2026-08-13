package application

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
)

const minPasswordLength = 8

// Register creates a new user account. Passwords are hashed with bcrypt
// before ever reaching the repository; the plaintext password is not
// retained.
func Register(ctx context.Context, repo domain.UserRepository, email, password string) (*domain.User, error) {
	email = normalizeEmail(email)
	if len(password) < minPasswordLength {
		return nil, ErrWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return repo.Create(ctx, email, string(hash))
}

// Login verifies credentials against the stored bcrypt hash, returning
// ErrInvalidCredentials for either an unknown email or a wrong password
// (deliberately not distinguished, so the API doesn't leak which emails
// are registered).
func Login(ctx context.Context, repo domain.UserRepository, email, password string) (*domain.User, error) {
	user, err := repo.GetByEmail(ctx, normalizeEmail(email))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func GetPreferences(ctx context.Context, repo domain.UserRepository, userID string) (domain.Preferences, error) {
	return repo.GetPreferences(ctx, userID)
}

func SavePreferences(ctx context.Context, repo domain.UserRepository, userID string, prefs domain.Preferences) error {
	return repo.SavePreferences(ctx, userID, prefs)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
