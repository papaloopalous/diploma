package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"tarantool_api/sessionpb"
	"time"

	"github.com/joho/godotenv"
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

	expiresAt := tuple[3].(int64)
	if time.Now().Unix() > expiresAt {
		s.db.Delete("sessions", "primary", []interface{}{req.SessionId})
		return nil, errors.New("session expired")
	}

	return &sessionpb.SessionResponse{
		UserId:    tuple[1].(string),
		Role:      tuple[2].(string),
		ExpiresAt: expiresAt,
	}, nil
}

func (s *server) SetSession(ctx context.Context, req *sessionpb.SetSessionRequest) (*sessionpb.Empty, error) {
	_, err := s.db.Replace("sessions", []interface{}{
		req.SessionId,
		req.UserId,
		req.Role,
		req.ExpiresAt,
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found")
	}

	dbHost := os.Getenv("TARANTOOL_HOST")
	dbPort := os.Getenv("TARANTOOL_PORT")
	dbUser := os.Getenv("TARANTOOL_USER")
	dbPass := os.Getenv("TARANTOOL_PASS")
	serverPort := os.Getenv("SERVER_PORT")

	opts := tarantool.Opts{
		User: dbUser,
		Pass: dbPass,
	}

	db, err := tarantool.Connect(dbHost+":"+dbPort, opts)
	if err != nil {
		log.Fatalf("failed to connect to tarantool: %v", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatalf("failed to close connection: %v\n", err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	sessionpb.RegisterSessionServiceServer(s, &server{db: db})
	reflection.Register(s)

	log.Printf("server is running on port %s", serverPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
