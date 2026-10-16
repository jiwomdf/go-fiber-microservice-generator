package test

import "testing"

func TestAuth(t *testing.T) {
	TestAuthRepository_CreateAuth(t)
	TestAuthRepository_GetAuthById(t)
	TestAuthRepository_GetAuthByEmail(t)
	TestAuthRepository_GetAllAuths(t)
	TestAuthRepository_UpdateAuth(t)
	TestAuthRepository_DeleteAuth(t)
}
