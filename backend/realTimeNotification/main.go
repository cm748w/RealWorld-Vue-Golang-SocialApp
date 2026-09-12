package main

import (
	"log"
	"realTimeNotification/realtime"
	"realTimeNotification/servegrpc"
)

func main() {
	// Hub 由 WS 服务与 gRPC 服务共享：WS 侧注册/注销连接，gRPC 侧按 userId 下发。
	hub := realtime.NewHub()

	// call grpc server
	if err := servegrpc.StartGRPCServer(hub); err != nil {
		log.Fatalf("failed to start grpc server : %v", err)
	}

	go realtime.StartWebSocketServer(hub)
	// block main goroutine to keep program running
	select {}
}
