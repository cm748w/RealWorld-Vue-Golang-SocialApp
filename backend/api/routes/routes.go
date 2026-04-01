package routes

import (
	"Server/controllers"
	"Server/validation"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// auth
	app.Post("/user/signup", validation.ValidateUser, controllers.Register)
}
