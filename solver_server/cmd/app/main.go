package main

import (
	kafka "github.com/Leonid-Sarmatov/golang-yandex-last-fight/solver_server/internal/kafka"
)

func main() {
	k := kafka.NewKafkaManager()
	k.Listner()

	for {
		
	}
}