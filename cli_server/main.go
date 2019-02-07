//go:generate protoc -I ../gcr --go_out=plugins=grpc:../gcr ../gcr/gcr.proto

package main

import (
	"flag"
	"net"

	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/faroyam/gcr/gcr"
	"github.com/faroyam/gcr/gcrserver"
)

var (
	opts []grpc.ServerOption
	port string
	tls  bool

	server = gcrserver.NewServer()
)

func init() {
	flag.StringVar(&port, "p", "50051", "port number")
	flag.BoolVar(&tls, "t", false, "enable tls ecryption")
	flag.Parse()

	if tls {
		creds, err := credentials.NewServerTLSFromFile(".cert.pem", ".cert.key")
		if err != nil {
			server.L.Fatal("failed to construct credentials", zap.Error(err))
		}

		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		server.L.Fatal("failed to listen", zap.String("port", port), zap.Error(err))
	}

	s := grpc.NewServer(opts...)

	gcr.RegisterChatRoomServer(s, server)

	server.L.Info("starting server...", zap.String("port", port))

	if err := s.Serve(listener); err != nil {
		server.L.Fatal("failed to serve", zap.Error(err))
	}
}
