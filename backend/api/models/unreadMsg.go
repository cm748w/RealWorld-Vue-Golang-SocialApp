package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UnreadMsg struct {
	ID                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	MainUserID          string             `json:"mainUserId" bson:"mainUserId"`
	OtherUserID         string             `json:"otherUserId" bson:"otherUserId"`
	NumOfUnreadMessages int                `json:"numOfUnreadMessages" bson:"numOfUnreadMessages"`
	IsRead              bool               `json:"isRead" bson:"isRead"`
}
