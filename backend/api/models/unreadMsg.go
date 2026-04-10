package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnreadMsg 未读消息模型
type UnreadMsg struct {
	ID                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // 未读消息记录ID
	MainUserID          string             `json:"mainUserId" bson:"mainUserId"`       // 消息接收者ID
	OtherUserID         string             `json:"otherUserId" bson:"otherUserId"`     // 消息发送者ID
	NumOfUnreadMessages int                `json:"numOfUnreadMessages" bson:"numOfUnreadMessages"` // 未读消息数量
	IsRead              bool               `json:"isRead" bson:"isRead"`               // 是否已读
}
