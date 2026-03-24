package domain

import (
	"context"
	"user-service/helper"
)

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type User struct {
	ID    int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (User) TableName() string {
	return helper.TableNameWithSchema("users")
}

type UserUsecase interface {
	CreateUser(c context.Context, req *CreateUserRequest) (*User, string, error)
	GetAllUsers(c context.Context) ([]*User, string, error)
	GetUserById(c context.Context, id string) (*User, string, error)
	UpdateUser(c context.Context, id string, req *UpdateUserRequest) (*User, string, error)
	DeleteUser(c context.Context, id string) (string, error)
}

type UserRepository interface {
	CreateUser(c context.Context, req *User) (*User, string, error)
	GetAllUsers(c context.Context) ([]*User, string, error)
	GetUserById(c context.Context, id string) (*User, string, error)
	UpdateUser(c context.Context, id string, updates map[string]any) (*User, string, error)
	DeleteUser(c context.Context, id string) (string, error)
}
