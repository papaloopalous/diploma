package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"load_balancer/backend"
	"load_balancer/balancer"
	configloading "load_balancer/config_loading"
	"load_balancer/internal/handler"
	"load_balancer/internal/logger"
	"load_balancer/internal/messages"
	"load_balancer/internal/middleware"
	"load_balancer/metrics"
	ratelimiter "load_balancer/rate_limiter"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// инициализация логгера, загрузка параметров из конфига
func init() {
	logger.Init()

	err := configloading.LoadConfig()
	if err != nil {
		logger.Log.Error(messages.ErrLoadConfig, zap.Error(err))
	}
}

func main() {
	metrics.Init()
	defer logger.Log.Sync()
	serverAddr, backendAddr, interval, dbAddr, salt, defaultMaxTokens, defaultRate := configloading.SetParams()

	rl := ratelimiter.NewBucket(dbAddr, defaultMaxTokens, defaultRate)
	middlewareHandler := &middleware.MiddlewareHandler{
		Limiter: rl,
		Salt:    salt,
	}

	setupHandler := &handler.LimiterHandler{
		Limiter: rl,
	}

	lb := balancer.NewBalancer()
	for _, addr := range backendAddr {
		lb.AddBack(backend.NewBackend(addr))
	}

	// контекст для завершения работы тикеров
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	go lb.HealthCheck(ctx, ticker.C)
	go middlewareHandler.Limiter.StopAllTickers(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/set_rate", setupHandler.SetRateHandler())
	mux.Handle("/set_max", setupHandler.SetMaxHandler())
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	mux.Handle("/", middlewareHandler.LimitMiddleware(lb))

	server := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	go func() {
		logger.Log.Info(messages.InfoBalancerON, zap.String(messages.Port, server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error(messages.ErrLAS, zap.Error(err))
		}
	}()

	<-stop

	logger.Log.Info(messages.InfoGracefulStopStart)
	cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Error(messages.ErrShutdown, zap.Error(err))
	}
	logger.Log.Info(messages.InfoGracefulStopFinish)
}
