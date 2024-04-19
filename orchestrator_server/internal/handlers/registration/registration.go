package registration

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

/*
Request структура запроса на регистрацию
*/
type Request struct {
	UserName string `json:"registerUsername"`
	Password string `json:"registerPassword"`
}

/*
Request структура ответа на запрос
*/
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type AccountCreater interface {
	СheckAccountExist(userName string) (bool, error)
	CreateNewAccount(userName, password string) error
}

/*
NewRegistrationHandler хендлер для создания нового пользователя.
1. Проверяем имя и пароль в базе, если такой пользователь есть,
возвращаем соответствующее сообщение
2. Если такого пользователя еще нет то создаем нового пользователя
accountCreater AccountCreater
*/
func NewRegistrationHandler(logger *slog.Logger, accountCreater AccountCreater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Переменная для запроса
		var request Request
		if err := render.DecodeJSON(r.Body, &request); err != nil {
			// Пишем в лог ошибку декодирования
			logger.Info("Decoding request body was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Decoding request body was failed"})
			return
		}

		// Проверяем есть ли пользователь в системе
		ok, err := accountCreater.СheckAccountExist(request.UserName)
		if err != nil {
			// Пишем в лог ошибку поиска
			logger.Info("Searching user was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Decoding request body was failed"})
			return
		}

		if !ok {
			// Если пользователя нет в системе создаем аккаунт
			err = accountCreater.CreateNewAccount(request.UserName, request.Password)
			if err != nil {
				// Пишем в лог ошибку создания аккаунта
				logger.Info("Searching user was failed", err.Error())
				// Создаем ответ с ошибкой
				render.JSON(w, r, Response{Status: "Error", Error: "Decoding request body was failed"})
				return
			}
		} else {
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "The account with the same name already exists"})
			return
		}

		// Отправляем ответ клиенту
		render.JSON(w, r, Response{
			Status: "OK",
			Message: "Successful registration",
		})
		//logger.Info("OK")
	}
}
