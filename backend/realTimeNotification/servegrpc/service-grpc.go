package servegrpc

import (
	"context"
	"fmt"
	"log"
	"net"
	pb "realTimeNotification/protos"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type notificationServer struct {
	pb.UnimplementedNotificationGrpcServiceServer
	wsMu *sync.Mutex
	ws   map[string]*websocket.Conn
}

type Notification struct {
	ID         string    `json:"_id"`
	Details    string    `json:"details"`
	MainUserId string    `json:"mainUserId"`
	TargetId   string    `json:"targetId"`
	IsRead     bool      `json:"isRead"`
	CreatedAt  time.Time `json:"createdAt"`
	User       User      `json:"user"`
}

type User struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

func (s *notificationServer) SendGrpcNotification(ctx context.Context, req *pb.NotificationGrpcRequest) (*emptypb.Empty, error) {

	fmt.Printf("Sending notification to user %s : %s\n", req.MainUserId, req.Details)

	// send the nptification to websocket server
	s.wsMu.Lock()
	defer s.wsMu.Unlock()

	if conn, ok := s.ws[req.MainUserId]; ok {
		notification := Notification{
			ID:         req.XId,
			MainUserId: req.MainUserId,
			Details:    req.Details,
			TargetId:   req.TargetId,
			IsRead:     req.IsRead,
			CreatedAt:  time.Unix(req.CreatedAt.Seconds, 0),
			User: User{
				Name:   req.User.Name,
				Avatar: req.User.Avatar,
			},
		}
		err := conn.WriteJSON(notification)
		if err != nil {
			log.Printf("Error sending notification to websocket server: %v", err)
		}
	}
	return &emptypb.Empty{}, nil
}

func StartGRPCServer(ws map[string]*websocket.Conn, wsMu *sync.Mutex) error {
	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		return fmt.Errorf("Faild to listen on port 8090: %v", err)
	}

	grpcServer := grpc.NewServer()
	notificationService := &notificationServer{ws: ws, wsMu: wsMu}

	pb.RegisterNotificationGrpcServiceServer(grpcServer, notificationService)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to server gRPC server : %v", err)
		}
	}()

	return nil
}
