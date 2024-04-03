package getlistoftask

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
	Status     string          `json:"status"`
	Error      string          `json:"error,omitempty"`
	ListOfTask []postgres.Task `json:"list_of_task,omitempty"`
}

type GetterListOfTask interface {
	GetListOfTask(userName string) ([]postgres.Task, error)
}

func NewGetListOfTaskHandler(logger *slog.Logger, getterListOfTask GetterListOfTask) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем имя пользователя из контекста
		userName := r.Context().Value("user_name").(string)

		listOfTask, err := getterListOfTask.GetListOfTask(userName)
		if err != nil {
			// Пишем в лог ошибку декодирования
			logger.Error("Get tasks was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Get tasks was failed"})
			return
		}

		// Выдаем список с задачами клиенту
		render.JSON(w, r, Response{
			Status:   "OK",
			ListOfTask:  listOfTask,
		})
	}
}
