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
	// 加载 .env 环境变量文件
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	if err := database.Connect(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to social app.")
	})

	// 注册路由
	routes.SetupAuthRoutes(app)
	routes.SetupUserRoutes(app)
	routes.SetupPostRoutes(app)
	routes.SetupChatRoutes(app)
	// 提供 Swagger 文档路由
	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Listen(":5000")
}
