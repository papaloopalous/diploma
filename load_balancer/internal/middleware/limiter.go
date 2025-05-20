package middleware

import (
	"net/http"
	"strconv"

	"load_balancer/internal/logger"
	"load_balancer/internal/messages"
	"load_balancer/internal/response"
	"load_balancer/internal/util"
	ratelimiter "load_balancer/rate_limiter"

	"go.uber.org/zap"
)

type MiddlewareHandler struct {
	Limiter ratelimiter.BucketIface
	Salt    string
}

// функция ограничителя кол-ва запросов
func (mh *MiddlewareHandler) LimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := util.GetClientIP(r)

		newIP := util.HashIP(ip, mh.Salt)
		tokens, err := mh.Limiter.GetTokens(newIP)

		if err != nil {
			logger.Log.Info(messages.InfoUserCreated, zap.String(messages.IP, ip), zap.Error(err))
			err = mh.Limiter.AddUser(newIP)
			if err != nil {
				response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrLimiter, nil)
				logger.Log.Error(messages.ErrLimiter, zap.Error(err))
				return
			}
			mh.Limiter.AddToken(newIP)
		}

		err = mh.Limiter.RemoveToken(newIP)
		if err != nil {
			response.WriteAPIResponse(w, http.StatusTooManyRequests, false, messages.ErrTooManyRequests, nil)
			logger.Log.Error(messages.ErrTooManyRequests, zap.Error(err))
			return
		}

		logger.Log.Info(messages.InfoAccessGranted, zap.String(messages.IP, ip), zap.String(messages.Tokens, strconv.Itoa(tokens)))
		next.ServeHTTP(w, r)
	})
}
