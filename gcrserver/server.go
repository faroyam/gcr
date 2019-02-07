package gcrserver

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"

	"github.com/faroyam/gcr/gcr"
)

// Server grpc server
type Server struct {
	sync.RWMutex
	clientCount int64
	m           map[uuid.UUID]gcr.ChatRoom_BroadcastServer
	L           *zap.Logger
}

func (s *Server) addClient(id uuid.UUID, stream gcr.ChatRoom_BroadcastServer) {
	s.Lock()
	defer s.Unlock()
	s.m[id] = stream
	s.clientCount++
	s.L.Info("client connected", zap.Uint32("id", id.ID()))
}

func (s *Server) delClient(id uuid.UUID) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, id)
	s.clientCount--
	s.L.Info("client disconnected", zap.Uint32("id", id.ID()))
}

func (s *Server) clientsCount() int64 {
	s.RLock()
	defer s.RUnlock()
	return s.clientCount
}

func (s *Server) sendBroadcast(author, text string) {
	s.RLock()
	defer s.RUnlock()
	for id, stream := range s.m {
		if err := stream.Send(&gcr.Message{Author: author, Text: text}); err != nil {
			s.L.Warn("broadcast error", zap.Error(err), zap.Uint32("id", id.ID()))
			continue
		}
	}
}

// Broadcast rpc implementation
func (s *Server) Broadcast(stream gcr.ChatRoom_BroadcastServer) error {
	id := uuid.New()
	s.addClient(id, stream)
	defer s.delClient(id)
	for {
		in, err := stream.Recv()
		if err != nil {
			s.L.Warn("stream read error", zap.Error(err), zap.Uint32("id", id.ID()))
			return err
		}
		go s.sendBroadcast(in.Author, in.Text)
	}
}

// RandName rpc implementation
func (s *Server) RandName(ctx context.Context, in *gcr.NameRequets) (*gcr.NameResponse, error) {
	return &gcr.NameResponse{Name: GenerateRandomName()}, nil
}

// Info rpc implementation
func (s *Server) Info(_ *gcr.InfoRequest, stream gcr.ChatRoom_InfoServer) error {
	for {
		<-time.After(3 * time.Second)
		if err := stream.Send(&gcr.InfoResponse{ClientsCount: s.clientsCount()}); err != nil {
			s.L.Warn("stream send error", zap.Error(err))
			return err
		}
	}
}

// NewServer returns new grpc server instance
func NewServer() *Server {
	logger, _ := zap.NewProduction()
	return &Server{m: make(map[uuid.UUID]gcr.ChatRoom_BroadcastServer), L: logger}
}
