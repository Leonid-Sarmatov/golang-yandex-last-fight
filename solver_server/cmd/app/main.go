package main

import (
	"log"
	"os"
	"os/signal"

	kafka "github.com/Leonid-Sarmatov/golang-yandex-last-fight/solver_server/internal/kafka"
)

type Kostul struct {}

func (p *Kostul) Ping(g string) error {
	return nil
}

func NewCostul() *Kostul {
	return &Kostul{}
}

func main() {
	jjj := NewCostul()
	s1 := kafka.NewKafkaManager(jjj)
	log.Println(s1.SolverMap)

	// Создаем канал с сигналом об остановки сервиса
	osSignalsChan := make(chan os.Signal, 1)
	signal.Notify(osSignalsChan, os.Interrupt)

	// Ждем сигнал об остановке (Ctrl + C в терминале)
	<-osSignalsChan
}