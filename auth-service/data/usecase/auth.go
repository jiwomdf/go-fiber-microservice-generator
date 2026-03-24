package usecase

import (
	jwt "auth-service/data/repository/jtw"
	"auth-service/domain"
	"context"

	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	repo domain.AuthRepository
}

func NewAuthUsecase(repo domain.AuthRepository) domain.AuthUsecase {
	return &AuthUsecase{
		repo: repo,
	}
}

func (u *AuthUsecase) Login(c context.Context, req *domain.CreateAuthRequest) (*domain.LoginResponse, string, error) {
	result, status, err := u.repo.GetAuthByEmail(c, req.Email)
	if err != nil {
		if status == domain.StatusNotFound {
			return nil, domain.StatusInvalidEmailPassword, domain.ErrBadRequest
		}
		return nil, status, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(req.Password)); err != nil {
		return nil, domain.StatusInvalidEmailPassword, domain.ErrBadRequest
	}

	return jwt.BuildLoginResponse(result.Email)
}

func (u *AuthUsecase) GetAllAuths(c context.Context) ([]*domain.Auth, string, error) {
	auths, status, err := u.repo.GetAllAuths(c)
	if err != nil {
		return nil, status, err
	}
	return auths, domain.StatusSuccess, nil
}

func (u *AuthUsecase) CreateAuth(c context.Context, req *domain.CreateAuthRequest) (*domain.Auth, string, error) {
	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, domain.StatusInternalServerError, err
	}
	request := &domain.Auth{
		Email:    req.Email,
		Password: string(password),
	}
	result, status, err := u.repo.CreateAuth(c, request)
	if err != nil {
		if status == domain.StatusNotFound {
			return nil, domain.StatusInvalidEmailPassword, domain.ErrBadRequest
		}
		return nil, status, err
	}

	return result, domain.StatusSuccess, nil
}

func (u *AuthUsecase) DeleteAuth(c context.Context, id string) (string, error) {
	status, err := u.repo.DeleteAuth(c, id)
	if err != nil {
		return status, err
	}
	return domain.StatusSuccess, nil
}

func (u *AuthUsecase) GetAuthById(c context.Context, id string) (*domain.Auth, string, error) {
	auth, status, err := u.repo.GetAuthById(c, id)
	if err != nil {
		return nil, status, err
	}
	return auth, domain.StatusSuccess, nil
}

func (u *AuthUsecase) UpdateAuth(c context.Context, id string, req *domain.UpdateAuthRequest) (*domain.Auth, string, error) {
	res, status, err := u.repo.GetAuthById(c, id)
	if err != nil {
		return nil, status, err
	}

	updates := map[string]any{}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Password != "" {
		password, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err != nil {
			return nil, domain.StatusInternalServerError, err
		}
		updates["password"] = string(password)
	}

	res, status, err = u.repo.UpdateAuth(c, id, updates)
	if err != nil {
		return nil, status, err
	}
	return res, domain.StatusSuccess, nil
}
