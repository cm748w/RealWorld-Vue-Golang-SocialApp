package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message 消息模型
type Message struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // 消息ID
	Content  string             `json:"content" bson:"content"`             // 消息内容
	Sender   string             `json:"sender" bson:"sender"`               // 发送者ID
	Receiver string             `json:"receiver" bson:"receiver"`           // 接收者ID
}

// SendMessageM 发送消息请求模型
type SendMessageM struct {
	Content  string `json:"content" bson:"content" validate:"required,min=5"` // 消息内容，必填，最少5个字符
	Sender   string `json:"sender" bson:"sender" validate:"required"`         // 发送者ID，必填
	Receiver string `json:"receiver" bson:"receiver" validate:"required"`     // 接收者ID，必填
}
