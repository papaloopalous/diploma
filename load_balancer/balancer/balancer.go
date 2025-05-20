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
	"load_balancer/strategy"

	"go.uber.org/zap"
)

type loadBalancer struct {
	mu      sync.RWMutex
	servers []backend.BackendIface
}

var _ BalancerIface = &loadBalancer{}

func NewBalancer() *loadBalancer {
	return &loadBalancer{}
}

func (lb *loadBalancer) AddBack(server backend.BackendIface) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	lb.servers = append(lb.servers, server)
}

func (lb *loadBalancer) GetServers() []backend.BackendIface {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.servers
}

// выбор сервера для обработки запроса
func (lb *loadBalancer) getNextBack() backend.BackendIface {
	return strategy.GetLeastConns(lb)
}

func (lb *loadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.mu.RLock()
	var maxRetries = len(lb.servers)
	lb.mu.RUnlock()

	var attempt int
	var reused bool

	for attempt = 0; attempt < maxRetries; attempt++ {
		server := lb.getNextBack()
		if server == nil {
			if attempt == 0 {
				response.WriteAPIResponse(w, http.StatusServiceUnavailable, false, messages.ErrNoBackends, nil)
			}
			logger.Log.Error(messages.ErrNoBackends)
			return
		}

		server.AddConn()
		defer server.RemoveConn()

		logger.Log.Info(messages.InfoForwardingURL, zap.String(messages.URL, server.GetURL()), zap.String(messages.InfoForwardingActive, strconv.Itoa(int(server.GetConns()))))

		if reused {
			r = r.Clone(r.Context())
		}

		recorder := httptest.NewRecorder()
		server.GetProxy().ServeHTTP(recorder, r)
		reused = true

		if recorder.Code >= 200 && recorder.Code < 400 {
			util.CopyHeadersAndBody(w, recorder)
			logger.Log.Info(messages.InfoSuccessfulProxy, zap.String(messages.URL, server.GetURL()))
			return
		}

		logger.Log.Error(messages.ErrAttemptFailed, zap.String(messages.Number, strconv.Itoa(attempt+1)),
			zap.String(messages.Code, strconv.Itoa(recorder.Code)), zap.String(messages.Status, http.StatusText(recorder.Code)))

		server.SetStatus(recorder.Code <= 500)
	}

	logger.Log.Error(messages.ErrAllAttemptsFailed)
	response.WriteAPIResponse(w, http.StatusServiceUnavailable, false, messages.ErrServiceUnavailable, nil)
}
