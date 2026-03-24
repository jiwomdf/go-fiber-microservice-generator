package usecase

import (
	"context"
	"user-service/domain"
)

type UserUsecase struct {
	repo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) domain.UserUsecase {
	return &UserUsecase{
		repo: repo,
	}
}

func (u *UserUsecase) CreateUser(c context.Context, req *domain.CreateUserRequest) (*domain.User, string, error) {
	entity := &domain.User{
		Name:  req.Name,
		Email: req.Email,
	}

	res, status, err := u.repo.CreateUser(c, entity)
	if err != nil {
		return nil, status, err
	}
	return res, domain.StatusSuccess, nil
}

func (u *UserUsecase) DeleteUser(c context.Context, id string) (string, error) {
	status, err := u.repo.DeleteUser(c, id)
	if err != nil {
		return status, err
	}
	return domain.StatusSuccess, nil
}

func (u *UserUsecase) GetAllUsers(c context.Context) ([]*domain.User, string, error) {
	res, status, err := u.repo.GetAllUsers(c)
	if err != nil {
		return nil, status, err
	}
	return res, domain.StatusSuccess, nil
}

func (u *UserUsecase) GetUserById(c context.Context, id string) (*domain.User, string, error) {
	res, status, err := u.repo.GetUserById(c, id)
	if err != nil {
		return nil, status, err
	}
	return res, domain.StatusSuccess, nil
}

func (u *UserUsecase) UpdateUser(c context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, string, error) {
	res, status, err := u.repo.GetUserById(c, id)
	if err != nil {
		return nil, status, err
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	res, status, err = u.repo.UpdateUser(c, id, updates)
	if err != nil {
		return nil, status, err
	}
	return res, domain.StatusSuccess, nil
}
