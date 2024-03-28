package validatetoken

import (
	"net/http"
	"log/slog"
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

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}