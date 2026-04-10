package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client MongoDB 客户端实例
var Client *mongo.Client
// DB MongoDB 数据库实例
var DB *mongo.Database

// Connect 连接到 MongoDB 数据库
// @Summary 连接到 MongoDB 数据库
// @Description 从环境变量获取 MongoDB 连接 URI，连接到 MongoDB 数据库
// @Tags Database
// @Success 200 {string} string "连接成功"
// @Failure 500 {string} string "连接失败"
// @Router /database/connect [get]
func Connect() error {
	// 创建上下文，设置30秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 从环境变量获取 MongoDB 连接 URI
	mongoUri := os.Getenv("MONGO_URI")
	if mongoUri == "" {
		// 如果环境变量未设置，使用默认连接地址
		mongoUri = "mongodb://localhost:27017"
	}

	// 连接到 MongoDB
	var err error
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))

	// 处理连接错误
	if err != nil {
		fmt.Printf("error connecting to db: %v\n", err)
		return err
	}

	// 连接成功
	fmt.Println("Connected to MongoDB")
	// 设置数据库实例为 "social"
	DB = Client.Database("social")
	return nil
}
