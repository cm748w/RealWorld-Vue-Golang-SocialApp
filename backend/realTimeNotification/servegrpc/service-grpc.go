package servegrpc

import (
	"context"
	"fmt"
	"log"
	"net"
	pb "realTimeNotification/protos"
	"realTimeNotification/realtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type notificationServer struct {
	pb.UnimplementedNotificationGrpcServiceServer
	hub *realtime.Hub
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

	// 交给 hub 下发：它只在该用户自己的连接锁上串行化写，并带写超时。
	// 旧实现是「持全局锁 + 无超时写」，一个卡住的客户端会阻塞所有通知。
	for _, err := range s.hub.Send(req.MainUserId, notification) {
		log.Printf("Error sending notification to websocket server: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// StartGRPCServer 监听 :8090 供 API 侧调用；收到的通知经 hub 推给对应用户。
func StartGRPCServer(hub *realtime.Hub) error {
	// #nosec G102 -- 必须监听所有网卡：API 容器要经 compose 网络按服务名访问本端口；
	// 宿主端口未对外映射（已实测外部不可达）。改为 127.0.0.1 会切断通知链路。
	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		return fmt.Errorf("failed to listen on port 8090: %v", err)
	}

	grpcServer := grpc.NewServer()
	notificationService := &notificationServer{hub: hub}

	pb.RegisterNotificationGrpcServiceServer(grpcServer, notificationService)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to server gRPC server : %v", err)
		}
	}()

	return nil
}
