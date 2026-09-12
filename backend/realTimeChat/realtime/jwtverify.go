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

// jwtClaims 仅解析校验所需的字段（对应 API 侧签发的 iss / exp）
type jwtClaims struct {
	Issuer    string `json:"iss"`
	ExpiresAt int64  `json:"exp"`
}

// jwtSecret 只从环境变量读取 JWT 密钥。
//
// 旧实现会在 CWD 附近探测 ../api/.env、../../.env 等文件，理由是方便本地开发；
// 但「密钥从哪来」因此不可预测，静态扫描（gosec G304）也无法确认访问边界。
// 本地开发请显式提供环境变量（compose 已注入），缺失时 VerifyJWT 会 fail-closed。
func jwtSecret() string {
	return os.Getenv("JWT_SECRET")
}

// VerifyJWT 校验 HS256 JWT，成功返回签发者（Issuer，即用户 ID）
// 未配置密钥或校验失败一律拒绝（fail closed）
func VerifyJWT(token string) (string, bool) {
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
