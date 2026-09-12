package controllers

import (
	"strings"
	"testing"

	"Server/models"
)

// TestSanitizeTextNeutralizesProvenBypasses 覆盖的是**已在生产环境复现**的绕过手法。
// 这些用例曾经全部失效（原实现是黑名单正则），因此每一个都值得留作回归护栏。
func TestSanitizeTextNeutralizesProvenBypasses(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		mustNot []string // 输出中不得出现（大小写不敏感）
		mustHas string   // 若不为空，输出中必须保留
	}{
		{
			name:    "无空格属性分隔 - svg/onload",
			in:      `<svg/onload=alert(1)>`,
			mustNot: []string{"<svg", "onload", "alert(1)>"},
		},
		{
			name:    "无空格属性分隔 - img/onerror",
			in:      `<img/onerror=alert(1) src=x>`,
			mustNot: []string{"<img", "onerror"},
		},
		{
			name:    "带空格的事件属性（原本就能拦，防回归）",
			in:      `<img src=x onerror=alert(1)>`,
			mustNot: []string{"onerror", "<img"},
		},
		{
			name:    "实体编码协议",
			in:      `<a href="java&#115;cript:alert(1)">x</a>`,
			mustNot: []string{"javascript", "<a ", "</a>"},
		},
		{
			name:    "十六进制实体协议",
			in:      `<a href="java&#x73;cript:alert(1)">x</a>`,
			mustNot: []string{"javascript", "<a "},
		},
		{
			name:    "未闭合 script 标签",
			in:      `<script>alert(1)`,
			mustNot: []string{"<script"},
		},
		{
			name:    "闭合 script 标签",
			in:      `<script>alert(1)</script>`,
			mustNot: []string{"<script", "</script>"},
		},
		{
			name:    "嵌套伪标签",
			in:      `<scr<script>ipt>alert(1)</script>`,
			mustNot: []string{"<script", "<scr<"},
		},
		{
			name:    "iframe + data:text/html",
			in:      `<iframe src="data:text/html;base64,PHNjcmlwdD4=">`,
			mustNot: []string{"<iframe", "data:text/html"},
		},
		{
			name:    "svg + animate 事件",
			in:      `<svg><animate onbegin=alert(1) attributeName=x dur=1s>`,
			mustNot: []string{"<svg", "onbegin"},
		},
		{
			name:    "纯文本中的 javscript 协议",
			in:      `点这里 javascript:alert(1) 看看`,
			mustNot: []string{"javascript:"},
		},
		{
			name:    "普通文本必须原样保留",
			in:      `今天写了 3 个 Go 服务，还读了《RealWorld》`,
			mustHas: `今天写了 3 个 Go 服务，还读了《RealWorld》`,
		},
		{
			name:    "数学不等式不应被当成标签删掉",
			in:      `当 a < b 且 b > c 时`,
			mustHas: `a`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeText(tc.in)
			low := strings.ToLower(got)
			for _, bad := range tc.mustNot {
				if strings.Contains(low, strings.ToLower(bad)) {
					t.Errorf("SanitizeText(%q) = %q，仍包含 %q", tc.in, got, bad)
				}
			}
			if tc.mustHas != "" && !strings.Contains(got, tc.mustHas) {
				t.Errorf("SanitizeText(%q) = %q，丢失了应当保留的内容 %q", tc.in, got, tc.mustHas)
			}
		})
	}
}

func TestSanitizeImageURL(t *testing.T) {
	allow := []string{
		"https://example.com/a.png",
		"HTTPS://example.com/a.png",
		"http://example.com/a.jpg",
		"data:image/png;base64,iVBORw0KGgo=",
		"data:image/jpeg;base64,/9j/4AAQ",
		"data:image/webp;base64,UklGRg==",
	}
	block := []string{
		"javascript:alert(1)",
		"JaVaScRiPt:alert(1)",
		"java&#115;cript:alert(1)",
		"vbscript:msgbox(1)",
		"data:text/html,<script>alert(1)</script>",
		"data:image/svg+xml;base64,PHN2ZyBvbmxvYWQ9YWxlcnQoMSk+",
		"data:image/svg+xml,<svg onload=alert(1)>",
		"//evil.example.com/a.png",
		"file:///etc/passwd",
		"",
	}

	for _, in := range allow {
		if got := SanitizeImageURL(in); got != in {
			t.Errorf("SanitizeImageURL(%q) = %q，应当放行", in, got)
		}
	}
	for _, in := range block {
		if got := SanitizeImageURL(in); got != "" {
			t.Errorf("SanitizeImageURL(%q) = %q，应当清空", in, got)
		}
	}
}

func TestSanitizeUserRemovesPasswordHash(t *testing.T) {
	u := models.UserModel{
		Name:     "Dash Audit",
		Email:    "audit@example.com",
		Password: "$2a$10$abcdefghijklmnopqrstuv",
	}
	got := SanitizeUser(u)
	if got.Password != "" {
		t.Errorf("SanitizeUser 未清空密码哈希，得到 %q", got.Password)
	}
	if got.Email != u.Email || got.Name != u.Name {
		t.Errorf("SanitizeUser 改动了其它字段: %+v", got)
	}
}
