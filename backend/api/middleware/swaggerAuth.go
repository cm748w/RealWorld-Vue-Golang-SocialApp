package middleware

import (
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

// SwaggerAuth 校验 Swagger 文档访问令牌：
// 支持 Authorization 头（curl / 脚本）或 ?token= 查询参数（浏览器直接打开）。
// 目的：文档仍需要有效 JWT 才能查看，但浏览器导航请求也能通过 URL 携带令牌。
func SwaggerAuth(c *fiber.Ctx) error {
	tok := c.Get("Authorization")
	if strings.HasPrefix(tok, "Bearer ") {
		tok = strings.TrimPrefix(tok, "Bearer ")
	}
	if tok == "" {
		tok = c.Query("token")
	}

	if tok == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticates",
		})
	}

	SecretKey := os.Getenv("JWT_SECRET")
	token, err := jwt.ParseWithClaims(tok, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticates",
		})
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticates",
		})
	}

	c.Locals("userId", claims.Issuer)
	return c.Next()
}
