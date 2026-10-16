package test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"user-service/data/usecase"
	"user-service/domain"
)

type mockUserRepo struct {
	users  map[int64]*domain.User
	nextID int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[int64]*domain.User),
		nextID: 1,
	}
}

func (m *mockUserRepo) CreateUser(c context.Context, user *domain.User) (*domain.User, string, error) {
	if user.Name == "" {
		return nil, domain.StatusBadRequest, errors.New("empty name")
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	return user, domain.StatusSuccess, nil
}

func (m *mockUserRepo) GetUserById(c context.Context, id string) (*domain.User, string, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, domain.StatusBadRequest, err
	}
	if u, ok := m.users[intID]; ok {
		return u, domain.StatusSuccess, nil
	}
	return nil, domain.StatusNotFound, errors.New("not found")
}

func (m *mockUserRepo) GetAllUsers(c context.Context) ([]*domain.User, string, error) {
	var list []*domain.User
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, domain.StatusSuccess, nil
}

func (m *mockUserRepo) UpdateUser(c context.Context, id string, updates map[string]any) (*domain.User, string, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, domain.StatusBadRequest, err
	}
	u, ok := m.users[intID]
	if !ok {
		return nil, domain.StatusNotFound, errors.New("not found")
	}
	if name, ok := updates["name"].(string); ok {
		u.Name = name
	}
	if email, ok := updates["email"].(string); ok {
		u.Email = email
	}
	return u, domain.StatusSuccess, nil
}

func (m *mockUserRepo) DeleteUser(c context.Context, id string) (string, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return domain.StatusBadRequest, err
	}
	if _, ok := m.users[intID]; ok {
		delete(m.users, intID)
		return domain.StatusSuccess, nil
	}
	return domain.StatusNotFound, errors.New("not found")
}

func TestUserUsecase_CreateUser(t *testing.T) {
	repo := newMockUserRepo()
	uc := usecase.NewUserUsecase(repo)
	ctx := context.Background()

	req := &domain.CreateUserRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	}

	res, status, err := uc.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.StatusSuccess {
		t.Errorf("expected status %s, got %s", domain.StatusSuccess, status)
	}
	if res.Name != req.Name || res.Email != req.Email {
		t.Errorf("expected %s / %s, got %s / %s", req.Name, req.Email, res.Name, res.Email)
	}
}

func TestUserUsecase_GetUserById(t *testing.T) {
	repo := newMockUserRepo()
	uc := usecase.NewUserUsecase(repo)
	ctx := context.Background()

	// Not found
	_, _, err := uc.GetUserById(ctx, "999")
	if err == nil {
		t.Errorf("expected error for non-existent user")
	}

	// Create and find
	created, _, _ := uc.CreateUser(ctx, &domain.CreateUserRequest{
		Name:  "Bob",
		Email: "bob@example.com",
	})

	found, status, err := uc.GetUserById(ctx, fmt.Sprintf("%d", created.ID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.StatusSuccess {
		t.Errorf("expected status %s, got %s", domain.StatusSuccess, status)
	}
	if found.Name != "Bob" {
		t.Errorf("expected name Bob, got %s", found.Name)
	}
}

func TestUserUsecase_GetAllUsers(t *testing.T) {
	repo := newMockUserRepo()
	uc := usecase.NewUserUsecase(repo)
	ctx := context.Background()

	_, _, _ = uc.CreateUser(ctx, &domain.CreateUserRequest{Name: "User 1", Email: "u1@example.com"})
	_, _, _ = uc.CreateUser(ctx, &domain.CreateUserRequest{Name: "User 2", Email: "u2@example.com"})

	list, status, err := uc.GetAllUsers(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.StatusSuccess {
		t.Errorf("expected status %s, got %s", domain.StatusSuccess, status)
	}
	if len(list) < 2 {
		t.Errorf("expected at least 2 users, got %d", len(list))
	}
}

func TestUserUsecase_UpdateUser(t *testing.T) {
	repo := newMockUserRepo()
	uc := usecase.NewUserUsecase(repo)
	ctx := context.Background()

	created, _, _ := uc.CreateUser(ctx, &domain.CreateUserRequest{
		Name:  "Original",
		Email: "orig@example.com",
	})

	idStr := fmt.Sprintf("%d", created.ID)
	updated, status, err := uc.UpdateUser(ctx, idStr, &domain.UpdateUserRequest{
		Name: "Updated Name",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.StatusSuccess {
		t.Errorf("expected status %s, got %s", domain.StatusSuccess, status)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected updated name, got %s", updated.Name)
	}
}

func TestUserUsecase_DeleteUser(t *testing.T) {
	repo := newMockUserRepo()
	uc := usecase.NewUserUsecase(repo)
	ctx := context.Background()

	created, _, _ := uc.CreateUser(ctx, &domain.CreateUserRequest{
		Name:  "To Delete",
		Email: "del@example.com",
	})

	idStr := fmt.Sprintf("%d", created.ID)
	status, err := uc.DeleteUser(ctx, idStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.StatusSuccess {
		t.Errorf("expected status %s, got %s", domain.StatusSuccess, status)
	}

	_, _, err = uc.GetUserById(ctx, idStr)
	if err == nil {
		t.Errorf("expected error getting deleted user")
	}
}

