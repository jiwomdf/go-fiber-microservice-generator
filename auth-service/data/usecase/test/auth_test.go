package test

import (
	"context"
	"errors"
	"testing"

	"auth-service/data/usecase"
	"auth-service/domain"
	"auth-service/helper"

	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type mockAuthRepo struct {
	auths map[string]*domain.Auth
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{
		auths: make(map[string]*domain.Auth),
	}
}

func (m *mockAuthRepo) CreateAuth(c context.Context, auth *domain.Auth) (*domain.Auth, string, error) {
	if auth.Email == "" {
		return nil, domain.StatusBadRequest, errors.New("empty email")
	}
	m.auths[auth.Email] = auth
	return auth, domain.StatusSuccess, nil
}

func (m *mockAuthRepo) GetAuthById(c context.Context, id string) (*domain.Auth, string, error) {
	if a, ok := m.auths[id]; ok {
		return a, domain.StatusSuccess, nil
	}
	return nil, domain.StatusNotFound, errors.New("not found")
}

func (m *mockAuthRepo) GetAuthByEmail(c context.Context, email string) (*domain.Auth, string, error) {
	for _, a := range m.auths {
		if a.Email == email {
			return a, domain.StatusSuccess, nil
		}
	}
	return nil, domain.StatusNotFound, errors.New("not found")
}

func (m *mockAuthRepo) GetAllAuths(c context.Context) ([]*domain.Auth, string, error) {
	var list []*domain.Auth
	for _, a := range m.auths {
		list = append(list, a)
	}
	return list, domain.StatusSuccess, nil
}

func (m *mockAuthRepo) UpdateAuth(c context.Context, id string, updates map[string]any) (*domain.Auth, string, error) {
	a, ok := m.auths[id]
	if !ok {
		return nil, domain.StatusNotFound, errors.New("not found")
	}
	if email, ok := updates["email"].(string); ok {
		a.Email = email
	}
	if pass, ok := updates["password"].(string); ok {
		a.Password = pass
	}
	return a, domain.StatusSuccess, nil
}

func (m *mockAuthRepo) DeleteAuth(c context.Context, id string) (string, error) {
	if _, ok := m.auths[id]; ok {
		delete(m.auths, id)
		return domain.StatusSuccess, nil
	}
	return domain.StatusNotFound, errors.New("not found")
}

func TestAuthUsecase_CreateAuth(t *testing.T) {
	repo := newMockAuthRepo()
	uc := usecase.NewAuthUsecase(repo)

	ctx := context.Background()
	req := &domain.CreateAuthRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	res, status, err := uc.CreateAuth(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.StatusSuccess {
		t.Errorf("expected status %s, got %s", domain.StatusSuccess, status)
	}
	if res.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, res.Email)
	}
}

func TestAuthUsecase_GetAuthById(t *testing.T) {
	repo := newMockAuthRepo()
	uc := usecase.NewAuthUsecase(repo)
	ctx := context.Background()

	// Not found
	_, _, err := uc.GetAuthById(ctx, "non-existent")
	if err == nil {
		t.Errorf("expected error for non-existent ID")
	}

	// Create and get
	created, _, _ := uc.CreateAuth(ctx, &domain.CreateAuthRequest{
		Email:    "get@example.com",
		Password: "password123",
	})

	found, status, err := uc.GetAuthById(ctx, created.Email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.StatusSuccess {
		t.Errorf("expected success status, got %s", status)
	}
	if found.Email != created.Email {
		t.Errorf("expected email %s, got %s", created.Email, found.Email)
	}
}

func TestAuthUsecase_GetAllAuths(t *testing.T) {
	repo := newMockAuthRepo()
	uc := usecase.NewAuthUsecase(repo)
	ctx := context.Background()

	_, _, _ = uc.CreateAuth(ctx, &domain.CreateAuthRequest{Email: "user1@example.com", Password: "pwd"})
	_, _, _ = uc.CreateAuth(ctx, &domain.CreateAuthRequest{Email: "user2@example.com", Password: "pwd"})

	list, status, err := uc.GetAllAuths(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.StatusSuccess {
		t.Errorf("expected success status, got %s", status)
	}
	if len(list) < 2 {
		t.Errorf("expected at least 2 items, got %d", len(list))
	}
}

func TestAuthUsecase_UpdateAuth(t *testing.T) {
	repo := newMockAuthRepo()
	uc := usecase.NewAuthUsecase(repo)
	ctx := context.Background()

	created, _, _ := uc.CreateAuth(ctx, &domain.CreateAuthRequest{
		Email:    "original@example.com",
		Password: "oldpassword",
	})

	updated, status, err := uc.UpdateAuth(ctx, created.Email, &domain.UpdateAuthRequest{
		Email: "updated@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.StatusSuccess {
		t.Errorf("expected success status, got %s", status)
	}
	if updated.Email != "updated@example.com" {
		t.Errorf("expected email updated@example.com, got %s", updated.Email)
	}
}

func TestAuthUsecase_DeleteAuth(t *testing.T) {
	repo := newMockAuthRepo()
	uc := usecase.NewAuthUsecase(repo)
	ctx := context.Background()

	created, _, _ := uc.CreateAuth(ctx, &domain.CreateAuthRequest{
		Email:    "delete@example.com",
		Password: "pwd",
	})

	status, err := uc.DeleteAuth(ctx, created.Email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != domain.StatusSuccess {
		t.Errorf("expected success status, got %s", status)
	}

	_, _, err = uc.GetAuthById(ctx, created.Email)
	if err == nil {
		t.Errorf("expected error getting deleted auth")
	}
}

func TestAuthUsecase_Login(t *testing.T) {
	viper.Set(helper.JWT_SIGNING_METHOD, "HS256")
	viper.Set(helper.JWT_SIGNATURE_KEY, "test-secret-key-1234567890123456")
	viper.Set(helper.JWT_EXPARATION, 3600)
	viper.Set(helper.ClientID, "auth-service")

	repo := newMockAuthRepo()
	uc := usecase.NewAuthUsecase(repo)
	ctx := context.Background()

	password := "plainpassword"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	repo.auths["login@example.com"] = &domain.Auth{
		Email:    "login@example.com",
		Password: string(hashed),
	}

	// Successful login
	resp, status, err := uc.Login(ctx, &domain.CreateAuthRequest{
		Email:    "login@example.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("unexpected error on login: %v", err)
	}
	if status != domain.StatusSuccessLogin {
		t.Errorf("expected %s, got %s", domain.StatusSuccessLogin, status)
	}
	if resp.Token == "" {
		t.Errorf("expected non-empty token")
	}

	// Wrong password
	_, _, err = uc.Login(ctx, &domain.CreateAuthRequest{
		Email:    "login@example.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Errorf("expected error on invalid password")
	}
}

