package servergrpc

import (
	"Server/models"
	"Server/protos"
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	conn   *grpc.ClientConn
	client protos.NotificationGrpcServiceClient
}

func NewClient() (*Client, error) {
	// conn, err := grpc.NewClient(":8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// DevOps docker Compose usage
	addr := os.Getenv("GOLANG_NOTIFY_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:8090"
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: protos.NewNotificationGrpcServiceClient(conn),
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) SendGrpcNotification(ctx context.Context, xId, details, mainUserId, targetId string,
	isRead bool, createdAt time.Time, userName, userAvatar string) error {
	// Prepare the request
	request := &protos.NotificationGrpcRequest{
		XId:        xId,
		Details:    details,
		MainUserId: mainUserId,
		TargetId:   targetId,
		IsRead:     isRead,
		CreatedAt:  &timestamppb.Timestamp{Seconds: createdAt.Unix()},
		User: &protos.Usergrpc{
			Name:   userName,
			Avatar: userAvatar,
		},
	}

	// call the grpc client func
	_, err := c.client.SendGrpcNotification(ctx, request)
	if err != nil {
		log.Printf("Failed To send notification : %v", err)
	}
	return err
}

func SendNotification(notification models.Notification) error {
	client, err := NewClient()
	if err != nil {
		log.Printf("Failed to create grpc client %v", err)
		return err
	}
	defer client.Close()

	ctx := context.Background()
	err = client.SendGrpcNotification(
		ctx, notification.ID.Hex(),
		notification.Details,
		notification.MainUserID,
		notification.TargetID,
		notification.IsRead,
		notification.CreatedAt,
		notification.User.Name,
		notification.User.Avatar,
	)
	if err != nil {
		log.Printf("Failed to send notification : %v", err)
		return err
	}
	return nil
}
