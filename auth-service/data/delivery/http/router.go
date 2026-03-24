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
	v1.Post("/login", resourceHandler.Login)
	v1.Get("/verify", resourceHandler.Verify)
	v1.Post("/auth", middleware.AuthMiddleware(), resourceHandler.CreateAuth)
	v1.Get("/auth", middleware.AuthMiddleware(), resourceHandler.GetAllAuths)
	v1.Get("/auth/:id", middleware.AuthMiddleware(), resourceHandler.GetAuthById)
	v1.Patch("/auth/:id", middleware.AuthMiddleware(), resourceHandler.UpdateAuth)
	v1.Delete("/auth/:id", middleware.AuthMiddleware(), resourceHandler.DeleteAuth)
}
