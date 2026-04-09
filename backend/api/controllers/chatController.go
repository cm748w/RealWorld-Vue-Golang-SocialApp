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

// SendMessage
// @Summary Send message to friend user
// @Description Send message from one user to another
// @Tags Chat
// @Accept json
// @Produce json
// @Param message body models.SendMessageM true "user SendMessage details"
// @Success 201 {object} models.Message
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
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

	// Verify sender identity - prevent spoofing
	currentUserID, ok := c.Locals("userId").(string)
	if !ok || currentUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Ensure sender matches authenticated user
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
		"result":  msg,
	})
}

// GetMsgsByNums
// @Summary Get messages by pagination
// @Description Get messages by number between two users by pagination
// @Tags Chat
// @Accept json
// @Produce json
// @Param from query int true "Starting point page num"
// @Param firstuid query string true "first user id"
// @Param seconduid query string true "second user id"
// @Success 200 {object} []models.Message
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /chat/getmsgsbynums [get]
func GetMsgsByNums(c *fiber.Ctx) error {

	var MessageSchema = database.DB.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ===== VALIDATION =====
	// 1. Validate 'from' parameter - prevent DoS via large skip values
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
	const MAX_PAGE_NUM = 10000 // Prevent DoS: max skip = 10000 * 2 = 20000
	if from > MAX_PAGE_NUM {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Page number too large",
			"max":   MAX_PAGE_NUM,
		})
	}

	// 2. Validate UIDs - prevent data leakage from empty strings
	firstuid := c.Query("firstuid")
	seconduid := c.Query("seconduid")

	if firstuid == "" || seconduid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Both firstuid and seconduid are required and cannot be empty",
		})
	}

	// 3. Validate ObjectID format
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

	// 4. Prevent querying same user with themselves (business logic)
	if firstuid == seconduid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot query messages with the same user",
		})
	}

	// 5. Verify user authentication and authorization
	currentUserID, ok := c.Locals("userId").(string)
	if !ok || currentUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Only allow users to view their own conversations
	if currentUserID != firstuid && currentUserID != seconduid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to view these messages",
		})
	}

	// construct the filter
	senderFilter := bson.M{"sender": firstuid, "recever": seconduid}
	receverFilter := bson.M{"sender": seconduid, "recever": firstuid}
	filter := bson.M{"$or": []bson.M{senderFilter, receverFilter}}
	// pagination options
	options := options.Find()
	options.SetSort(bson.D{{Key: "_id", Value: -1}})
	options.SetSkip(int64(from * 2))
	options.SetLimit(2)

	// query the db
	cursor, err := MessageSchema.Find(ctx, filter, options)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve messages",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	// iterate over the cursor and build the res array
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

	// Check for cursor iteration errors
	if err := cursor.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database cursor error during iteration",
			"error":   err.Error(),
		})
	}

	// initialize empty messages array if no results
	if len(messages) == 0 {
		messages = []models.Message{}
	} else {
		// reverse the message array
		for i := 0; i < len(messages)/2; i++ {
			j := len(messages) - 1 - i
			messages[i], messages[j] = messages[j], messages[i]
		}
	}

	// Return the messages
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msgs":        messages,
		"count":       len(messages),
		"currentPage": from,
	})
}
