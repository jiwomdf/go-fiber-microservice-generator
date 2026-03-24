package postgres

import (
	"context"
	"user-service/domain"

	"gorm.io/gorm"
)

type UserRepo struct {
	Conn *gorm.DB
}

func NewUserRepo(conn *gorm.DB) domain.UserRepository {
	return &UserRepo{Conn: conn}
}

func (u *UserRepo) CreateUser(ctx context.Context, req *domain.User) (response *domain.User, status string, err error) {
	result := u.Conn.WithContext(ctx).Create(req)
	err = result.Error
	if err != nil {
		status = domain.StatusInternalServerError
		return
	}
	response = req
	status = domain.StatusSuccessCreate
	return
}

func (u *UserRepo) DeleteUser(ctx context.Context, id string) (status string, err error) {
	result := u.Conn.WithContext(ctx).Delete(&domain.User{}, id)
	err = result.Error
	if err != nil {
		status = domain.StatusInternalServerError
		return
	}
	status = domain.StatusSuccessCreate
	return
}

func (u *UserRepo) GetAllUsers(ctx context.Context) ([]*domain.User, string, error) {
	var res []*domain.User
	err := u.Conn.WithContext(ctx).Order("id DESC").Find(&res).Error
	if err != nil {
		status := domain.StatusInternalServerError
		return nil, status, err
	}
	return res, domain.StatusSuccess, nil
}

func (u *UserRepo) GetUserById(ctx context.Context, id string) (*domain.User, string, error) {
	var entity domain.User
	err := u.Conn.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		status := domain.StatusInternalServerError
		return nil, status, err
	}
	return &entity, domain.StatusSuccess, nil
}

func (u *UserRepo) UpdateUser(ctx context.Context, id string, updates map[string]any) (*domain.User, string, error) {
	var entity domain.User

	err := u.Conn.WithContext(ctx).Model(&entity).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		return nil, domain.StatusInternalServerError, err
	}

	err = u.Conn.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, domain.StatusInternalServerError, err
	}

	return &entity, domain.StatusSuccess, nil
}
