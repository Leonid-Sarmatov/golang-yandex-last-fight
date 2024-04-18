package grpc

import (
	"context"
	"log"

	pb "github.com/Leonid-Sarmatov/golang-yandex-last-fight/solver_server/internal/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
GRPCManager структура прослойка между GRPC,
отвечающая за создание и инициализацию клиента
*/
type GRPCManager struct {
	Connection *grpc.ClientConn
	Client     *pb.GeometryServiceClient
}

func NewGRPCManager() *GRPCManager {
	var m GRPCManager
	conn, err := grpc.Dial("orchestrator_server:5000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Can not create connection to GRPC server %v", err)
	}
	m.Connection = conn

	grpcClient := pb.NewGeometryServiceClient(conn)
	m.Client = &grpcClient

	return &m
}

func (m *GRPCManager) Ping(key, expression string) error {
	_, err := (*m.Client).PingSolver(context.Background(),
		&pb.Solver{
			SolverName:           key,
			SolvingNowExpression: expression,
		},
	)

	return err
}
