package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		nextStatusCode int
	}{
		{
			name:           "#1 success 200",
			nextStatusCode: http.StatusOK,
		},
		{
			name:           "#2 not found 404",
			nextStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаем zap logger с observer для проверки логов
			core, logs := observer.New(zap.InfoLevel)
			logger := zap.New(core)

			// тестовый хендлер next
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.nextStatusCode)
			})

			// создаем request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			// вызываем middleware
			Logger(logger)(nextHandler).ServeHTTP(w, req)

			// проверяем, что next хендлер отработал
			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.nextStatusCode, res.StatusCode)

			// проверяем, что лог был записан
			assert.Equal(t, 1, logs.Len())
			entry := logs.All()[0]
			assert.Equal(t, "Request", entry.Message)
			status := entry.ContextMap()["status"].(int64)
			assert.Equal(t, int64(tt.nextStatusCode), status)
			assert.Equal(t, "/test", entry.ContextMap()["uri"])
			assert.Equal(t, "GET", entry.ContextMap()["method"])

		})
	}
}
