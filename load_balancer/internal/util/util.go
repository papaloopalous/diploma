package util

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
)

// вспомогательная функция для копирования заголовков и тела ответа
func CopyHeadersAndBody(w http.ResponseWriter, recorder *httptest.ResponseRecorder) {
	headers := w.Header()
	for k, vs := range recorder.Header() {
		headers[k] = vs
	}
	w.WriteHeader(recorder.Code)
	recorder.Body.WriteTo(w) //nolint:errcheck
}

// вспомогательная функция для кодирования ip пользователя
func HashIP(ip, salt string) string {
	hasher := sha256.New()
	hasher.Write([]byte(salt + ip))
	return hex.EncodeToString(hasher.Sum(nil))
}

// вспомогательная функция для получения ip пользователя
func GetClientIP(r *http.Request) string {
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ips := strings.Split(fwd, ",")
		return strings.TrimSpace(ips[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
