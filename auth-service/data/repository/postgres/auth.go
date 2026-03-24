package postgres

import (
	"auth-service/domain"
	"context"

	"gorm.io/gorm"
)

type AuthRepo struct {
	Conn *gorm.DB
}

func NewAuthRepo(conn *gorm.DB) domain.AuthRepository {
	return &AuthRepo{Conn: conn}
}

func (u *AuthRepo) CreateAuth(ctx context.Context, req *domain.Auth) (res *domain.Auth, status string, err error) {
	result := u.Conn.WithContext(ctx).Create(req)
	err = result.Error
	if err != nil {
		status = domain.StatusInternalServerError
		return
	}
	res = req
	status = domain.StatusSuccessCreate
	return
}

func (u *AuthRepo) DeleteAuth(ctx context.Context, id string) (status string, err error) {
	result := u.Conn.WithContext(ctx).Delete(&domain.Auth{}, id)
	err = result.Error
	if err != nil {
		status = domain.StatusInternalServerError
		return
	}
	if result.RowsAffected == 0 {
		status = domain.StatusNotFound
		err = domain.ErrNotFound
		return
	}
	status = domain.StatusSuccessCreate
	return
}

func (u *AuthRepo) GetAllAuths(ctx context.Context) ([]*domain.Auth, string, error) {
	var res []*domain.Auth
	err := u.Conn.WithContext(ctx).Order("id DESC").Find(&res).Error
	if err != nil {
		status := domain.StatusInternalServerError
		return nil, status, err
	}
	return res, domain.StatusSuccess, nil
}

func (u *AuthRepo) GetAuthById(ctx context.Context, id string) (*domain.Auth, string, error) {
	var entity domain.Auth
	err := u.Conn.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.StatusNotFound, domain.ErrNotFound
		}
		status := domain.StatusInternalServerError
		return nil, status, err
	}
	return &entity, domain.StatusSuccess, nil
}

func (u *AuthRepo) GetAuthByEmail(ctx context.Context, email string) (*domain.Auth, string, error) {
	var entity domain.Auth
	err := u.Conn.WithContext(ctx).Where("email = ?", email).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.StatusNotFound, domain.ErrNotFound
		}
		status := domain.StatusInternalServerError
		return nil, status, err
	}
	return &entity, domain.StatusSuccess, nil
}

func (u *AuthRepo) UpdateAuth(ctx context.Context, id string, updates map[string]any) (*domain.Auth, string, error) {
	var entity domain.Auth

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
