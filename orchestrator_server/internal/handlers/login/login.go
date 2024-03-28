package login

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
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

type JWTCreater interface {
	CreateJWTToken(userName string) (string, error)
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
func NewLoginHandler(logger *slog.Logger, j JWTCreater, checkAccount CheckAccount) http.HandlerFunc {
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

		// Проверяем есть пользователь в системе
		ok, err := checkAccount.СheckAccountExist(request.UserName, request.Password)
		if err != nil || !ok {
			// Пишем в лог ошибку поиска в системе
			logger.Error("Searching user was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Searching user was failed"})
			return
		}

		// Создаем токен
		token, err := j.CreateJWTToken(request.UserName)
		if err != nil {
			// Пишем в лог ошибку поиска в системе
			logger.Error("Create token was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Create token was failed"})
			return
		}

		// Выдаем токен клиенту
		render.JSON(w, r, Response{
			Status: "OK",
			JWTToken: token,
		})
	}
}
