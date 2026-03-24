package handler

import (
	"user-service/domain"
	"user-service/helper"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	usecase domain.UserUsecase
}

func NewUserHandler(iu domain.UserUsecase) *UserHandler {
	return &UserHandler{
		usecase: iu,
	}
}

func (as *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req domain.CreateUserRequest
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

	entity, status, err := as.usecase.CreateUser(c.Context(), &req)
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(
		helper.NewResponse(domain.StatusSuccessCreate, "User created successfully", nil, entity),
	)
}

func (as *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	entities, status, err := as.usecase.GetAllUsers(c.Context())
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		helper.NewResponse(domain.StatusSuccess, "Users retrieved successfully", nil, entities),
	)
}

func (as *UserHandler) GetUserById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "User ID is required", nil, nil),
		)
	}

	entity, status, err := as.usecase.GetUserById(c.Context(), id)
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		helper.NewResponse(domain.StatusSuccess, "User retrieved successfully", nil, entity),
	)
}

func (as *UserHandler) UpdateUser(c *fiber.Ctx) error {
	var req domain.UpdateUserRequest
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
			helper.NewResponse(domain.StatusBadRequest, "User ID is required", nil, nil),
		)
	}

	entity, status, err := as.usecase.UpdateUser(c.Context(), id, &req)
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		helper.NewResponse(domain.StatusSuccess, "User updated successfully", nil, entity),
	)
}

func (as *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			helper.NewResponse(domain.StatusBadRequest, "User ID is required", nil, nil),
		)
	}

	status, err := as.usecase.DeleteUser(c.Context(), id)
	if err != nil {
		return c.Status(domain.GetHttpStatusCode(status)).JSON(
			helper.NewResponse(status, err.Error(), nil, nil),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		helper.NewResponse(domain.StatusSuccess, "User deleted successfully", nil, nil),
	)
}
