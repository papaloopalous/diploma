package balancer

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"

	"load_balancer/backend"
	"load_balancer/internal/logger"
	"load_balancer/internal/messages"
	"load_balancer/internal/response"
	"load_balancer/internal/util"
	"load_balancer/metrics"
	"load_balancer/strategy"

	"go.uber.org/zap"
)

// BalancerIface - интерфейс балансировщика
type loadBalancer struct {
	mu      sync.RWMutex           // мьютекс для безопасного доступа к серверам
	servers []backend.BackendIface // список серверов
}

var _ BalancerIface = &loadBalancer{} // проверяем, что loadBalancer реализует интерфейс BalancerIface

func NewBalancer() *loadBalancer {
	return &loadBalancer{}
}

// AddBack - добавление сервера в список
func (lb *loadBalancer) AddBack(server backend.BackendIface) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.servers = append(lb.servers, server)
}

// GetBack - получение серверов
func (lb *loadBalancer) GetServers() []backend.BackendIface {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.servers
}

// выбор сервера
func (lb *loadBalancer) getNextBack() backend.BackendIface {
	return strategy.GetLeastConns(lb)
}

// ServeHTTP - обработка HTTP-запросов
func (lb *loadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.mu.RLock()
	maxRetries := len(lb.servers)
	lb.mu.RUnlock()

	for attempt := 0; attempt < maxRetries; attempt++ {
		server := lb.getNextBack()
		if server == nil {
			if attempt == 0 {
				response.WriteAPIResponse(w, http.StatusServiceUnavailable, false, messages.ErrNoBackends, nil)
			}
			logger.Log.Error(messages.ErrNoBackends)
			return
		}

		server.AddConn()
		metrics.BackendConnections.
			WithLabelValues(server.GetURL()).
			Set(float64(server.GetConns()))
		defer func(s backend.BackendIface) {
			s.RemoveConn()
			metrics.BackendConnections.
				WithLabelValues(s.GetURL()).
				Set(float64(s.GetConns()))
		}(server)

		metrics.ProxiedRequestCount.
			WithLabelValues(server.GetURL()).
			Inc()

		logger.Log.Info(messages.InfoForwardingURL,
			zap.String(messages.URL, server.GetURL()),
			zap.String(messages.InfoForwardingActive, strconv.Itoa(int(server.GetConns()))),
		)

		recorder := httptest.NewRecorder()
		server.GetProxy().ServeHTTP(recorder, r)

		statusCode := recorder.Code
		metrics.BackendResponseStatus.
			WithLabelValues(server.GetURL(), strconv.Itoa(statusCode)).
			Inc()

		if statusCode >= 200 && statusCode < 500 {
			util.CopyHeadersAndBody(w, recorder)
			logger.Log.Info(messages.InfoSuccessfulProxy,
				zap.String(messages.URL, server.GetURL()))
			return
		}

		metrics.ProxiedFailuresTotal.
			WithLabelValues(server.GetURL()).
			Inc()

		logger.Log.Error(messages.ErrAttemptFailed,
			zap.String(messages.Number, strconv.Itoa(attempt+1)),
			zap.String(messages.Code, strconv.Itoa(statusCode)),
			zap.String(messages.Status, http.StatusText(statusCode)),
			zap.String(messages.URL, r.URL.String()),
		)

		// помечаем backend как «плохой», если код ≥ 500
		server.SetStatus(statusCode < 500)
	}

	logger.Log.Error(messages.ErrAllAttemptsFailed)
	response.WriteAPIResponse(w, http.StatusServiceUnavailable, false, messages.ErrServiceUnavailable, nil)
}
