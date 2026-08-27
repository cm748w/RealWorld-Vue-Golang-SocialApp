package main

import (
	"Server/database"
	"Server/middleware"
	"Server/routes"
	"Server/servergrpc"
	"log"
	"net"
	"net/url"
	"os"
	"strings"

	_ "Server/docs"
	pb "Server/protos"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
		log.Println("Warning: .env file not found, using environment variables")
	}

	// 连接到 MongoDB 数据库
	if err := database.Connect(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	// 创建 Fiber 应用实例（自定义错误处理：对外返回通用信息，不泄露内部错误细节；
	// 真实错误详情记录到服务端日志，便于排查）
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("HTTP error on %s %s: %v", c.Method(), c.OriginalURL(), err)
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"message": "internal server error",
			})
		},
	})

	// 配置 CORS 中间件：
	//  - 同源请求（Docker 部署经 nginx 反代，浏览器与 API 同源）不需要 CORS，天然可用；
	//  - 跨源仅放行 本地/内网 与 CORS_ALLOWED_ORIGINS 配置的域名，其余不响应 ACAO 头
	//    （Fiber 的 CORS 中间件只控制响应头、不会拦截请求，因此不影响任何合法流量）
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOriginsFunc: func(origin string) bool {
			u, err := url.Parse(origin)
			if err != nil {
				return false
			}
			host := u.Hostname()
			if host == "localhost" || host == "127.0.0.1" || host == "::1" {
				return true
			}
			if ip := net.ParseIP(host); ip != nil && (ip.IsPrivate() || ip.IsLoopback()) {
				return true
			}
			// 生产环境显式配置的允许来源（如 https://example.com）
			for _, o := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
				if o != "" && strings.TrimSpace(o) == origin {
					return true
				}
			}
			return false
		},
	}))

	// Setup Grpc Server
	lis, err := net.Listen("tcp", ":5001")
	if err != nil {
		log.Fatalf("faild to listen : %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRealtimeChatServiceServer(grpcServer, &servergrpc.Server{})
	reflection.Register(grpcServer)
	log.Println("gRPC Server Running on Port 5001")
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to server : %v", err)
		}
	}()
	// end of setup gRPC server

	// 设置根路径的响应
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to social app.") // 返回欢迎信息
	})
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// 注册各种路由
	routes.SetupAuthRoutes(app)         // 认证相关路由（注册、登录）
	routes.SetupUserRoutes(app)         // 用户相关路由（获取用户信息、更新资料等）
	routes.SetupPostRoutes(app)         // 帖子相关路由（创建、获取、更新、删除帖子等）
	routes.SetupChatRoutes(app)         // 聊天相关路由（发送消息、获取消息等）
	routes.SetupNotificationRoutes(app) // 通知相关路由（获取通知、标记已读等）

	// 提供 Swagger 文档路由（挂鉴权，防止匿名获取完整 API 路由图）
	app.Get("/swagger/*", middleware.AuthMiddleware, swagger.HandlerDefault)

	// 启动服务器，监听 5000 端口
	app.Listen(":5000")
}
