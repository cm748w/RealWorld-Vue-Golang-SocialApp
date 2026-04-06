package routes

import (
	"Server/controllers"
	"Server/middleware"
	"Server/validation"

	"github.com/gofiber/fiber/v2"
)

func SetupPostRoutes(app *fiber.App) {
	// post
	// app.Get("/post/any", func(c *fiber.Ctx) error {
	// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
	// 		"success": true,
	// 	})
	// })

	// create
	app.Post("/posts", middleware.AuthMiddleware, validation.ValidatePost, controllers.CreatePost)
	// getall
	// search
	// get one
	app.Get("/posts/:id", controllers.GetPost)
	// comment
	// like
	// delete
}
