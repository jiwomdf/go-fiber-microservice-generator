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

func (u *AuthRepo) CreateAuth(ctx context.Context, req *domain.Auth) (response *domain.Auth, status string, err error) {
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
