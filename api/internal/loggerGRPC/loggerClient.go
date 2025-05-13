package loggergrpc

import (
	"context"
	"log"
	"time"

	"diploma/logservice"

	"google.golang.org/grpc"
)

type LogClient struct {
	client logservice.LogServiceClient
}

var LC *LogClient

func NewLogClient(conn *grpc.ClientConn) *LogClient {
	return &LogClient{
		client: logservice.NewLogServiceClient(conn),
	}
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
