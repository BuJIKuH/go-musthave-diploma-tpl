package middleware

import (
	"context"
	"go-musthave-diploma-tpl/internal/service"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type contextKey string

const UserCtxKey contextKey = "user_id"

func AuthMiddleware(secret string, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing auth header", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if !(len(parts) == 2 && parts[0] == "Bearer") {
				http.Error(w, "invalid auth header", http.StatusUnauthorized)
				return
			}

			claims, err := service.ValidateToken(parts[1], secret)
			if err != nil {
				logger.Debug("invalid token", zap.Error(err))
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserCtxKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(UserCtxKey).(string)
	return userID, ok
}
