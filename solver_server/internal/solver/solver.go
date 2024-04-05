package solver

import (
	kafka "github.com/Leonid-Sarmatov/golang-yandex-last-fight/solver_server/internal/kafka"
	"log/slog"
	"time"
)

type Transmitter interface {
	GetTask() (kafka.Task, error)
	SendResult(string) error
}

type Heartbeat interface {
	Ping(string) error
}

type Solver struct {
	SolverName           string
	SolvingNowExpression string
	Result               string
	Transmitter          Transmitter
	Heartbeat            Heartbeat
}

func NewSolver(logger *slog.Logger, name string, transmitter Transmitter, heartbeat Heartbeat) *Solver {
	// Создаем объект вычислителя
	var solver Solver
	solver.Heartbeat = heartbeat
	solver.Transmitter = transmitter
	solver.SolverName = name

	// Создаем тикер на одну секунду
	ticker := time.NewTicker(1 * time.Second)

	// Запускаем сердцебиение
	go func() {
		for {
			select {
			case <-ticker.C:
				heartbeat.Ping(solver.SolverName)
			}
		}
	}()

	// Запускаем поток выполнения задач
	go func() {
		for {
			// Пробуем получить задачу
			task, err := solver.Transmitter.GetTask()
			if err != nil {
				logger.Info("Can not get task", err.Error())
				continue
			}

			// Пробуем решить ее
			res, err := solver.Calculate(task)
			if err != nil {
				logger.Info("Can not calculate task")
				continue
			}

			// Пробуем отправить результат
			err = solver.Transmitter.SendResult(res)
			if err != nil {
				logger.Info("Can not send result")
				continue
			}
		}
	}()
	return &solver
}

func (s *Solver) Calculate(kafka.Task) (string, error) {
	return "", nil
}

type SolverManager struct {
	SolverMap          map[string]*Solver
	TimeOfOperationMap map[string]int
}

func NewSolverManager() *SolverManager {
	var solverManager SolverManager
	return &solverManager
}
