package solver

import (
	kafka "github.com/Leonid-Sarmatov/golang-yandex-last-fight/solver_server/internal/kafka"
)

type Transmitter interface {
	GetTask() (kafka.Task, error)
	SendResult() error
}

type Heartbeat interface {
	Ping() error
}

type Solver struct {
	SolverName string
	SolvingNowExpression string
	Result string
	Transmitter Transmitter
	Heartbeat Heartbeat
}

func NewSolver(transmitter Transmitter, heartbeat Heartbeat) *Solver {
	var solver Solver

	// настраиваем сердцебиение
	return &solver
}

type SolverManager struct {
	SolverMap          map[string]*Solver
	TimeOfOperationMap map[string]int
}

func NewSolverManager() *SolverManager {
	var solverManager SolverManager
	return &solverManager
}
