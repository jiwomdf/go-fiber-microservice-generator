package domain

import (
	"auth-service/helper"
	"context"
)

type CreateAuthRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
}

type UpdateAuthRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Auth struct {
	ID    int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (Auth) TableName() string {
	return helper.TableNameWithSchema("auths")
}

type AuthUsecase interface {
	CreateAuth(c context.Context, req *CreateAuthRequest) (*Auth, string, error)
}

type AuthRepository interface {
	CreateAuth(c context.Context, req *Auth) (*Auth, string, error)
}
