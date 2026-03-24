package handler

import (
	"auth-service/domain"
	"auth-service/helper"

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
