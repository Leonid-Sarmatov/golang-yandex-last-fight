package grpc

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"time"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/config"
	pb "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/grpc/proto"
	"google.golang.org/grpc"
)

type SolverInfo struct {
	Key        string    `json:"solverName"`
	LastPing   time.Time `json:"lastPing"`
	Expression string    `json:"solvingExpression"`
	Info       string    `json:"infoString"`
}

/*
GRPCManager структура прослойка между GRPC,
отвечающая за создание и инициализацию сервера
*/
type GRPCManager struct {
	Server *Server
}

/*
 */
type Server struct {
	SolverMap map[string]*SolverInfo
	MX        sync.Mutex
	pb.UnimplementedGeometryServiceServer
}

/*
 */
func NewServer(logger *slog.Logger, cfg *config.Config) *Server {
	var server Server
	server.SolverMap = make(map[string]*SolverInfo)

	return &server
}

func (s *Server) PingSolver(ctx context.Context, solver *pb.Solver) (*pb.Empty, error) {
	if _, ok := s.SolverMap[solver.SolverName]; !ok {
		// Если вычислителя нет в системе, то регистрируем его
		s.MX.Lock()
		s.SolverMap[solver.SolverName] = &SolverInfo{
			Key:        solver.SolverName,
			LastPing:   time.Now(),
			Expression: solver.SolvingNowExpression,
			Info:       "Solver has been successfully registered",
		}
		s.MX.Unlock()
	} else {
		// Если вычислитель уже есть в системе, то обновляем данные о нем
		s.MX.Lock()
		s.SolverMap[solver.SolverName].Expression = solver.SolvingNowExpression
		s.SolverMap[solver.SolverName].LastPing = time.Now()
		if solver.SolvingNowExpression == "" {
			s.SolverMap[solver.SolverName].Info = "Solver is free"
		} else {
			s.SolverMap[solver.SolverName].Info = "Solver is working"
		}
		s.MX.Unlock()
	}

	return &pb.Empty{}, nil
}

/*
 */
func NewGRPCManager(logger *slog.Logger, cfg *config.Config) *GRPCManager {
	var m GRPCManager
	lis, err := net.Listen("tcp", ":"+cfg.GRPSServerConfig.Port)
	if err != nil {
		logger.Info("Can not create GRPC", err)

	}

	// Создаем структуры для запуска сервера
	grpcServer := grpc.NewServer()
	myServer := NewServer(logger, cfg)
	pb.RegisterGeometryServiceServer(grpcServer, myServer)
	m.Server = myServer

	// Запускаем сервер
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Info("Error serving grpc", err)
		}
	}()

	// Запускаем горутину для проверки времени пинга вычислителей
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.Server.MX.Lock()
				for _, s := range m.Server.SolverMap {
					// Если рукопожатие нет очень долго
					if time.Now().Sub(s.LastPing) >= 5*time.Second {
						s.Info = "Solver is died"
					}
				}
				m.Server.MX.Unlock()
			}
		}
	}()
	return &m
}

/*
GetListOfSolvers возвращает список с вычислителями и их параметрами
*/
func (grpc *GRPCManager) GetListOfSolvers() ([]*SolverInfo, error) {
	arr := make([]*SolverInfo, 0)
	grpc.Server.MX.Lock()
	for _, val := range grpc.Server.SolverMap {
		arr = append(arr, val)
	}
	grpc.Server.MX.Unlock()
	return arr, nil
}

/*
GetLivingSolvers возвращает список с живыми вычислителями
*/
func (grpc *GRPCManager) GetLivingSolvers() ([]*SolverInfo, error) {
	arr := make([]*SolverInfo, 0)
	grpc.Server.MX.Lock()
	for _, val := range grpc.Server.SolverMap {
		if val.Info != "Solver is died" {
			arr = append(arr, val)
		}
	}
	grpc.Server.MX.Unlock()
	return arr, nil
}
