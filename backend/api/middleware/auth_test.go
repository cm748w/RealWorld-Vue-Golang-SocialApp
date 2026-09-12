package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func protectedApp() *fiber.App {
	app := fiber.New()
	app.Get("/protected", AuthMiddleware, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"userId": c.Locals("userId")})
	})
	return app
}

// statusOf 发一次请求并只取状态码，避免把 body 解析混进断言。
func statusOf(t *testing.T, app *fiber.App, authorization string) int {
	t.Helper()
	req := httptest.NewRequest("GET", "/protected", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func sign(t *testing.T, secret string, claims jwt.RegisteredClaims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

// TestAuthMiddlewareFailsClosedWhenSecretMissing 是本文件最重要的用例：
// 历史实现会在 JWT_SECRET 为空时用空密钥验签，等于把「漏配环境变量」变成
// 「任意用户可伪造」。修复后必须 fail-closed。
func TestAuthMiddlewareFailsClosedWhenSecretMissing(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	app := protectedApp()

	// 用空密钥签出来的 token 曾经可以通过校验
	empty := sign(t, "", jwt.RegisteredClaims{Issuer: "6aa4bbdcd852d3089a58990d"})
	if got := statusOf(t, app, "Bearer "+empty); got != fiber.StatusInternalServerError {
		t.Errorf("空密钥 + 空密钥签发的 token: 期望 500（配置缺失），实际 %d", got)
	}
	if got := statusOf(t, app, ""); got != fiber.StatusUnauthorized {
		t.Errorf("空密钥 + 无 token: 期望 401，实际 %d", got)
	}
}

func TestAuthMiddlewareRejectsBadTokens(t *testing.T) {
	const secret = "test-jwt-secret-key"
	t.Setenv("JWT_SECRET", secret)
	app := protectedApp()
	uid := "6aa4bbdcd852d3089a58990d"
	now := time.Now()

	cases := []struct {
		name string
		auth string
	}{
		{"无 Authorization 头", ""},
		{"只有 Bearer 前缀", "Bearer "},
		{"乱码 token", "Bearer garbage.token.here"},
		{"用其它密钥签发", "Bearer " + sign(t, "another-secret", jwt.RegisteredClaims{Issuer: uid, ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))})},
		{"已过期", "Bearer " + sign(t, secret, jwt.RegisteredClaims{Issuer: uid, ExpiresAt: jwt.NewNumericDate(now.Add(-time.Minute))})},
		{"缺少 iss", "Bearer " + sign(t, secret, jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))})},
		{"alg=none 伪造", "Bearer " + func() string {
			// 手工拼一个 alg=none 的 token：头.载荷. 空签名
			h := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{Issuer: uid, ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))})
			s, err := h.SignedString(jwt.UnsafeAllowNoneSignatureType)
			if err != nil {
				t.Fatalf("none token: %v", err)
			}
			return s
		}()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := statusOf(t, app, tc.auth); got != fiber.StatusUnauthorized {
				t.Errorf("%s: 期望 401，实际 %d", tc.name, got)
			}
		})
	}
}

func TestAuthMiddlewareAcceptsValidToken(t *testing.T) {
	const secret = "test-jwt-secret-key"
	t.Setenv("JWT_SECRET", secret)
	app := protectedApp()
	uid := "6aa4bbdcd852d3089a58990d"

	tok := sign(t, secret, jwt.RegisteredClaims{
		Issuer:    uid,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("合法 token: 期望 200，实际 %d", resp.StatusCode)
	}
	buf := make([]byte, 256)
	n, _ := resp.Body.Read(buf)
	if got := string(buf[:n]); !strings.Contains(got, uid) {
		t.Errorf("Locals(userId) 未透传，响应为 %q", got)
	}

	// 兼容裸 token（不带 Bearer 前缀）
	req2 := httptest.NewRequest("GET", "/protected", nil)
	req2.Header.Set("Authorization", tok)
	resp2, err := app.Test(req2, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != fiber.StatusOK {
		t.Errorf("裸 token: 期望 200，实际 %d", resp2.StatusCode)
	}
}

func TestIssueTokenFailsClosedWithoutSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if _, err := IssueToken("abc"); err == nil {
		t.Fatal("IssueToken 在 JWT_SECRET 为空时必须返回错误，而不是用空密钥签名")
	}
}

func TestIssueTokenRoundTrip(t *testing.T) {
	t.Setenv("JWT_SECRET", "round-trip-secret")
	tok, err := IssueToken("6aa4bbdcd852d3089a58990d")
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	app := protectedApp()
	if got := statusOf(t, app, "Bearer "+tok); got != fiber.StatusOK {
		t.Errorf("自签发的 token 应当通过校验，实际 %d", got)
	}
}
