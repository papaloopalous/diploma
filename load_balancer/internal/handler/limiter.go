package handler

import (
	"net/http"
	"strconv"

	"load_balancer/internal/logger"
	"load_balancer/internal/messages"
	"load_balancer/internal/response"
	ratelimiter "load_balancer/rate_limiter"

	"go.uber.org/zap"
)

type LimiterHandler struct {
	Limiter ratelimiter.BucketIface
}

// обработчик установки rate
func (lh *LimiterHandler) SetRateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.URL.Query().Get("ip")
		valStr := r.URL.Query().Get("value")

		if ip == "" || valStr == "" {
			logger.Log.Info(messages.ErrNoIPORVal)
			response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoIPORVal, nil)
			return
		}

		rate, err := strconv.Atoi(valStr)
		if err != nil {
			logger.Log.Info(messages.ErrBadValue, zap.Error(err))
			response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadValue, nil)
			return
		}

		if err := lh.Limiter.SetRate(ip, rate); err != nil {
			logger.Log.Info(messages.ErrSetRate, zap.Error(err))
			response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrSetRate, nil)
			return
		}

		logger.Log.Info(messages.InfoRateUPD, zap.String(messages.IP, ip))
		response.WriteAPIResponse(w, http.StatusOK, true, messages.InfoRateUPD, nil)
	}
}

// обработчик установки max tokens
func (lh *LimiterHandler) SetMaxHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.URL.Query().Get("ip")
		valStr := r.URL.Query().Get("value")

		if ip == "" || valStr == "" {
			logger.Log.Info(messages.ErrNoIPORVal)
			response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoIPORVal, nil)
			return
		}

		max, err := strconv.Atoi(valStr)
		if err != nil {
			logger.Log.Info(messages.ErrBadValue, zap.Error(err))
			response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadValue, nil)
			return
		}

		if err := lh.Limiter.SetMaxTokens(ip, max); err != nil {
			logger.Log.Info(messages.ErrSetMax, zap.Error(err))
			response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrSetMax, nil)
			return
		}

		w.WriteHeader(http.StatusOK)
		logger.Log.Info(messages.InfoMaxUPD, zap.String(messages.IP, ip))
		response.WriteAPIResponse(w, http.StatusOK, true, messages.InfoMaxUPD, nil)
	}
}
