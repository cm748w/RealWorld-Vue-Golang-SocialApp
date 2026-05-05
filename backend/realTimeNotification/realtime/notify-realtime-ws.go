package realtime

import (
	"fmt"
	"log"
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
			return true
		},
	}))

	app.Get("/ws/:userId", websocket.New(func(c *websocket.Conn) {
		userId := c.Params("userId")
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
