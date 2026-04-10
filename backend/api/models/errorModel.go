package models

// IError 验证错误模型
type IError struct {
	Field string // 错误字段
	Tag   string // 错误标签
}
