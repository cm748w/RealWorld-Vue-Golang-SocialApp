package routes

import (
	"Server/controllers"
	"Server/middleware"

	// "Server/validation"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(app *fiber.App) {
	// 根据 ID 获取用户信息
	app.Get("/user/getUser/:id", controllers.GetUserByID)
	// 获取推荐用户
	app.Get("/user/getSug", middleware.AuthMiddleware, controllers.GetSugUser)
	// 更新用户资料
	app.Patch("/user/update", middleware.AuthMiddleware, controllers.UpdateUser)
	// 关注/取消关注用户
	app.Patch("/user/:id/following", middleware.AuthMiddleware, controllers.FollowingUser)
	// 删除当前用户
	app.Delete("/user/delete", middleware.AuthMiddleware, controllers.DeleteUser)
}
