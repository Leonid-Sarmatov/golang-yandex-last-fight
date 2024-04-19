package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	grpc "github.com/Leonid-Sarmatov/golang-yandex-last-fight/solver_server/internal/grpc"
	rabbit "github.com/Leonid-Sarmatov/golang-yandex-last-fight/solver_server/internal/rabbit"
)

type Kostul struct {}

func (p *Kostul) Ping(g string) error {
	return nil
}

func NewCostul() *Kostul {
	return &Kostul{}
}

func main() {
	time.Sleep(5 * time.Second)

	jjj := grpc.NewGRPCManager()
	s1 := rabbit.NewRabbitManager("Solver 1", jjj)
	log.Printf("Successfull start solver with key: %v", s1.Key)
	
	s2 := rabbit.NewRabbitManager("Solver 2", jjj)
	log.Printf("Successfull start solver with key: %v", s2.Key)

	/*s3 := rabbit.NewRabbitManager("Solver 3", jjj)
	log.Printf("Successfull start solver with key: %v", s3.Key)
	
	s4 := rabbit.NewRabbitManager("Solver 4", jjj)
	log.Printf("Successfull start solver with key: %v", s4.Key)

	s5 := rabbit.NewRabbitManager("Solver 5", jjj)
	log.Printf("Successfull start solver with key: %v", s5.Key)*/

	// Создаем канал с сигналом об остановки сервиса
	osSignalsChan := make(chan os.Signal, 1)
	signal.Notify(osSignalsChan, os.Interrupt)

	// Ждем сигнал об остановке (Ctrl + C в терминале)
	<-osSignalsChan
}