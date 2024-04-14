package send_time_of_operation

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"

	postgres "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/postgres"
)

/*
Request структура ответа на запрос
*/
type Response struct {
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

type SetterTimeOfOperation interface {
	SetTimeOfOperation(userName string, t postgres.TimeOfOperation) error
}

func NewSendTimeOfOperationsHandler(logger *slog.Logger, sto SetterTimeOfOperation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Переменная для запроса
		var request postgres.TimeOfOperation
		if err := render.DecodeJSON(r.Body, &request); err != nil {
			// Пишем в лог ошибку декодирования
			logger.Error("Decoding request body was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Decoding request body was failed"})
			return
		}

		// Получаем имя пользователя из контекста
		userName := r.Context().Value("user_name").(string)

		// Записываем время выполнения операций
		err := sto.SetTimeOfOperation(userName, request)
		if err != nil {
			// Пишем в лог ошибку декодирования
			logger.Error("Set time of operation was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Set time of operation was failed"})
			return
		}

		// Выдаем ответ клиенту
		render.JSON(w, r, Response{
			Status:   "OK",
		})
	}
}