package healthcheck

import (
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type GrpcHealthChecker struct {
	connections map[string]*grpc.ClientConn
	mu          sync.RWMutex
	interval    time.Duration
}

func NewHealthChecker(checkInterval time.Duration) *GrpcHealthChecker {
	checker := &GrpcHealthChecker{
		connections: make(map[string]*grpc.ClientConn),
		interval:    checkInterval,
	}
	go checker.startChecking()
	return checker
}

func (c *GrpcHealthChecker) AddConnection(name string, conn *grpc.ClientConn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connections[name] = conn
}

func (c *GrpcHealthChecker) startChecking() {
	ticker := time.NewTicker(c.interval)
	for range ticker.C {
		c.checkConnections()
	}
}

func (c *GrpcHealthChecker) checkConnections() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for name, conn := range c.connections {
		state := conn.GetState()
		if state != connectivity.Ready {
			loggergrpc.LC.LogInfo(messages.ServiceHealthcheck, fmt.Sprintf(messages.StatusHealth, name, state), nil)
			conn.Connect()
		}
	}
}
