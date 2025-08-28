package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"

	"load_balancer/internal/logger"
	"load_balancer/internal/messages"
	"load_balancer/internal/response"

	"go.uber.org/zap"
)

type backend struct {
	url          *url.URL
	reverseProxy *httputil.ReverseProxy
	activeConns  int64
	mu           sync.RWMutex
	alive        bool
}

var _ BackendIface = &backend{}

func (back *backend) AddConn() {
	atomic.AddInt64(&back.activeConns, 1)
}

func (back *backend) RemoveConn() {
	atomic.AddInt64(&back.activeConns, -1)
}

func (back *backend) GetConns() int64 {
	back.mu.RLock()
	defer back.mu.RUnlock()
	return back.activeConns
}

func (back *backend) SetStatus(alive bool) {
	back.mu.Lock()
	defer back.mu.Unlock()
	back.alive = alive
}

func (back *backend) IsAlive() bool {
	back.mu.RLock()
	defer back.mu.RUnlock()
	return back.alive
}

func (back *backend) GetURL() string {
	back.mu.RLock()
	defer back.mu.RUnlock()
	return back.url.String()
}

func (back *backend) GetProxy() *httputil.ReverseProxy {
	back.mu.RLock()
	defer back.mu.RUnlock()
	return back.reverseProxy
}

// создать структуру сервера
func NewBackend(rawurl string) *backend {
	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		logger.Log.Error(messages.ErrInvalidBackendURL, zap.String(messages.URL, rawurl))
		return nil
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	// переопределение обработчика ошибок прокси
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Log.Error(messages.ErrProxy, zap.String(messages.URL, rawurl), zap.Error(err))
		response.WriteAPIResponse(w, http.StatusBadGateway, false, messages.ErrProxy, nil)
	}

	return &backend{
		url:          parsedURL,
		reverseProxy: proxy,
		alive:        true,
	}
}
