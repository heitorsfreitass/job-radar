//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

func newTestUsersRepo(t *testing.T) *UsersRepository {
	t.Helper()
	t.Cleanup(func() {
		if _, err := testPool.Exec(context.Background(), "TRUNCATE TABLE user_preferences, users"); err != nil {
			t.Fatalf("truncate users: %v", err)
		}
	})
	return NewUsersRepository(testPool)
}

func TestUsersCreate_InsertsAndReturnsUser(t *testing.T) {
	repo := newTestUsersRepo(t)

	user, err := repo.Create(context.Background(), "user@example.com", "hashed-password")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if user.ID == 0 {
		t.Error("ID not populated")
	}
	if user.Email != "user@example.com" {
		t.Errorf("Email = %q, want \"user@example.com\"", user.Email)
	}
	if user.PasswordHash != "hashed-password" {
		t.Errorf("PasswordHash = %q, want \"hashed-password\"", user.PasswordHash)
	}
}

func TestUsersCreate_DuplicateEmail(t *testing.T) {
	repo := newTestUsersRepo(t)
	ctx := context.Background()

	if _, err := repo.Create(ctx, "user@example.com", "hash-1"); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err := repo.Create(ctx, "user@example.com", "hash-2")
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("second Create() error = %v, want ErrEmailTaken", err)
	}
}

func TestUsersGetByEmail_NotFound(t *testing.T) {
	repo := newTestUsersRepo(t)

	user, err := repo.GetByEmail(context.Background(), "nobody@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v, want nil error for a missing row", err)
	}
	if user != nil {
		t.Errorf("GetByEmail() = %+v, want nil", user)
	}
}

func TestUsersGetByID_RoundTrips(t *testing.T) {
	repo := newTestUsersRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, "user@example.com", "hashed-password")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil || got.Email != "user@example.com" {
		t.Errorf("GetByID() = %+v, want a user with email user@example.com", got)
	}
}

func TestPreferences_DefaultsToZeroValueWhenUnset(t *testing.T) {
	repo := newTestUsersRepo(t)
	ctx := context.Background()

	user, err := repo.Create(ctx, "user@example.com", "hashed-password")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	prefs, err := repo.GetPreferences(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetPreferences() error = %v", err)
	}
	if prefs != (domain.Preferences{}) {
		t.Errorf("GetPreferences() = %+v, want zero value before any SavePreferences call", prefs)
	}
}

func TestPreferences_SaveThenUpdate(t *testing.T) {
	repo := newTestUsersRepo(t)
	ctx := context.Background()

	user, err := repo.Create(ctx, "user@example.com", "hashed-password")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	first := domain.Preferences{Country: "Germany", Workplace: domain.WorkplaceRemote, Seniority: domain.SenioritySenior, Tag: "go", Keyword: "backend"}
	if err := repo.SavePreferences(ctx, user.ID, first); err != nil {
		t.Fatalf("SavePreferences() #1 error = %v", err)
	}

	got, err := repo.GetPreferences(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetPreferences() error = %v", err)
	}
	if got != first {
		t.Errorf("GetPreferences() = %+v, want %+v", got, first)
	}

	second := domain.Preferences{Country: "France", Workplace: domain.WorkplaceOnsite}
	if err := repo.SavePreferences(ctx, user.ID, second); err != nil {
		t.Fatalf("SavePreferences() #2 error = %v", err)
	}

	got, err = repo.GetPreferences(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetPreferences() error = %v", err)
	}
	if got != second {
		t.Errorf("GetPreferences() after update = %+v, want %+v (overwritten, not merged)", got, second)
	}
}
