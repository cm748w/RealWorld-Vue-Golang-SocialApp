package servergrpc

import (
	"Server/database"
	"Server/models"
	pb "Server/protos"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	pb.UnimplementedRealtimeChatServiceServer
}

func (s *Server) GetUserFollowingFollowers(ctx context.Context, req *pb.UserID) (*pb.UsersIDsListResponse, error) {
	UserSchema := database.DB.Collection("users")
	userID := req.GetUserid()

	if userID == "" {
		return nil, fmt.Errorf("use id is required")
	}

	var user models.UserModel

	uid, _ := primitive.ObjectIDFromHex(userID)

	err := UserSchema.FindOne(ctx, bson.M{"_id": uid}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	friendsMap := make(map[string]bool)
	for _, id := range user.Following {
		friendsMap[id] = true
	}
	for _, id := range user.Followers {
		friendsMap[id] = true
	}

	var friends []string
	for id := range friendsMap {
		friends = append(friends, id)
	}

	return &pb.UsersIDsListResponse{
		UserIDsLists: []*pb.UserIDsList{
			{UserIdsList: friends},
		},
	}, nil
}

func (s *Server) SendMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	if req.GetSender() == "" || req.GetReceiver() == "" {
		return nil, fmt.Errorf("sender and receiver are required")
	}

	UserSchema := database.DB.Collection("users")

	// check s & r are exists
	senderID, _ := primitive.ObjectIDFromHex(req.GetSender())
	receiverID, _ := primitive.ObjectIDFromHex(req.GetReceiver())

	var sender, receiver models.UserModel
	err := UserSchema.FindOne(ctx, bson.M{"_id": senderID}).Decode(&sender)
	if err != nil {
		return nil, fmt.Errorf("sender not found")
	}
	err = UserSchema.FindOne(ctx, bson.M{"_id": receiverID}).Decode(&receiver)
	if err != nil {
		return nil, fmt.Errorf("receiver not found")
	}

	// save msg
	message := models.Message{
		Content:  req.GetContent(),
		Sender:   req.GetSender(),
		Receiver: req.GetReceiver(),
	}

	_, err = database.DB.Collection("messages").InsertOne(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("Failed to save message to db")
	}

	// update unreaded messages (atomic upsert)
	unreadedmessagesSchema := database.DB.Collection("unreadmessages")
	filter := bson.M{"mainUserId": req.GetReceiver(), "otherUserId": req.GetSender()}
	update := bson.M{"$inc": bson.M{"numOfUnreadMessages": 1}, "$set": bson.M{"isRead": false}}
	_, err = unreadedmessagesSchema.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		// In case of duplicate-key race on upsert, retry as normal update.
		if mongo.IsDuplicateKeyError(err) {
			_, retryErr := unreadedmessagesSchema.UpdateOne(ctx, filter, update)
			if retryErr != nil {
				return nil, fmt.Errorf("Failed to update unreaded messages: %v", retryErr)
			}
		} else {
			return nil, fmt.Errorf("Failed to update unreaded messages: %v", err)
		}
	}

	// res
	return &pb.MessageResponse{
		Message: req.GetContent(),
	}, nil
}
