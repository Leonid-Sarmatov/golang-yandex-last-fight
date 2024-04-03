package get_list_of_solvers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

type Solver struct {
}

/*
Request структура ответа на запрос
*/
type Response struct {
	Status        string   `json:"status"`
	Error         string   `json:"error,omitempty"`
	ListOfSolvers []Solver `json:"list_of_solvers,omitempty"`
}

func NewGetListOfSolversHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*listOfTask, err := getterListOfTask.GetListOfTask()
		if err != nil {
			// Пишем в лог ошибку декодирования
			logger.Error("Get tasks was failed", err.Error())
			// Создаем ответ с ошибкой
			render.JSON(w, r, Response{Status: "Error", Error: "Get tasks was failed"})
			return
		}*/

		logger.Info("=====================================================")

		// Выдаем список с задачами клиенту
		render.JSON(w, r, Response{
			Status:   "OK",
			ListOfSolvers:  make([]Solver, 0),
		})
	}
}