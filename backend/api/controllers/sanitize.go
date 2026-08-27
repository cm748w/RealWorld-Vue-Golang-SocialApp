package controllers

import (
	"Server/models"
	"regexp"
	"strings"
)

// 常见的 XSS 向量正则
var (
	reScriptBlock = regexp.MustCompile(`(?is)<\s*script\b[^>]*>.*?<\s*/\s*script\s*>`)
	reEventAttr   = regexp.MustCompile(`(?i)\son\w+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	reJSURI       = regexp.MustCompile(`(?i)\b(javascript|vbscript)\s*:`)
	reDataURI     = regexp.MustCompile(`(?i)\bdata\s*:\s*text\s*/\s*html`)
)

// SanitizeText 去除常见的 XSS 向量，保留普通文本内容。
// 前端以文本插值渲染，这里做服务端纵深防御，防止将来引入 v-html 或第三方消费方被攻击。
func SanitizeText(s string) string {
	s = reScriptBlock.ReplaceAllString(s, "")
	s = reEventAttr.ReplaceAllString(s, "")
	s = reJSURI.ReplaceAllString(s, "blocked:")
	s = reDataURI.ReplaceAllString(s, "blocked:")
	return strings.TrimSpace(s)
}

// SanitizeUser 返回剔除了密码哈希的用户对象（密码仅用于服务端 bcrypt 比对，绝不下发）
func SanitizeUser(u models.UserModel) models.UserModel {
	u.Password = ""
	return u
}
