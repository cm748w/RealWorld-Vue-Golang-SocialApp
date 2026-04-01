package controllers

import (
	"Server/database"
	"Server/models"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// Register
// @Summary Gegister a new user
// @Description Register an ew user by providing email, password , first name , last name
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.CreateUser true "user register details"
// @Success 201 {object} models.UserModel
// @Failure 400 {object} map[string]interface{}
// @Router /user/signup [post]
func Register(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var body models.CreateUser
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	CheckUser := UserSchema.FindOne(ctx, bson.D{{Key: "email", Value: body.Email}}).Decode(&body)

	if CheckUser == nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"message": "user with email" + body.Email + "Already Exist!",
		})
	}

	// hashing password
	hashPassword, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	newUser := models.UserModel{
		Name:      body.FirstName + " " + body.LastName,
		Email:     body.Email,
		Password:  string(hashPassword),
		Followers: make([]string, 0),
		Following: make([]string, 0),
	}

	result, err := UserSchema.InsertOne(ctx, &newUser)

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": result.InsertedID,
	})
}
