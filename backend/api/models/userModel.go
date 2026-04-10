package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// UserModel 用户模型
type UserModel struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`                 // 用户ID
	Name      string             `json:"name" bson:"name"`                                   // 用户名
	Email     string             `json:"email" bson:"email" validate:"required"`             // 邮箱，必填
	Password  string             `json:"password" bson:"password" validate:"required,min=5"` // 密码，必填，最少5个字符
	ImageUrl  string             `json:"imageUrl" bson:"imageUrl"`                           // 头像URL
	Bio       string             `json:"bio" bson:"bio"`                                     // 个人简介
	Followers []string           `json:"followers" bson:"followers"`                         // 关注者列表
	Following []string           `json:"following" bson:"following"`                         // 关注列表
}

// 请求体模型

// CreateUser 创建用户请求模型
type CreateUser struct {
	Email     string // 邮箱
	Password  string // 密码
	FirstName string // 名
	LastName  string // 姓
}

// LoginUser 登录用户请求模型
type LoginUser struct {
	Email    string // 邮箱
	Password string // 密码
}

// UpdateUser 更新用户请求模型
type UpdateUser struct {
	Name     string // 用户名
	ImageUrl string // 头像URL
	Bio      string // 个人简介
}
