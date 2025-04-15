package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"diploma/logservice"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var logger *zap.Logger

type LogServer struct {
	logservice.UnimplementedLogServiceServer
}

func (s *LogServer) WriteLog(ctx context.Context, req *logservice.LogRequest) (*logservice.LogResponse, error) {

	sugaredLogger := zap.SugaredLogger(*logger.Sugar())

	logEntry := sugaredLogger.With(
		"service", req.Service,
	)

	for k, v := range req.Metadata {
		logEntry = logEntry.With(k, v)
	}

	switch req.Level {
	case "INFO":
		logEntry.Info(req.Message)
	case "ERROR":
		logEntry.Error(req.Message)
	case "DEBUG":
		logEntry.Debug(req.Message)
	default:
		logEntry.Warn("unknown log level")
	}

	return &logservice.LogResponse{Success: true}, nil
}

func main() {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	logger, _ = cfg.Build()
	defer logger.Sync()

	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env was not found")
	}

	grpcPort := os.Getenv("LOGGER_PORT")

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("error starting a server: %v", err)
	}
	grpcServer := grpc.NewServer()
	logservice.RegisterLogServiceServer(grpcServer, &LogServer{})

	fmt.Println("LogService is running on " + grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("grpc error: %v", err)
	}
}
