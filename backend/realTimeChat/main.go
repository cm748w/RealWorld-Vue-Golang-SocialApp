package main

import (
	"log"
	"net"
	"net/url"
	"os"
	"realTimeChat/realtime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 环境变量文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	app := fiber.New()

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
			// 生产环境显式配置的允许来源（Docker 部署经反代同源访问，正常无需配置）
			for _, o := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
				if o != "" && strings.TrimSpace(o) == origin {
					return true
				}
			}
			return false
		},
	}))

	manager := realtime.NewConnectionManager(realtime.GetUserFriends)
	// register ws route
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		id := c.Params("id")

		// 鉴权：必须携带有效 token 且 iss 与路径 id 一致，否则拒绝连接
		token := c.Query("token")
		issuer, ok := realtime.VerifyJWT(token)
		if !ok || issuer != id {
			log.Printf("WS auth rejected for user %s\n", id)
			c.WriteMessage(websocket.CloseMessage, []byte("unauthorized"))
			c.Close()
			return
		}

		if manager == nil {
			return
		}
		manager.AddConnection(id, c)
		defer func() {
			manager.RemoveConnection(id)
			c.Close()
		}()

		var msg realtime.Message
		for {
			err := c.ReadJSON(&msg)
			if err != nil {
				handleWebSocketError(err, id)
				manager.RemoveConnection(id)
				c.Close()
				break
			}

			log.Printf("Received message from %s to %s : %s", msg.Sender, msg.Receiver, msg.Content)
			manager.SendToReceiver(msg)
		}

	}))

	log.Fatal(app.Listen(":8001"))
}

func handleWebSocketError(err error, userID string) {
	log.Printf("WebSocket error for user %s : %v", userID, err)
}
