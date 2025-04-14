package loggergrpc

import (
	"context"
	"log"
	"os"
	"time"

	"diploma/logservice"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LogClient struct {
	client logservice.LogServiceClient
}

var LC *LogClient

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env was not found")
	}

	grpcPort := os.Getenv("LOGGER_PORT")
	LC = NewLogClient("localhost:" + grpcPort)
}

func NewLogClient(address string) *LogClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("logger connection error: %v", err)
	}
	return &LogClient{client: logservice.NewLogServiceClient(conn)}
}

func (lc *LogClient) Log(level, service, message string, metadata map[string]string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := lc.client.WriteLog(ctx, &logservice.LogRequest{
		Service:  service,
		Level:    level,
		Message:  message,
		Metadata: metadata,
	})
	if err != nil {
		log.Printf("error sending a log (level=%s): %v", level, err)
	}
}

func (lc *LogClient) LogError(service, message string, metadata map[string]string) {
	lc.Log("ERROR", service, message, metadata)
}

func (lc *LogClient) LogInfo(service, message string, metadata map[string]string) {
	lc.Log("INFO", service, message, metadata)
}
