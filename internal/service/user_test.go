package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wi1low/go-web-template/internal/model"
	"github.com/Wi1low/go-web-template/internal/repository"
)

// fakeRepo is the pattern for testing services without a database.
type fakeRepo struct {
	users  map[uint64]*model.User
	nextID uint64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: map[uint64]*model.User{}, nextID: 1}
}

func (r *fakeRepo) Create(_ context.Context, u *model.User) error {
	u.ID = r.nextID
	r.nextID++
	r.users[u.ID] = u
	return nil
}

func (r *fakeRepo) GetByID(_ context.Context, id uint64) (*model.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (r *fakeRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (r *fakeRepo) List(_ context.Context, offset, limit int) ([]model.User, int64, error) {
	var out []model.User
	for _, u := range r.users {
		out = append(out, *u)
	}
	return out, int64(len(out)), nil
}

func (r *fakeRepo) Update(_ context.Context, u *model.User) error {
	r.users[u.ID] = u
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, id uint64) error {
	if _, ok := r.users[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.users, id)
	return nil
}

func TestService(t *testing.T) {
	ctx := context.Background()
	svc := NewUserService(newFakeRepo())

	t.Run("create", func(t *testing.T) {
		u, err := svc.Create(ctx, CreateUserInput{Name: "alice", Email: "alice@example.com"})
		if err != nil || u.ID == 0 {
			t.Fatalf("Create() = (%v, %v)", u, err)
		}
	})

	t.Run("create duplicate email", func(t *testing.T) {
		_, err := svc.Create(ctx, CreateUserInput{Name: "bob", Email: "alice@example.com"})
		if !errors.Is(err, ErrEmailTaken) {
			t.Fatalf("Create() error = %v, want ErrEmailTaken", err)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := svc.Get(ctx, 999)
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("Get() error = %v, want ErrUserNotFound", err)
		}
	})

	t.Run("list rejects bad paging", func(t *testing.T) {
		if _, _, err := svc.List(ctx, 0, 10); !errors.Is(err, ErrInvalidPaging) {
			t.Fatalf("List() error = %v, want ErrInvalidPaging", err)
		}
		if _, _, err := svc.List(ctx, 1, 101); !errors.Is(err, ErrInvalidPaging) {
			t.Fatalf("List() error = %v, want ErrInvalidPaging", err)
		}
	})

	t.Run("update partial", func(t *testing.T) {
		u, err := svc.Create(ctx, CreateUserInput{Name: "carol", Email: "carol@example.com"})
		if err != nil {
			t.Fatal(err)
		}
		newName := "carol-2"
		got, err := svc.Update(ctx, u.ID, UpdateUserInput{Name: &newName})
		if err != nil || got.Name != newName || got.Email != "carol@example.com" {
			t.Fatalf("Update() = (%v, %v)", got, err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		u, _ := svc.Create(ctx, CreateUserInput{Name: "dave", Email: "dave@example.com"})
		if err := svc.Delete(ctx, u.ID); err != nil {
			t.Fatalf("Delete() = %v", err)
		}
		if err := svc.Delete(ctx, u.ID); !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("Delete() again = %v, want ErrUserNotFound", err)
		}
	})
}
