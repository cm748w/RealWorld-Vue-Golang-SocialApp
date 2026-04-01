package main

import (
	"Server/database"
	"Server/routes"

	_ "Server/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
)

// @title Fiber Golang Mongo Grpc Websocet etc..
// @version 1.0
// @description This is Swagger docs for rest api golang fiber
// @host localhost:5000
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the token
func main() {
	database.Connect()
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

	// setup routes
	routes.SetupRoutes(app)
	// Serve swager doctionation
	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Listen(":5000")
}
