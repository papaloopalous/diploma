package strategy

import (
	"math"

	"load_balancer/backend"
)

type ServerSlice interface {
	GetServers() []backend.BackendIface // получить список серверов
}

// выбор сервера с наименьшим количеством подключений для обработки запроса
func GetLeastConns(lb ServerSlice) backend.BackendIface {
	var selected backend.BackendIface
	minConns := math.MaxInt64

	for _, back := range lb.GetServers() {
		conns := back.GetConns()
		if back.IsAlive() && conns < int64(minConns) {
			minConns = int(conns)
			selected = back
		}
	}

	return selected
}
