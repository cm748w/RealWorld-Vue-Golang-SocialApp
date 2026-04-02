package controllers

import (
	"Server/database"
	"Server/models"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetUserBy ID
// @Summary Get User By ID
// @Description GetUser Deatils By ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 201 {object} models.UserModel
// @Failure 400 {object} map[string]interface{}
// @Router /user/getUser/{id} [get]
func GetUserByID(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var user models.UserModel

	objId, _ := primitive.ObjectIDFromHex(c.Params("id"))
	// strID := c.Params("id")
	// TODO GET and REturn user posts

	// get nuser data
	userResult := UserSchema.FindOne(ctx, bson.M{"_id": objId})

	if userResult.Err() != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"message": "User Not found",
		})
	}

	userResult.Decode(&user)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":  user,
		"posts": "posts",
	})
}
