package balancer

import (
	"context"
	"net/http"
	"time"

	"load_balancer/backend"
)

type BalancerIface interface {
	AddBack(server backend.BackendIface)                    //добавить сервер в список доступных
	ServeHTTP(w http.ResponseWriter, r *http.Request)       //обработка запросов
	HealthCheck(ctx context.Context, tick <-chan time.Time) //проверка статуса серверов
}
