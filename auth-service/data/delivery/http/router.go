package http

import (
	"auth-service/data/delivery/http/handler"
	"auth-service/data/delivery/middleware"
	"auth-service/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

// RouterAPI configures the routes for the auth-service service.
func RouterAPI(
	app *fiber.App,
	usecase domain.AuthUsecase,
) {
	// Initialize handlers
	resourceHandler := handler.NewAuthHandler(usecase)

	// Group routes
	basePath := viper.GetString("server.base_path")
	api := app.Group(basePath) // e.g., /api
	v1 := api.Group("/v1")     // e.g., /api/auth-service/v1

	// 1. Auth Routes
	v1.Post("/login", middleware.AuthMiddleware(), resourceHandler.CreateAuth)
}
