package application

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

type fakeUserRepo struct {
	byEmail map[string]*domain.User
	prefs   map[int64]domain.Preferences
	nextID  int64
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]*domain.User{}, prefs: map[int64]domain.Preferences{}}
}

func (f *fakeUserRepo) Create(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	if _, exists := f.byEmail[email]; exists {
		return nil, domain.ErrEmailTaken
	}
	f.nextID++
	user := &domain.User{ID: f.nextID, Email: email, PasswordHash: passwordHash}
	f.byEmail[email] = user
	return user, nil
}

func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return f.byEmail[email], nil
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	for _, u := range f.byEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (f *fakeUserRepo) GetPreferences(ctx context.Context, userID int64) (domain.Preferences, error) {
	return f.prefs[userID], nil
}

func (f *fakeUserRepo) SavePreferences(ctx context.Context, userID int64, prefs domain.Preferences) error {
	f.prefs[userID] = prefs
	return nil
}

func TestRegister_HashesPasswordAndNormalizesEmail(t *testing.T) {
	repo := newFakeUserRepo()

	user, err := Register(context.Background(), repo, "  User@Example.com ", "hunter22")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.Email != "user@example.com" {
		t.Errorf("Email = %q, want normalized \"user@example.com\"", user.Email)
	}
	if user.PasswordHash == "hunter22" {
		t.Error("PasswordHash stores the plaintext password, want a bcrypt hash")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("hunter22")) != nil {
		t.Error("stored hash does not verify against the original password")
	}
}

func TestRegister_RejectsShortPassword(t *testing.T) {
	repo := newFakeUserRepo()

	_, err := Register(context.Background(), repo, "user@example.com", "short")
	if !errors.Is(err, ErrWeakPassword) {
		t.Errorf("Register() error = %v, want ErrWeakPassword", err)
	}
}

func TestLogin_Succeeds(t *testing.T) {
	repo := newFakeUserRepo()
	if _, err := Register(context.Background(), repo, "user@example.com", "correct-password"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, err := Login(context.Background(), repo, "USER@example.com", "correct-password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if user.Email != "user@example.com" {
		t.Errorf("Email = %q, want \"user@example.com\"", user.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newFakeUserRepo()
	if _, err := Register(context.Background(), repo, "user@example.com", "correct-password"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, err := Login(context.Background(), repo, "user@example.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	repo := newFakeUserRepo()

	_, err := Login(context.Background(), repo, "nobody@example.com", "whatever1")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() error = %v, want ErrInvalidCredentials (not a distinct \"not found\" error)", err)
	}
}

func TestPreferences_SaveAndGet(t *testing.T) {
	repo := newFakeUserRepo()
	prefs := domain.Preferences{Country: "Germany", Workplace: domain.WorkplaceRemote, Seniority: domain.SenioritySenior}

	if err := SavePreferences(context.Background(), repo, 1, prefs); err != nil {
		t.Fatalf("SavePreferences() error = %v", err)
	}
	got, err := GetPreferences(context.Background(), repo, 1)
	if err != nil {
		t.Fatalf("GetPreferences() error = %v", err)
	}
	if got != prefs {
		t.Errorf("GetPreferences() = %+v, want %+v", got, prefs)
	}
}
