package http

import (
	"user-service/data/delivery/http/handler"
	"user-service/data/delivery/middleware"
	"user-service/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

// RouterAPI configures the routes for the user-service service.
func RouterAPI(
	app *fiber.App,
	usecase domain.UserUsecase,
) {
	// Initialize handlers
	resourceHandler := handler.NewUserHandler(usecase)

	// Group routes
	basePath := viper.GetString("server.base_path")
	api := app.Group(basePath) // e.g., /api
	v1 := api.Group("/v1")     // e.g., /api/user-service/v1

	// 1. User Routes
	v1.Post("/user", middleware.AuthMiddleware(), resourceHandler.CreateUser)
	v1.Get("/user", middleware.AuthMiddleware(), resourceHandler.GetAllUsers)
	v1.Get("/user/:id", middleware.AuthMiddleware(), resourceHandler.GetUserById)
	v1.Patch("/user/:id", middleware.AuthMiddleware(), resourceHandler.UpdateUser)
	v1.Delete("/user/:id", middleware.AuthMiddleware(), resourceHandler.DeleteUser)
}
