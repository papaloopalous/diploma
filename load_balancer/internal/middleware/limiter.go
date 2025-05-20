package middleware

import (
	"net/http"
	"strconv"
	"time"

	"load_balancer/internal/logger"
	"load_balancer/internal/messages"
	"load_balancer/internal/response"
	"load_balancer/internal/util"
	"load_balancer/metrics"
	ratelimiter "load_balancer/rate_limiter"

	"go.uber.org/zap"
)

type MiddlewareHandler struct {
	Limiter ratelimiter.BucketIface
	Salt    string
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// LimitMiddleware - middleware для ограничения количества запросов
func (mh *MiddlewareHandler) LimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ip := util.GetClientIP(r)
		hashedIP := util.HashIP(ip, mh.Salt)

		metrics.ActiveConnections.Inc()
		defer metrics.ActiveConnections.Dec()

		tokens, err := mh.Limiter.GetTokens(hashedIP)
		if err != nil {
			logger.Log.Info(messages.InfoUserCreated, zap.String(messages.IP, ip), zap.Error(err))
			if err := mh.Limiter.AddUser(hashedIP); err != nil {
				metrics.RequestCount.WithLabelValues(r.Method, r.URL.Path, "500").Inc()
				response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrLimiter, nil)
				return
			}
			mh.Limiter.AddToken(hashedIP)
		}

		if err := mh.Limiter.RemoveToken(hashedIP); err != nil {
			metrics.RequestCount.WithLabelValues(r.Method, r.URL.Path, "429").Inc()
			metrics.RequestDuration.WithLabelValues(r.URL.Path).Observe(time.Since(start).Seconds())
			response.WriteAPIResponse(w, http.StatusTooManyRequests, false, messages.ErrTooManyRequests, nil)
			return
		}

		logger.Log.Info(messages.InfoAccessGranted,
			zap.String(messages.IP, ip),
			zap.String(messages.Tokens, strconv.Itoa(tokens)),
		)

		// оборачиваем writer для захвата итогового статуса
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		metrics.RequestCount.
			WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rec.statusCode)).
			Inc()
		metrics.RequestDuration.
			WithLabelValues(r.URL.Path).
			Observe(time.Since(start).Seconds())
	})
}
