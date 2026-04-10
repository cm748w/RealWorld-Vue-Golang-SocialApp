package main

import (
	"Server/database"
	"Server/routes"
	"log"

	_ "Server/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
)

// @title Fiber Golang Mongo Grpc Websocket 等服务
// @version 1.0
// @description 这是基于 Golang Fiber 的 REST API Swagger 文档
// @host localhost:5000
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 使用 Bearer 认证，格式为 "Bearer 空格 Token"
func main() {
	// 加载 .env 环境变量文件，用于读取配置信息如数据库连接字符串等
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// 连接到 MongoDB 数据库
	if err := database.Connect(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	// 创建 Fiber 应用实例
	app := fiber.New()

	// 配置 CORS 中间件，允许跨域请求
	app.Use(cors.New(cors.Config{
		AllowCredentials: true, // 允许携带凭证（如 cookies）
		AllowOriginsFunc: func(origin string) bool {
			return true // 允许所有来源的请求
		},
	}))

	// 设置根路径的响应
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to social app.") // 返回欢迎信息
	})

	// 注册各种路由
	routes.SetupAuthRoutes(app)         // 认证相关路由（注册、登录）
	routes.SetupUserRoutes(app)         // 用户相关路由（获取用户信息、更新资料等）
	routes.SetupPostRoutes(app)         // 帖子相关路由（创建、获取、更新、删除帖子等）
	routes.SetupChatRoutes(app)         // 聊天相关路由（发送消息、获取消息等）
	routes.SetupNotificationRoutes(app) // 通知相关路由（获取通知、标记已读等）

	// 提供 Swagger 文档路由，用于 API 文档的访问
	app.Get("/swagger/*", swagger.HandlerDefault)

	// 启动服务器，监听 5000 端口
	app.Listen(":5000")
}
