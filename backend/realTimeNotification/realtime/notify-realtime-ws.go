package realtime

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"sync"
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

func StartWebSocketServer(ws map[string]*websocket.Conn, wsMu *sync.Mutex) {
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
			c.WriteMessage(websocket.CloseMessage, []byte("unauthorized"))
			c.Close()
			return
		}

		fmt.Printf("User %s connected\n", userId)

		// store the we conn
		wsMu.Lock()
		ws[userId] = c
		wsMu.Unlock()

		// handle disconnection
		defer func() {
			fmt.Printf("user %s Disconnected\n", userId)

			wsMu.Lock()
			delete(ws, userId)
			wsMu.Unlock()

			c.Close()
		}()

		// list of incoming notification from grpc server
		for {
			var notificationData Notification
			err := c.ReadJSON(&notificationData)
			if err != nil {
				log.Printf("Error reading notification data from ws : %v ", err)
				break
			}
			c.WriteJSON(notificationData)
		}
	}))

	log.Fatal(app.Listen(":8088"))

}
