package realtime

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
)

type Notification struct {
	ID         string    `json:"_id"`
	Details    string    `json:"details"`
	MainUserId string    `json:"mainUserId"`
	TargetId   string    `json:"targetId"`
	IsRead     bool      `json:"isRead"`
	CreatedAt  time.Time `json:"createdAt"`
	User       User      `json:"user"`
}

type User struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// StartWebSocketServer 启动对外 WS 服务：按 userId 订阅通知推送。
// 连接注册与下发统一交给 hub，不再由本函数自己持有 map 与互斥锁。
func StartWebSocketServer(hub *Hub) {
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

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/ws/:userId", websocket.New(func(c *websocket.Conn) {
		userId := c.Params("userId")

		// 鉴权：必须携带有效 token 且 iss 与路径 userId 一致，否则拒绝连接
		token := c.Query("token")
		issuer, ok := verifyJWT(token)
		if !ok || issuer != userId {
			log.Printf("WS auth rejected for user %s\n", userId)
			if err := c.WriteMessage(websocket.CloseMessage, []byte("unauthorized")); err != nil {
				log.Printf("WS: write close message: %v", err)
			}
			if err := c.Close(); err != nil {
				log.Printf("WS: close rejected conn: %v", err)
			}
			return
		}

		fmt.Printf("User %s connected\n", userId)

		// 注册连接（同一用户可多端在线）
		hc := hub.Add(userId, c)
		log.Printf("user %s connected (connections=%d)", userId, hub.Count(userId))

		// handle disconnection
		defer func() {
			hub.Remove(userId, hc)
			log.Printf("user %s disconnected (connections=%d)", userId, hub.Count(userId))
			if err := c.Close(); err != nil {
				log.Printf("WS: close conn: %v", err)
			}
		}()

		// 读循环只用于感知对端断开；下发由 gRPC 侧经 hub.Send 完成。
		// 这里保留原有的回显行为，但走 hc.writeJSON 以串行化该连接上的写，
		// 避免与 hub.Send 并发写同一连接。
		for {
			var notificationData Notification
			if err := c.ReadJSON(&notificationData); err != nil {
				log.Printf("Error reading notification data from ws : %v ", err)
				break
			}
			if err := hc.writeJSON(notificationData); err != nil {
				log.Printf("Error echoing notification data to ws : %v", err)
				break
			}
		}
	}))

	if err := app.Listen(":8088"); err != nil {
		log.Fatalf("failed to listen on :8088: %v", err)
	}
}
