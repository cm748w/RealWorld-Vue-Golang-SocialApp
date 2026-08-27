package routes

import (
	"Server/controllers"
	"Server/middleware"
	"Server/validation"

	"github.com/gofiber/fiber/v2"
)

func SetupPostRoutes(app *fiber.App) {
	// 创建帖子
	app.Post("/posts", middleware.AuthMiddleware, validation.ValidatePost, controllers.CreatePost)
	// 获取帖子列表（挂鉴权，防止匿名抓取全部内容）
	app.Get("/posts", middleware.AuthMiddleware, controllers.GetAllPosts)
	// 搜索帖子和用户
	app.Get("/posts/search", middleware.AuthMiddleware, controllers.GetPostsUsersBySearch)
	// 获取单个帖子
	app.Get("/posts/:id", middleware.AuthMiddleware, controllers.GetPost)
	// 更新帖子
	app.Patch("/posts/:id", middleware.AuthMiddleware, validation.ValidatePost, controllers.UpdatePost)
	// 评论帖子
	app.Post("/posts/:id/commentPost", middleware.AuthMiddleware, controllers.CommentPost)
	// 点赞/取消点赞
	app.Patch("/posts/:id/likePost", middleware.AuthMiddleware, controllers.LikePost)
	// 删除帖子
	app.Delete("/posts/:id", middleware.AuthMiddleware, controllers.DeletePost)
}
