package controllers

import (
	"regexp"
	"strconv"
	"strings"

	"Server/models"
)

// 本文件是**纵深防御**，不是安全边界。
//
// 真正的防线在输出侧：前端全部使用文本插值渲染，从未使用 v-html / innerHTML
// （已对生产 bundle 实测确认）。这里的目的是：即使将来有人引入 v-html，或出现
// 以 HTML 渲染这些数据的第三方消费方，库里也不该躺着可直接执行的载荷。
//
// 历史教训：旧实现是「黑名单正则」，实测存在 3 类绕过（已在生产环境复现）：
//  1. 无空格属性分隔：<svg/onload=alert(1)>、<img/onerror=alert(1) src=x>
//  2. 实体编码协议：<a href="java&#115;cript:alert(1)">
//  3. 未闭合标签：<script>alert(1)（没有 </script>，旧正则要求闭合标签才删）
//
// 因此改为「先解码实体 → 再整体删除标签 → 最后清理残留」的顺序，而不是逐个枚举
// 危险标签。白名单式 HTML 消毒库（如 bluemonday）是更彻底的方案，但会引入新依赖；
// 当前这些字段都是纯文本语义，删除标签即可满足需求。
var (
	// 需要解码的 HTML 实体。刻意**不解码 lt/gt/quot/apos**：
	// 把 &lt; 还原成 "<" 会凭空造出标签，反而丢掉用户原本想表达的纯文本。
	reEntity = regexp.MustCompile(`(?i)&(?:#(x?[0-9a-f]{1,6})|(amp|colon|tab|newline|sol|bsol|period));?`)

	// 任意标签整体删除（含末尾未闭合片段）：这是拦住 <svg/onload=...> 的关键，
	// 因为 "<svg/onload=alert(1)>" 会被整体吃掉。
	// 要求 "<" 后紧跟字母或 "/"，这样纯文本里的数学写法（"a < b"）不会被误删。
	reTag = regexp.MustCompile(`(?s)</?[a-zA-Z][^>]*>?`)

	// 标签删除后可能残留在纯文本里的事件属性与危险协议
	reEventAttr = regexp.MustCompile(`(?i)\bon[a-z]{3,}\s*=`)
	reDangerURI = regexp.MustCompile(`(?i)\b(javascript|vbscript|data)\s*:`)
)

// decodeEntities 把数字实体与少量命名实体解码为真实字符，
// 使 "java&#115;cript:" 这类编码混淆失去作用。
func decodeEntities(s string) string {
	return reEntity.ReplaceAllStringFunc(s, func(m string) string {
		body := strings.TrimSuffix(strings.TrimPrefix(m, "&"), ";")
		if strings.HasPrefix(body, "#") {
			digits := body[1:]
			base := 10
			if len(digits) > 0 && (digits[0] == 'x' || digits[0] == 'X') {
				digits, base = digits[1:], 16
			}
			n, err := strconv.ParseInt(digits, base, 32)
			if err != nil || n < 0 || n > 0x10FFFF {
				return ""
			}
			return string(rune(n))
		}
		switch strings.ToLower(body) {
		case "amp":
			return "&"
		case "colon":
			return ":"
		case "tab":
			return "\t"
		case "newline":
			return "\n"
		case "sol":
			return "/"
		case "bsol":
			return `\`
		case "period":
			return "."
		}
		return m
	})
}

// SanitizeText 把用户输入的纯文本洗净：删除全部标签、解码实体以消除编码混淆、
// 清理残留的事件属性与危险协议。普通文本内容保持原样。
func SanitizeText(s string) string {
	s = decodeEntities(s)
	s = reTag.ReplaceAllString(s, "")
	s = reEventAttr.ReplaceAllString(s, "")
	s = reDangerURI.ReplaceAllString(s, "blocked:")
	return strings.TrimSpace(s)
}

// SanitizeUser 返回剔除了密码哈希的用户对象（密码仅用于服务端 bcrypt 比对，绝不下发）
func SanitizeUser(u models.UserModel) models.UserModel {
	u.Password = ""
	return u
}

// allowedDataImage 是允许入库的 data:image 子类型白名单。
// 刻意排除 data:image/svg+xml —— SVG 可以内嵌脚本，在 <img> 里虽然不执行，
// 但一旦被 <object>/<iframe> 打开或直接导航就会执行；头像/配图用位图足够。
var allowedDataImage = []string{
	"data:image/png",
	"data:image/jpeg",
	"data:image/jpg",
	"data:image/gif",
	"data:image/webp",
	"data:image/bmp",
}

// SanitizeImageURL 只放行 http/https 与位图类 data:image/*，其余协议一律清空。
func SanitizeImageURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	// 先解码实体：防止 "java&#115;cript:" / "data&#58;text/html" 绕过协议判断
	lower := strings.ToLower(decodeEntities(u))

	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return u
	}
	for _, prefix := range allowedDataImage {
		if strings.HasPrefix(lower, prefix) {
			return u
		}
	}
	return ""
}
