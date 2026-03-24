package handler

import (
	tokenjwt "auth-service/data/repository/jtw"
	"auth-service/domain"
	"auth-service/helper"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	usecase domain.AuthUsecase
}

func NewAuthHandler(iu domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		usecase: iu,
	}
}

func (as *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.CreateAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "Cannot parse JSON", nil, nil),
		)
	}

	if err := helper.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "Validation failed", nil, nil),
		)
	}

	entity, status, err := as.usecase.Login(c.Context(), &req)
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(
		helper.NewResponse(domain.StatusSuccessCreate, "Auth created successfully", nil, entity),
	)
}

func (as *AuthHandler) CreateAuth(c *fiber.Ctx) error {
	var req domain.CreateAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "Cannot parse JSON", nil, nil),
		)
	}

	if err := helper.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "Validation failed", nil, nil),
		)
	}

	entity, status, err := as.usecase.CreateAuth(c.Context(), &req)
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(
		helper.NewResponse(domain.StatusSuccessCreate, "Auth created successfully", nil, entity),
	)
}

func (as *AuthHandler) GetAllAuths(c *fiber.Ctx) error {
	entities, status, err := as.usecase.GetAllAuths(c.Context())
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		helper.NewResponse(domain.StatusSuccess, "Auths retrieved successfully", nil, entities),
	)
}

func (as *AuthHandler) GetAuthById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "Auth ID is required", nil, nil),
		)
	}

	entity, status, err := as.usecase.GetAuthById(c.Context(), id)
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		helper.NewResponse(domain.StatusSuccess, "Auth retrieved successfully", nil, entity),
	)
}

func (as *AuthHandler) UpdateAuth(c *fiber.Ctx) error {
	var req domain.UpdateAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "Cannot parse JSON", nil, nil),
		)
	}

	if err := helper.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "Validation failed", nil, nil),
		)
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "Auth ID is required", nil, nil),
		)
	}

	entity, status, err := as.usecase.UpdateAuth(c.Context(), id, &req)
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		helper.NewResponse(domain.StatusSuccess, "Auth updated successfully", nil, entity),
	)
}

func (as *AuthHandler) DeleteAuth(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "Auth ID is required", nil, nil),
		)
	}

	status, err := as.usecase.DeleteAuth(c.Context(), id)
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		helper.NewResponse(domain.StatusSuccess, "Auth deleted successfully", nil, nil),
	)
}

func (as *AuthHandler) Verify(c *fiber.Ctx) error {
	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(
			helper.NewResponse(domain.StatusUnauthorized, "Missing authorization header", nil, nil),
		)
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return c.Status(fiber.StatusUnauthorized).JSON(
			helper.NewResponse(domain.StatusUnauthorized, "Invalid authorization header", nil, nil),
		)
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(
			helper.NewResponse(domain.StatusUnauthorized, "Invalid authorization header", nil, nil),
		)
	}

	result, status, err := tokenjwt.VerifyToken(tokenString)
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	c.Set("X-Auth-Email", result.Email)
	c.Set("X-Auth-Issuer", result.Issuer)
	c.Set("X-Auth-Token-Type", result.TokenType)

	return c.SendStatus(fiber.StatusOK)
}
