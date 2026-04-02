package routes

import (
	"Server/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(app *fiber.App) {
	// auth
	app.Get("/user/getUser/:id", controllers.GetUserByID)

}
