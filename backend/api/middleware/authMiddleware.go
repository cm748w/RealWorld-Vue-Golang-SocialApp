package middleware

import (
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware JWT 认证中间件
// @Summary JWT 认证中间件
// @Description 验证请求中的 JWT 令牌，确保用户已登录
// @Tags Middleware
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {string} string "认证通过"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Router /middleware/auth [get]
func AuthMiddleware(c *fiber.Ctx) error {
	// 从请求头获取 Authorization 令牌
	tok := c.Get("Authorization")

	// 检查令牌是否存在
	if tok == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticates",
		})
	}

	// 处理令牌格式，确保去除 "Bearer " 前缀
	if !strings.HasPrefix(tok, "Bearer ") {
		tok = strings.TrimSpace(tok)
	} else {
		splited := strings.Split(tok, "Bearer ")
		if len(splited) != 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "unauthenticates",
			})
		}
		tok = splited[1]
	}

	// 获取 JWT 密钥
	SecretKey := os.Getenv("JWT_SECRET")

	// 解析 JWT 令牌
	token, err := jwt.ParseWithClaims(tok, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	// 处理解析错误
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticates",
		})
	}

	// 验证令牌有效性
	claims, ok := token.Claims.(*jwt.StandardClaims)

	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticates",
		})
	}

	// 将用户 ID 存储到上下文
	c.Locals("userId", claims.Issuer)
	// 认证通过，继续下一个中间件
	return c.Next()

}
