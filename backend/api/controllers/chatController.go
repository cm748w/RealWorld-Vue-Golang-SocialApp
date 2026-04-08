package controllers

import (
	"Server/database"
	"Server/models"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SendMessage
// @Summary Send message to friend user
// @Description Send message from one user to another
// @Tags Chat
// @Accept json
// @Produce json
// @Param message body models.SendMessageM true "user SendMessage details"
// @Success 200 {object} models.Message
// @Failure 400 {object} map[string]interface{}
// @Router /chat/sendmessage [post]
func SendMessage(c *fiber.Ctx) error {

	var MessageSchema = database.DB.Collection("messages")
	var UnReadedMsgSchema = database.DB.Collection("unReadedmessages")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var body models.SendMessageM
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// 从 body 构造 msg
	msg := models.Message{
		Content: body.Content,
		Sender:  body.Sender,
		Recever: body.Recever,
	}
	// save the message to db
	result, err := MessageSchema.InsertOne(ctx, &msg)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":   "Failed to save message",
			"details": err.Error(),
		})
	}

	// update or create the unReaded message count and is readed
	var unReadedMsg models.UnReadedMsg
	filter := bson.M{"mainUserid": msg.Recever, "otherUserid": msg.Sender}
	update := bson.M{"$inc": bson.M{"numOfUnreadedMessages": 1}, "$set": bson.M{"isReaded": false}}
	opts := options.FindOneAndUpdate().SetUpsert(true)
	err = UnReadedMsgSchema.FindOneAndUpdate(ctx, filter, update, opts).Decode(&unReadedMsg)
	if err != nil && err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"message": "Failed to update unReaded message",
			"details": err.Error(),
		})
	}

	// 设置 msg 的 ID 并返回完整消息
	msg.ID = result.InsertedID.(primitive.ObjectID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Message sent successfully",
		"result":  msg, // 老外的代码只返回了id！！！
	})
}
