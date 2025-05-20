package balancer

import (
	"context"
	"net/http"
	"time"

	"load_balancer/backend"
	"load_balancer/internal/logger"
	"load_balancer/internal/messages"

	"go.uber.org/zap"
)

// периодическая проверка доступности серверов обработки запросов
func (lb *loadBalancer) HealthCheck(ctx context.Context, tick <-chan time.Time) {
	for {
		select {

		case <-ctx.Done():
			logger.Log.Info(messages.InfoShutdownHealth)
			return

		case <-tick:
			for _, b := range lb.servers {
				go func(backend backend.BackendIface) {
					client := http.Client{Timeout: 2 * time.Second}
					resp, err := client.Get(backend.GetURL() + "/health")

					if err != nil || resp.StatusCode != http.StatusOK {
						logger.Log.Info(messages.InfoUnreachable, zap.String(messages.URL, backend.GetURL()))
						backend.SetStatus(false)
					} else {
						logger.Log.Info(messages.InfoReachable, zap.String(messages.URL, backend.GetURL()))
						backend.SetStatus(true)
					}
				}(b)
			}
		}
	}
}
