package controllers

import (
	"Server/database"
	"Server/models"
	"context"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Register
// @Summary Gegister a new user
// @Description Register an ew user by providing email, password , first name , last name
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.CreateUser true "user register details"
// @Success 200 {object} models.UserModel
// @Failure 400 {object} map[string]interface{}
// @Router /user/signup [post]
func Register(c *fiber.Ctx) error {

	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var body models.CreateUser
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	var existingUser models.UserModel
	if err := UserSchema.FindOne(ctx, bson.D{{Key: "email", Value: body.Email}}).Decode(&existingUser); err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "user with email " + body.Email + " already exists",
		})
	} else if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	// hashing password
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	newUser := models.UserModel{
		Name:      body.FirstName + " " + body.LastName,
		Email:     body.Email,
		Password:  string(hashPassword),
		Followers: make([]string, 0),
		Following: make([]string, 0),
	}

	result, err := UserSchema.InsertOne(ctx, &newUser)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	// get the new user
	var createdUser *models.UserModel
	query := bson.M{"_id": result.InsertedID}

	if err := UserSchema.FindOne(ctx, query).Decode(&createdUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}
	// create the tocken
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    createdUser.ID.Hex(),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	JwtSecret := os.Getenv("JWT_SECRET")

	token, _ := claims.SignedString([]byte(JwtSecret))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"result": createdUser,
		"token":  token,
	})
}

// Login
// @Summary login a  user
// @Description Login an user by providing email, password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.LoginUser true "user Login details"
// @Success 201 {object} models.UserModel
// @Failure 400 {object} map[string]interface{}
// @Router /user/signin [post]
func Login(c *fiber.Ctx) error {
	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var body models.LoginUser
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	var user models.UserModel
	if err := UserSchema.FindOne(ctx, bson.D{{Key: "email", Value: body.Email}}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "invalid email or password",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	// check if we have the same pass or not
	checkPass := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if checkPass != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "invalid email or password",
		})
	}

	// create the tocken
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.ID.Hex(),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	JwtSecret := os.Getenv("JWT_SECRET")

	token, _ := claims.SignedString([]byte(JwtSecret))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": user,
		"token":  token,
	})
}
