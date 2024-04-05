package kafka

import (
	"log"
	"encoding/json"

	"github.com/IBM/sarama"
)

type TimeOfOperation struct {
	Addition       int `json:"addition"`
	Subtraction    int `json:"subtraction"`
	Division       int `json:"division"`
	Multiplication int `json:"multiplication"`
}

type Task struct {
	Expression      string `json:"expression"`
	UserName        string `json:"user_name"`
	TimeOfOperation `json:"time_of_operation"`
}

type KafkaManager struct {
	PartConsumer    sarama.PartitionConsumer
	TaskTopicName   string
	ResultTopicName string
}

func NewKafkaManager() *KafkaManager {
	var kafkaManager KafkaManager
	kafkaManager.TaskTopicName = "topic-solver"
	kafkaManager.ResultTopicName = "topic-result"

	consumer, err := sarama.NewConsumer([]string{"kafka:9092"}, nil)
	if err != nil {
		log.Fatalf("[ERROR] Can not create new conumer: %v", err)
	}

	partConsumer, err := consumer.ConsumePartition(kafkaManager.TaskTopicName, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("[ERROR] Failed to consume partition: %v", err)
	}

	kafkaManager.PartConsumer = partConsumer

	return &kafkaManager
}

func (k *KafkaManager) Listner() {
	// Горутина для обработки входящих сообщений от Kafka
	go func() {
		for {
			select {
			// Чтение сообщения из Kafka
			case msg, ok := <-k.PartConsumer.Messages():
				if !ok {
					log.Println("Channel closed, exiting goroutine")
					return
				}
				//log.Printf("[OK] Messange was recived: %v", msg)

				var task Task
				err := json.Unmarshal(msg.Value, &task)
				if err != nil {
					log.Printf("[ERROR] Encoding JSON was failed: %v", err)
				}

				log.Printf("[OK] Messange was recived: %v", task)
			}
		}
	}()
}


