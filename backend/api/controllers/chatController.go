package controllers

import (
	"Server/database"
	"Server/models"
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SendMessage 发送私信
// @Summary 发送私信
// @Description 在两个用户之间发送消息
// @Tags Chat
// @Accept json
// @Produce json
// @Param message body models.SendMessageM true "消息发送参数"
// @Success 201 {object} models.Message
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /chat/sendmessage [post]
func SendMessage(c *fiber.Ctx) error {

	var MessageSchema = database.DB.Collection("messages")
	var unreadMsgSchema = database.DB.Collection("unreadmessages")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var body models.SendMessageM
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// 校验发送者身份，防止冒充
	currentUserID, ok := c.Locals("userId").(string)
	if !ok || currentUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// 确保发送者与当前登录用户一致
	if body.Sender != currentUserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannot send messages on behalf of another user",
		})
	}

	// 从 body 构造 msg
	msg := models.Message{
		Content: body.Content,
		Sender:  body.Sender,
		Recever: body.Recever,
	}
	// 将消息写入数据库
	result, err := MessageSchema.InsertOne(ctx, &msg)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":   "Failed to save message",
			"details": err.Error(),
		})
	}

	// 更新或创建未读消息计数
	var unreadMsg models.UnreadMsg
	filter := bson.M{"mainUserId": msg.Recever, "otherUserId": msg.Sender}
	update := bson.M{"$inc": bson.M{"numOfUnreadMessages": 1}, "$set": bson.M{"isRead": false}}
	opts := options.FindOneAndUpdate().SetUpsert(true)
	err = unreadMsgSchema.FindOneAndUpdate(ctx, filter, update, opts).Decode(&unreadMsg)
	if err != nil && err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"message": "Failed to update unread message",
			"details": err.Error(),
		})
	}

	// 设置 msg 的 ID 并返回完整消息
	msg.ID = result.InsertedID.(primitive.ObjectID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Message sent successfully",
		"result":  msg,
	})
}

// GetMsgsByNums 分页获取消息
// @Summary 分页获取聊天消息
// @Description 按页获取两个用户之间的聊天记录
// @Tags Chat
// @Accept json
// @Produce json
// @Param from query int true "起始页码"
// @Param firstuid query string true "第一个用户 ID"
// @Param seconduid query string true "第二个用户 ID"
// @Success 200 {object} []models.Message
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /chat/getmsgsbynums [get]
func GetMsgsByNums(c *fiber.Ctx) error {

	var MessageSchema = database.DB.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ===== 参数校验 =====
	// 1. 校验 from 参数，防止大偏移导致的 DoS 压力
	from, err := strconv.Atoi(c.Query("from"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid value for from",
			"details": err.Error(),
		})
	}
	if from < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Page number cannot be negative",
		})
	}
	const MAX_PAGE_NUM = 10000 // 防止 DoS：最大 skip = 10000 * 2 = 20000
	if from > MAX_PAGE_NUM {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Page number too large",
			"max":   MAX_PAGE_NUM,
		})
	}

	// 2. 校验 UID，防止空字符串导致的数据泄露
	firstuid := c.Query("firstuid")
	seconduid := c.Query("seconduid")

	if firstuid == "" || seconduid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Both firstuid and seconduid are required and cannot be empty",
		})
	}

	// 3. 校验 ObjectID 格式
	_, err = primitive.ObjectIDFromHex(firstuid)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid firstuid format - must be valid MongoDB ObjectID",
		})
	}
	_, err = primitive.ObjectIDFromHex(seconduid)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid seconduid format - must be valid MongoDB ObjectID",
		})
	}

	// 4. 业务校验：不允许查询同一用户自己的会话
	if firstuid == seconduid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot query messages with the same user",
		})
	}

	// 5. 校验登录状态与访问权限
	currentUserID, ok := c.Locals("userId").(string)
	if !ok || currentUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// 仅允许用户访问自己参与的会话
	if currentUserID != firstuid && currentUserID != seconduid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to view these messages",
		})
	}

	// 组装查询条件
	senderFilter := bson.M{"sender": firstuid, "recever": seconduid}
	receverFilter := bson.M{"sender": seconduid, "recever": firstuid}
	filter := bson.M{"$or": []bson.M{senderFilter, receverFilter}}
	// 分页参数
	options := options.Find()
	options.SetSort(bson.D{{Key: "_id", Value: -1}})
	options.SetSkip(int64(from * 2))
	options.SetLimit(2)

	// 查询数据库
	cursor, err := MessageSchema.Find(ctx, filter, options)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve messages",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	// 遍历游标并组装结果数组
	var messages []models.Message
	for cursor.Next(ctx) {
		var msg models.Message
		err := cursor.Decode(&msg)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to decode messages",
				"error":   err.Error(),
			})
		}
		messages = append(messages, msg)
	}

	// 检查游标遍历错误
	if err := cursor.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database cursor error during iteration",
			"error":   err.Error(),
		})
	}

	// 无结果时返回空数组
	if len(messages) == 0 {
		messages = []models.Message{}
	} else {
		// 反转消息顺序为时间正序
		for i := 0; i < len(messages)/2; i++ {
			j := len(messages) - 1 - i
			messages[i], messages[j] = messages[j], messages[i]
		}
	}

	// 返回消息列表
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msgs":        messages,
		"count":       len(messages),
		"currentPage": from,
	})
}

// GetUserUnreadMsg 获取未读消息
// @Summary 获取未读消息
// @Description 从 Bearer Token 中识别当前用户，返回该用户的未读消息列表与未读总数
// @Tags Chat
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "messages: 未读消息列表, totalUnreadMessageCount: 未读总数"
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /chat/get-user-unreadmsg [get]
func GetUserUnreadMsg(c *fiber.Ctx) error {

	var unreadMsgSchema = database.DB.Collection("unreadmessages")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 0. 校验登录状态
	currentUserID, ok := c.Locals("userId").(string)
	if !ok || currentUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// 过滤器
	filter := bson.M{"mainUserId": currentUserID, "isRead": false}

	// 查询数据库
	cursor, err := unreadMsgSchema.Find(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve unread messages",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)
	// 遍历游标并构建返回数组
	var urms []models.UnreadMsg
	totalUnreadMessageCount := 0

	for cursor.Next(ctx) {
		var urm models.UnreadMsg
		err := cursor.Decode(&urm)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to decode unread messages",
				"error":   err.Error(),
			})
		}
		if !urm.IsRead {
			urms = append(urms, urm)
		}
		totalUnreadMessageCount += urm.NumOfUnreadMessages
	}

	if err := cursor.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database cursor error during iteration",
			"error":   err.Error(),
		})
	}

	if len(urms) == 0 {
		urms = []models.UnreadMsg{}
	}

	// 返回消息列表
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"messages":                urms,
		"totalUnreadMessageCount": totalUnreadMessageCount,
	})
}
