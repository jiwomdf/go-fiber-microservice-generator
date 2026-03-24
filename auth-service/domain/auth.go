package domain

import (
	"auth-service/helper"
	"context"
)

type CreateAuthRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	Expired string `json:"expired"`
}

type VerifyTokenResponse struct {
	Email     string `json:"email"`
	Subject   string `json:"subject"`
	Issuer    string `json:"issuer"`
	TokenType string `json:"token_type"`
}

type UpdateAuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Auth struct {
	Email    string `json:"email" gorm:"column:email"`
	Password string `json:"-" gorm:"column:hashed_password"`
}

func (*Auth) TableName() string {
	return helper.TableNameWithSchema("auths")
}

type AuthUsecase interface {
	Login(c context.Context, req *CreateAuthRequest) (*LoginResponse, string, error)
	GetAllAuths(c context.Context) ([]*Auth, string, error)
	CreateAuth(c context.Context, req *CreateAuthRequest) (*Auth, string, error)
	GetAuthById(c context.Context, id string) (*Auth, string, error)
	UpdateAuth(c context.Context, id string, req *UpdateAuthRequest) (*Auth, string, error)
	DeleteAuth(c context.Context, id string) (string, error)
}

type AuthRepository interface {
	GetAllAuths(c context.Context) ([]*Auth, string, error)
	CreateAuth(c context.Context, req *Auth) (*Auth, string, error)
	GetAuthById(c context.Context, id string) (*Auth, string, error)
	GetAuthByEmail(c context.Context, email string) (*Auth, string, error)
	UpdateAuth(c context.Context, id string, updates map[string]any) (*Auth, string, error)
	DeleteAuth(c context.Context, id string) (string, error)
}
