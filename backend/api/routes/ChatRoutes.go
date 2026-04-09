package routes

import (
	"Server/controllers"
	"Server/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupChatRoutes(app *fiber.App) {
	app.Post("/chat/sendmessage", middleware.AuthMiddleware, controllers.SendMessage)
	app.Get("/chat/getmsgsbynums", middleware.AuthMiddleware, controllers.GetMsgsByNums)
	app.Get("/chat/get-user-unreadmsg", middleware.AuthMiddleware, controllers.GetUserUnreadMsg)
}
