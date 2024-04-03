package validatetoken

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type JWTValidator interface {
	ValidateJWTToken(tokenString string) (string, error)
}

/*
ValidateJWTToken проверяет токен для авторизации
*/
func ValidateJWTToken(logger *slog.Logger, j JWTValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			// Выделяем токен из тела запроса
			authHeader := r.Header.Get("Authorization")
			authHeaderParts := strings.Split(authHeader, " ")
			if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header", http.StatusBadRequest)
				return
			}
			tokenString := authHeaderParts[1]

			// Проверяем токен
			name, err := j.ValidateJWTToken(tokenString)
			if err != nil || name == "" {
				http.Error(w, "Invalid or outdated token", http.StatusBadRequest)
				return
			}

            // Если токен верный, то передаем имя пользователя из токена в контекст
			ctx := context.WithValue(r.Context(), "user_name", name)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
