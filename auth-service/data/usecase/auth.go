package usecase

import (
	"auth-service/domain"
	"context"
)

type AuthUsecase struct {
	repo domain.AuthRepository
}

func NewAuthUsecase(repo domain.AuthRepository) domain.AuthUsecase {
	return &AuthUsecase{
		repo: repo,
	}
}

func (u *AuthUsecase) CreateAuth(c context.Context, req *domain.CreateAuthRequest) (*domain.Auth, string, error) {
	entity := &domain.Auth{
		Name:  req.Name,
		Email: req.Email,
	}

	res, status, err := u.repo.CreateAuth(c, entity)
	if err != nil {
		return nil, status, err
	}
	return res, domain.StatusSuccess, nil
}
