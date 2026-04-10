package controllers

import (
	"Server/database"
	"Server/models"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReadNotification 读取通知
// @Summary 标记通知为已读
// @Description 标记用户的通知为已读
// @Tags Notifications
// @Accept json
// @Produce json
// @Param userid path string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /notification/mark-notification-as-readed/{userid} [patch]
func ReadNotification(c *fiber.Ctx) error {

	// 读取路径参数中的用户ID
	userid := c.Params("userid")
	if userid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userid is required",
		})
	}

	// 根据用户ID构建通知过滤条件，只处理该用户的通知
	filter := bson.M{"mainUserId": bson.M{"$regex": userid, "$options": "i"}}
	// 将匹配到的通知统一标记为已读
	update := bson.M{"$set": bson.M{"isRead": true}}

	// 获取通知集合并设置超时时间，避免请求无限等待
	var NotificationSchema = database.DB.Collection("notifications")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 执行批量更新
	_, err := NotificationSchema.UpdateMany(ctx, filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to mark notification as readed",
			"error":   err.Error(),
		})
	}

	// 按创建时间倒序查询更新后的通知列表
	findOptions := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := NotificationSchema.Find(ctx, filter, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve updated notification",
			"error":   err.Error(),
		})
	}

	defer cursor.Close(ctx)

	// 将查询结果反序列化到通知数组中
	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to decode notifications",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":       "Notification marked as read",
		"notifications": notifications,
	})

}

// GetUserNotification 读取用户通知
// @Summary 读取用户通知
// @Description 读取用户的通知
// @Tags Notifications
// @Accept json
// @Produce json
// @Param userid path string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /notification/{userid} [get]
func GetUserNotification(c *fiber.Ctx) error {

	// 读取路径参数中的用户ID
	userid := c.Params("userid")
	if userid == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "userid is required",
		})
	}

	// 获取通知集合并设置查询超时时间
	var NotificationSchema = database.DB.Collection("notifications")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 根据用户ID筛选该用户的所有通知
	filter := bson.M{"mainUserId": bson.M{"$regex": userid, "$options": "i"}}
	// 查询结果按创建时间倒序排列，方便前端直接展示最新通知
	findOptions := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	// 执行查询
	cursor, err := NotificationSchema.Find(ctx, filter, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve updated notification",
			"error":   err.Error(),
		})
	}

	defer cursor.Close(ctx)

	// 解析查询结果
	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to decode notifications",
			"error":   err.Error(),
		})
	}

	if len(notifications) == 0 {
		// 如果没有通知，返回空数组，避免前端处理空值分支
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"notifications": []models.Notification{},
		})
	}

	// 返回查询到的通知列表
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":       "Notifications retrieved successfully",
		"notifications": notifications,
	})

}
