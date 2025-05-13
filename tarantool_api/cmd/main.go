package main

import (
	"context"
	"log"
	"net"
	"tarantool_api/sessionpb"

	"github.com/tarantool/go-tarantool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	sessionpb.UnimplementedSessionServiceServer
	db *tarantool.Connection
}

func (s *server) GetSession(ctx context.Context, req *sessionpb.SessionIDRequest) (*sessionpb.SessionResponse, error) {
	resp, err := s.db.Select("sessions", "primary", 0, 1, tarantool.IterEq, []interface{}{req.SessionId})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, err
	}

	tuple := resp.Data[0].([]interface{})
	return &sessionpb.SessionResponse{
		UserId: tuple[1].(string),
		Role:   tuple[2].(string),
	}, nil
}

func (s *server) SetSession(ctx context.Context, req *sessionpb.SetSessionRequest) (*sessionpb.Empty, error) {
	_, err := s.db.Replace("sessions", []interface{}{
		req.SessionId,
		req.UserId,
		req.Role,
	})
	if err != nil {
		log.Printf("Failed to insert session: %v", err)
		return nil, err
	}
	return &sessionpb.Empty{}, nil
}

func (s *server) DeleteSession(ctx context.Context, req *sessionpb.SessionIDRequest) (*sessionpb.DeleteSessionResponse, error) {
	resp, err := s.db.Delete("sessions", "primary", []interface{}{req.SessionId})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, err
	}

	tuple := resp.Data[0].([]interface{})
	return &sessionpb.DeleteSessionResponse{
		UserId: tuple[1].(string),
	}, nil
}

func main() {
	opts := tarantool.Opts{
		User: "guest",
		Pass: "",
	}

	db, err := tarantool.Connect("localhost:3301", opts)
	if err != nil {
		log.Fatalf("failed to connect to tarantool: %v", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatalf("failed to close connection: %v\n", err)
		}
	}()

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	sessionpb.RegisterSessionServiceServer(s, &server{db: db})
	reflection.Register(s)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
