package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-service/data/delivery/http/handler"
	"auth-service/domain"

	"github.com/gofiber/fiber/v2"
)

type mockAuthUsecase struct {
	auths map[string]*domain.Auth
}

func newMockAuthUsecase() *mockAuthUsecase {
	return &mockAuthUsecase{
		auths: make(map[string]*domain.Auth),
	}
}

func (m *mockAuthUsecase) Login(c context.Context, req *domain.CreateAuthRequest) (*domain.LoginResponse, string, error) {
	if req.Email == "invalid@example.com" {
		return nil, domain.StatusInvalidEmailPassword, domain.ErrBadRequest
	}
	return &domain.LoginResponse{
		Token:   "dummy-jwt-token",
		Expired: "2099-01-01T00:00:00Z",
		Email:   req.Email,
	}, domain.StatusSuccessLogin, nil
}

func (m *mockAuthUsecase) CreateAuth(c context.Context, req *domain.CreateAuthRequest) (*domain.Auth, string, error) {
	if req.Email == "" {
		return nil, domain.StatusBadRequest, errors.New("empty email")
	}
	auth := &domain.Auth{
		Email:    req.Email,
		Password: req.Password,
	}
	m.auths[auth.Email] = auth
	return auth, domain.StatusSuccess, nil
}

func (m *mockAuthUsecase) GetAllAuths(c context.Context) ([]*domain.Auth, string, error) {
	var list []*domain.Auth
	for _, a := range m.auths {
		list = append(list, a)
	}
	return list, domain.StatusSuccess, nil
}

func (m *mockAuthUsecase) GetAuthById(c context.Context, id string) (*domain.Auth, string, error) {
	if a, ok := m.auths[id]; ok {
		return a, domain.StatusSuccess, nil
	}
	return nil, domain.StatusNotFound, errors.New("not found")
}

func (m *mockAuthUsecase) UpdateAuth(c context.Context, id string, req *domain.UpdateAuthRequest) (*domain.Auth, string, error) {
	a, ok := m.auths[id]
	if !ok {
		return nil, domain.StatusNotFound, errors.New("not found")
	}
	if req.Email != "" {
		a.Email = req.Email
	}
	return a, domain.StatusSuccess, nil
}

func (m *mockAuthUsecase) DeleteAuth(c context.Context, id string) (string, error) {
	if _, ok := m.auths[id]; ok {
		delete(m.auths, id)
		return domain.StatusSuccess, nil
	}
	return domain.StatusNotFound, errors.New("not found")
}

func setupAuthApp(uc domain.AuthUsecase) *fiber.App {
	app := fiber.New()
	h := handler.NewAuthHandler(uc)

	app.Post("/api/v1/login", h.Login)
	app.Post("/api/v1/auth", h.CreateAuth)
	app.Get("/api/v1/auth", h.GetAllAuths)
	app.Get("/api/v1/auth/:id", h.GetAuthById)
	app.Put("/api/v1/auth/:id", h.UpdateAuth)
	app.Delete("/api/v1/auth/:id", h.DeleteAuth)

	return app
}

func TestAuthHandler_CreateAuth(t *testing.T) {
	uc := newMockAuthUsecase()
	app := setupAuthApp(uc)

	body, _ := json.Marshal(domain.CreateAuthRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestAuthHandler_GetAllAuths(t *testing.T) {
	uc := newMockAuthUsecase()
	uc.auths["test@example.com"] = &domain.Auth{Email: "test@example.com"}
	app := setupAuthApp(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestAuthHandler_GetAuthById(t *testing.T) {
	uc := newMockAuthUsecase()
	uc.auths["test@example.com"] = &domain.Auth{Email: "test@example.com"}
	app := setupAuthApp(uc)

	// Found
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/test@example.com", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Not Found
	reqNotFound := httptest.NewRequest(http.MethodGet, "/api/v1/auth/non-existent", nil)
	respNotFound, err := app.Test(reqNotFound)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if respNotFound.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, respNotFound.StatusCode)
	}
}

func TestAuthHandler_UpdateAuth(t *testing.T) {
	uc := newMockAuthUsecase()
	uc.auths["test@example.com"] = &domain.Auth{Email: "test@example.com"}
	app := setupAuthApp(uc)

	body, _ := json.Marshal(domain.UpdateAuthRequest{
		Email: "updated@example.com",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/test@example.com", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestAuthHandler_DeleteAuth(t *testing.T) {
	uc := newMockAuthUsecase()
	uc.auths["test@example.com"] = &domain.Auth{Email: "test@example.com"}
	app := setupAuthApp(uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/test@example.com", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestAuthHandler_Login(t *testing.T) {
	uc := newMockAuthUsecase()
	app := setupAuthApp(uc)

	body, _ := json.Marshal(domain.CreateAuthRequest{
		Email:    "login@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

