package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostModel 帖子模型
type PostModel struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`     // 帖子ID
	Creator string             `json:"creator" bson:"creator"`                 // 创建者ID
	Title   string             `json:"title" bson:"title" validate:"required"` // 标题，必填
	// Message      string             `json:"message" bson:"message" validate:"required,min=5"` // 内容，必填，最少5个字符
	Message      string    `json:"message" bson:"message" validate:"required"` // 内容，必填，删掉五个字符的限制，避免前端的测试压力
	Name         string    `json:"name" bson:"name"`                           // 创建者名称
	SelectedFile string    `json:"selectedFile" bson:"selectedFile"`           // 选中的文件URL
	Likes        []string  `json:"likes" bson:"likes"`                         // 点赞列表
	Comments     []string  `json:"comments" bson:"comments"`                   // 评论列表
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`                 // 创建时间
}

// CreateOrUpdatePost 创建或更新帖子请求模型
type CreateOrUpdatePost struct {
	Title string `json:"title" bson:"title" validate:"required"` // 标题，必填
	// Message      string `json:"message" bson:"message" validate:"required,min=5"` // 内容，必填，最少5个字符
	Message      string `json:"message" bson:"message" validate:"required"` // 内容，必填，删掉5个字符的限制，减轻前端的测试压力
	SelectedFile string `json:"selectedFile" bson:"selectedFile"`           // 选中的文件URL
}

// CommentPost 评论帖子请求模型
type CommentPost struct {
	Value string `json:"value" bson:"value" validate:"required"` // 评论内容，必填
}
