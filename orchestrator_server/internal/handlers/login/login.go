package login

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
)

/*
Request структура запроса на регистрацию
*/
type Request struct {
	UserName string `json:"loginUsername"`
	Password string `json:"loginPassword"`
}

/*
Request структура ответа на запрос
*/
type Response struct {
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	JWTToken string `json:"message,omitempty"`
}

type JWTmanager interface {
	CreateJWTToken(userName string) (string, error)
	ValidateJWTToken(tokenString string) (string, error)
}

type CheckAccount interface {
	СheckAccountExist(userName, password string) (bool, error)
}

/*
NewLoginHandler выполняет вход в аккаунт
1. Проверяем есть ли пользователь, если его нет отправляем сообщение об этом
2. Если пользователь зарегистрирован, то создает и возвращаем ему токен
3. Сохраняем токен для последующих сравнений
*/
func NewLoginHandler(logger *slog.Logger, j JWTmanager, checkAccount CheckAccount) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Переменная для запроса
		var request Request
		if err := render.DecodeJSON(r.Body, &request); err != nil {
			// Пишем в лог ошибку декодирования
			logger.Error("Decoding request body was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Decoding request body was failed"})
			return
		}

		const hmacSampleSecret = "super_secret_signature"
		now := time.Now()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"name": "user_name",
			"nbf":  now.Add(time.Minute).Unix(),
			"exp":  now.Add(5 * time.Minute).Unix(),
			"iat":  now.Unix(),
		})

		logger.Info(token.Raw)
	}
}
