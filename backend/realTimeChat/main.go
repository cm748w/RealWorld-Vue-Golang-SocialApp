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

	hub := realtime.NewHub(realtime.GetUserFriends)
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
			if err := c.WriteMessage(websocket.CloseMessage, []byte("unauthorized")); err != nil {
				log.Printf("WS: write close message: %v", err)
			}
			if err := c.Close(); err != nil {
				log.Printf("WS: close rejected conn: %v", err)
			}
			return
		}

		if hub == nil {
			return
		}
		hc := hub.AddConnection(id, c)
		defer func() {
			hub.RemoveConnection(id, hc)
			if err := c.Close(); err != nil {
				log.Printf("WS: close conn: %v", err)
			}
		}()

		var msg realtime.Message
		for {
			err := c.ReadJSON(&msg)
			if err != nil {
				handleWebSocketError(err, id)
				hub.RemoveConnection(id, hc)
				if err := c.Close(); err != nil {
					log.Printf("WS: close conn after read error: %v", err)
				}
				break
			}

			// 发送者一律以已鉴权身份为准，绝不采信报文里的 sender：
			// 否则任何登录用户都能把消息伪造成别人发的（已在审计中发现该伪造面）。
			if msg.Sender != "" && msg.Sender != id {
				log.Printf("WS: sender spoofing attempt: authenticated=%s claimed=%s", id, msg.Sender)
			}
			msg.Sender = id

			log.Printf("Received message from %s to %s : %s", msg.Sender, msg.Receiver, msg.Content)
			hub.SendToReceiver(msg)
		}
	}))

	if err := app.Listen(":8001"); err != nil {
		log.Fatalf("failed to listen on :8001: %v", err)
	}
}

func handleWebSocketError(err error, userID string) {
	log.Printf("WebSocket error for user %s : %v", userID, err)
}
