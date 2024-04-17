package send_task

import (
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/render"

	rabbit "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/rabbit"
	//postgres "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/postgres"
)

/*
Request структура запроса
*/
type Request struct {
	Expression string `json:"expression"`
}

/*
Request структура ответа на запрос
*/
type Response struct {
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Message  string `json:"message,omitempty"`
}

type SenderTask interface {
	SendTaskToSolver(userName, expression string, gto rabbit.GetterTimeOfOperation) error
}

type SaverTask interface {
	SaveTask(userName, expression string) error
}

/*
NewSendTaskHandler принимает задачу и отправляет ее вычислителю 
1. Сохраняем задачу в базе данных
2. Отправляем задачу в брокер сообщений
*/
func NewSendTaskHandler(logger *slog.Logger, 
						senderTask SenderTask, 
						saverTask SaverTask,
						gto rabbit.GetterTimeOfOperation) http.HandlerFunc {
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

		// Проверяем выражение на валидность
		ok := isValidExpression(request.Expression)
		if !ok {
			// Пишем в лог ошибку запроса
			logger.Error("Expression is not correct")
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Expression is not correct"})
			return
		}

		// Получаем имя пользователя из контекста
		userName := r.Context().Value("user_name").(string)

		// Сохраняем задачу в базе данных
		err := saverTask.SaveTask(userName, request.Expression)
		if err != nil {
			// Пишем в лог ошибку декодирования
			logger.Error("Save in database was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Save in database was failed"})
			return
		}

		// Отправляем задачу в вычислитель
		err = senderTask.SendTaskToSolver(userName, request.Expression, gto)
		if err != nil {
			// Пишем в лог ошибку декодирования
			logger.Error("Send to solver was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Send to solver was failed"})
			return
		}

		// Выдаем токен клиенту
		render.JSON(w, r, Response{
			Status:   "OK",
			Message:  "The task was successfully accepted and sent to the solver",
		})
	}
}

func isValidExpression(expr string) bool {
	pattern1 := regexp.MustCompile(`[\d\+\-\*/]`)
	pattern2 := regexp.MustCompile(`[\+\-\*/]`)
	arr := strings.Split(expr, "")
	for i, ch := range arr {
		if !pattern1.MatchString(ch) {
			return false
		}

		if (i != len(expr)-1) &&
			pattern2.MatchString(arr[i]) &&
			pattern2.MatchString(arr[i+1]) {
			return false
		}

		if i == 0 && (ch == "*" || ch == "/") {
			return false
		}

		if (i == len(expr)-1) && pattern2.MatchString(ch) {
			return false
		}
	}

	return true
}