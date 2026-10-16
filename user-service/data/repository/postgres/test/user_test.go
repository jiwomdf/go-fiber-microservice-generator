package test

import "testing"

func TestUser(t *testing.T) {
	TestUserRepository_CreateUser(t)
	TestUserRepository_GetUserById(t)
	TestUserRepository_GetAllUsers(t)
	TestUserRepository_UpdateUser(t)
	TestUserRepository_DeleteUser(t)
}
