package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"user-service/data/delivery/http/handler"
	"user-service/domain"

	"github.com/gofiber/fiber/v2"
)

type mockUserUsecase struct {
	users  map[int64]*domain.User
	nextID int64
}

func newMockUserUsecase() *mockUserUsecase {
	return &mockUserUsecase{
		users:  make(map[int64]*domain.User),
		nextID: 1,
	}
}

func (m *mockUserUsecase) CreateUser(c context.Context, req *domain.CreateUserRequest) (*domain.User, string, error) {
	if req.Name == "" || req.Email == "" {
		return nil, domain.StatusBadRequest, errors.New("missing required fields")
	}
	user := &domain.User{
		ID:    m.nextID,
		Name:  req.Name,
		Email: req.Email,
	}
	m.nextID++
	m.users[user.ID] = user
	return user, domain.StatusSuccess, nil
}

func (m *mockUserUsecase) GetAllUsers(c context.Context) ([]*domain.User, string, error) {
	var list []*domain.User
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, domain.StatusSuccess, nil
}

func (m *mockUserUsecase) GetUserById(c context.Context, id string) (*domain.User, string, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, domain.StatusBadRequest, err
	}
	if u, ok := m.users[intID]; ok {
		return u, domain.StatusSuccess, nil
	}
	return nil, domain.StatusNotFound, errors.New("not found")
}

func (m *mockUserUsecase) UpdateUser(c context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, string, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, domain.StatusBadRequest, err
	}
	u, ok := m.users[intID]
	if !ok {
		return nil, domain.StatusNotFound, errors.New("not found")
	}
	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	return u, domain.StatusSuccess, nil
}

func (m *mockUserUsecase) DeleteUser(c context.Context, id string) (string, error) {
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

func setupUserApp(uc domain.UserUsecase) *fiber.App {
	app := fiber.New()
	h := handler.NewUserHandler(uc)

	app.Post("/api/v1/user", h.CreateUser)
	app.Get("/api/v1/user", h.GetAllUsers)
	app.Get("/api/v1/user/:id", h.GetUserById)
	app.Put("/api/v1/user/:id", h.UpdateUser)
	app.Delete("/api/v1/user/:id", h.DeleteUser)

	return app
}

func TestUserHandler_CreateUser(t *testing.T) {
	uc := newMockUserUsecase()
	app := setupUserApp(uc)

	body, _ := json.Marshal(domain.CreateUserRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	uc := newMockUserUsecase()
	uc.users[1] = &domain.User{ID: 1, Name: "User 1", Email: "u1@example.com"}
	app := setupUserApp(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestUserHandler_GetUserById(t *testing.T) {
	uc := newMockUserUsecase()
	uc.users[1] = &domain.User{ID: 1, Name: "User 1", Email: "u1@example.com"}
	app := setupUserApp(uc)

	// Found
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Not Found
	reqNotFound := httptest.NewRequest(http.MethodGet, "/api/v1/user/999", nil)
	respNotFound, err := app.Test(reqNotFound)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if respNotFound.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, respNotFound.StatusCode)
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	uc := newMockUserUsecase()
	uc.users[1] = &domain.User{ID: 1, Name: "Original", Email: "orig@example.com"}
	app := setupUserApp(uc)

	body, _ := json.Marshal(domain.UpdateUserRequest{
		Name:  "Updated",
		Email: "updated@example.com",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/user/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	uc := newMockUserUsecase()
	uc.users[1] = &domain.User{ID: 1, Name: "To Delete", Email: "del@example.com"}
	app := setupUserApp(uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user/1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

