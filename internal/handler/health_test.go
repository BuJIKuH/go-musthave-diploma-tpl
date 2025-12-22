package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	tests := []struct {
		name           string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "#1 health OK",
			wantStatusCode: http.StatusOK,
			wantBody:       `{"status":"OK"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			Health(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

			var body struct {
				Status string `json:"status"`
			}
			err := json.NewDecoder(res.Body).Decode(&body)
			assert.NoError(t, err)
			assert.Equal(t, "OK", body.Status)
		})
	}
}
