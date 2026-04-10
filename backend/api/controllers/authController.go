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

// Register 注册用户
// @Summary 注册新用户
// @Description 通过邮箱、密码、名和姓注册新用户
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.CreateUser true "用户注册信息"
// @Success 200 {object} models.UserModel
// @Failure 400 {object} map[string]interface{}
// @Router /user/signup [post]
func Register(c *fiber.Ctx) error {
	// 获取 users 集合并创建带超时的上下文，避免数据库操作长时间阻塞
	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 解析前端传入的注册参数（名、姓、邮箱、密码）
	var body models.CreateUser
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// 根据邮箱检查用户是否已存在；仅当确认为“未找到文档”时才允许继续注册
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

	// 对明文密码进行哈希处理，数据库中仅保存哈希值
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	// 组装待写入的用户文档，初始化关注/粉丝列表
	newUser := models.UserModel{
		Name:      body.FirstName + " " + body.LastName,
		Email:     body.Email,
		Password:  string(hashPassword),
		Followers: make([]string, 0),
		Following: make([]string, 0),
	}

	// 写入数据库并获取插入结果（包含 InsertedID）
	result, err := UserSchema.InsertOne(ctx, &newUser)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	// 根据插入后的 ID 回查完整用户信息，用于响应返回
	var createdUser *models.UserModel
	query := bson.M{"_id": result.InsertedID}

	if err := UserSchema.FindOne(ctx, query).Decode(&createdUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}
	// 注册成功后签发 JWT，Issuer 写入用户 ID，过期时间为 24 小时
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    createdUser.ID.Hex(),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	// 从环境变量读取签名密钥并生成 Token
	JwtSecret := os.Getenv("JWT_SECRET")

	token, _ := claims.SignedString([]byte(JwtSecret))

	// 返回新注册用户信息和访问令牌
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"result": createdUser,
		"token":  token,
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 通过邮箱和密码进行登录
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.LoginUser true "用户登录信息"
// @Success 201 {object} models.UserModel
// @Failure 400 {object} map[string]interface{}
// @Router /user/signin [post]
func Login(c *fiber.Ctx) error {
	// 获取 users 集合并创建带超时的上下文，避免数据库操作长时间阻塞
	var UserSchema = database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 解析前端传入的登录参数（邮箱、密码）
	var body models.LoginUser
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// 根据邮箱查询用户；未找到时返回统一的认证失败信息，避免泄露账号是否存在
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

	// 使用 bcrypt 对明文密码与数据库中的哈希密码进行比对
	checkPass := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if checkPass != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "invalid email or password",
		})
	}

	// 登录成功后生成 JWT，Issuer 写入用户 ID，过期时间设置为 24 小时
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.ID.Hex(),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	// 从环境变量读取签名密钥并签发 Token
	JwtSecret := os.Getenv("JWT_SECRET")

	token, _ := claims.SignedString([]byte(JwtSecret))

	// 返回登录用户信息和访问令牌
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": user,
		"token":  token,
	})
}
