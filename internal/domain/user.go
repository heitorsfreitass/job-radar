package domain

import (
	"errors"
	"time"
)

// ErrEmailTaken is returned by UserRepository.Create when the email is
// already registered.
var ErrEmailTaken = errors.New("email already registered")

// User is an account holder. PasswordHash is a bcrypt hash, never the
// plaintext password.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// Preferences is a user's saved default search area, applied by the
// frontend as the initial filters when they log in.
type Preferences struct {
	Country   string
	Workplace WorkplaceType
	Seniority SeniorityLevel
	Tag       string
	Keyword   string
}
