package routes

import (
	"Server/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupNotificationRoutes(app *fiber.App) {
	// 通知相关接口
	app.Patch("/notification/mark-notification-as-readed/:userid", controllers.ReadNotification)
	app.Get("/notification/:userid", controllers.GetUserNotification)

}
