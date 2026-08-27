package realtime

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"time"
)

// jwtClaims 仅解析校验所需的字段（与 API 侧 dgrijalva/jwt-go StandardClaims 一致）
type jwtClaims struct {
	Issuer    string `json:"iss"`
	ExpiresAt int64  `json:"exp"`
}

// jwtSecret 从环境变量或 .env 文件读取 JWT 密钥（与 API 服务保持一致）
func jwtSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	// 本地开发：API 的 .env 位于 backend/api/.env，按常见 CWD 探测
	for _, p := range []string{"../api/.env", "../../.env", ".env", "../../api/.env"} {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "JWT_SECRET=") {
				v := strings.TrimPrefix(line, "JWT_SECRET=")
				return strings.Trim(v, `"'`)
			}
		}
	}
	return ""
}

// verifyJWT 校验 HS256 JWT，成功返回签发者（Issuer，即用户 ID）
// 未配置密钥或校验失败一律拒绝（fail closed）
func verifyJWT(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", false
	}
	secret := jwtSecret()
	if secret == "" {
		return "", false
	}

	// 校验 HMAC-SHA256 签名
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return "", false
	}

	// 解析 payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}
	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", false
	}
	if claims.ExpiresAt > 0 && claims.ExpiresAt < time.Now().Unix() {
		return "", false
	}
	if claims.Issuer == "" {
		return "", false
	}
	return claims.Issuer, true
}
