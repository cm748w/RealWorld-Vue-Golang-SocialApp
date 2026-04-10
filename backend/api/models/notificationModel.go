package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User 通知中的用户信息结构
type User struct {
	Name   string `json:"name" bson:"name"`                         // 用户名
	Avatar string `json:"avatar,omitempty" bson:"avatar,omitempty"` // 用户头像URL
}

// Notification 通知模型
type Notification struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // 通知ID
	Details    string             `json:"details" bson:"details"`             // 通知详情内容
	MainUserID string             `json:"mainUserId" bson:"mainUserId"`       // 通知接收者ID
	TargetID   string             `json:"targetId" bson:"targetId"`           // 目标ID（如帖子ID、用户ID等）
	IsRead     bool               `json:"isRead" bson:"isRead"`               // 是否已读
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`         // 创建时间
	User       User               `json:"user" bson:"user"`                   // 操作用户信息
}
