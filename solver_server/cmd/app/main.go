package main

import (
	"log"

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
	s1 := kafka.NewKafkaManager("Solver 1", jjj)
	s2 := kafka.NewKafkaManager("Solver 2", jjj)

	log.Printf("Init OK: %v", s1.Name)
	log.Printf("Init OK: %v", s2.Name)

	for {
		
	}
}