package routes

import (
	"Server/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupChatRoutes(app *fiber.App) {
	app.Post("/chat/sendmessage", controllers.SendMessage)

}
