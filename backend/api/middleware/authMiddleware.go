package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTSecret returns the configured signing secret and whether it is usable.
//
// An empty secret must never be accepted: HMAC verification with an empty key
// succeeds for any token an attacker signs with that same empty key, which
// turns a missing environment variable into a full authentication bypass. The
// chat and notification services already fail closed on an empty secret; this
// helper keeps the API consistent with them.
func JWTSecret() (string, bool) {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		return "", false
	}
	return s, true
}

// AuthMiddleware rejects any request that does not carry a valid HS256 bearer
// token, and stores the token issuer (the user id) in Locals("userId").
//
// Fail-closed by design: unset secret, unexpected signing method, bad
// signature, expired token and missing issuer all produce 401.
func AuthMiddleware(c *fiber.Ctx) error {
	unauthorized := func() error {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticates",
		})
	}

	// 从请求头获取 Authorization 令牌
	tok := c.Get("Authorization")
	if tok == "" {
		return unauthorized()
	}

	// 处理令牌格式，兼容 "Bearer <token>" 与裸 token 两种写法
	if strings.HasPrefix(tok, "Bearer ") {
		splited := strings.Split(tok, "Bearer ")
		if len(splited) != 2 {
			return unauthorized()
		}
		tok = splited[1]
	}
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return unauthorized()
	}

	secret, ok := JWTSecret()
	if !ok {
		// 配置缺失属于服务端错误：直接拒绝，绝不用空密钥验签
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "authentication is not configured",
		})
	}

	token, err := jwt.ParseWithClaims(
		tok,
		&jwt.RegisteredClaims{},
		func(t *jwt.Token) (interface{}, error) {
			// 固定算法：否则攻击者可切换签名算法族来试探校验逻辑
			if _, isHMAC := t.Method.(*jwt.SigningMethodHMAC); !isHMAC {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return unauthorized()
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid || claims.Issuer == "" {
		return unauthorized()
	}

	// 将用户 ID 存储到上下文
	c.Locals("userId", claims.Issuer)
	return c.Next()
}
