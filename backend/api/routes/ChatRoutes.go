package routes

import (
	"Server/controllers"
	"Server/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupChatRoutes 设置聊天相关路由
func SetupChatRoutes(app *fiber.App) {
	// 发送私信
	app.Post("/chat/sendmessage", middleware.AuthMiddleware, controllers.SendMessage)
	// 分页获取聊天消息
	app.Get("/chat/getmsgsbynums", middleware.AuthMiddleware, controllers.GetMsgsByNums)
	// 获取未读消息
	app.Get("/chat/get-user-unreadmsg", middleware.AuthMiddleware, controllers.GetUserUnreadMsg)
	// 标记未读消息为已读
	app.Patch("/chat/read-msg", middleware.AuthMiddleware, controllers.ReadMsg)

}
