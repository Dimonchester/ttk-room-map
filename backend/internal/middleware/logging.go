package middleware

import (
	"net/http"
	"time"

	"github.com/Dimonchester/ttk-room-map/backend/internal/logger"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start).Round(time.Millisecond)
		if rec.status >= 500 {
			logger.Error(logger.CatHTTP, "%s %s -> %d (%s) [ошибка сервера/бэкенда]", r.Method, r.URL.Path, rec.status, duration)
		} else if rec.status >= 400 {
			logger.Warn(logger.CatHTTP, "%s %s -> %d (%s) [ошибка запроса клиента]", r.Method, r.URL.Path, rec.status, duration)
		} else {
			logger.Info(logger.CatHTTP, "%s %s -> %d (%s)", r.Method, r.URL.Path, rec.status, duration)
		}
	})
}
