package routes

import (
	"Server/controllers"
	"Server/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupNotificationRoutes(app *fiber.App) {
	// 通知相关接口（挂鉴权，防止匿名读写他人通知）
	app.Patch("/notification/mark-notification-as-readed/:userid", middleware.AuthMiddleware, controllers.ReadNotification)
	app.Get("/notification/:userid", middleware.AuthMiddleware, controllers.GetUserNotification)

}
