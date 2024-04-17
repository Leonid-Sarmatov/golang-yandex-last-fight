package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"sync"

	"github.com/IBM/sarama"
)

type Task struct {
	Instructions string `json:"instructions"`
	FileName     string `json:"fileName"`
	FileData     string `json:"fileData"`
}

type TimeOfOperation struct {
	Addition       int `json:"addition"`
	Subtraction    int `json:"subtraction"`
	Division       int `json:"division"`
	Multiplication int `json:"multiplication"`
}

/*type Task struct {
	Expression      string `json:"expression"`
	UserName        string `json:"user_name"`
	TimeOfOperation `json:"time_of_operation"`
}*/

type Result struct {
	Expression string `json:"expression"`
	UserName   string `json:"user_name"`
	Result     string `json:"result"`
}

type KafkaManager struct {
	Produser        sarama.AsyncProducer
	Consumer1       *sarama.ConsumerGroup
	Consumer2       *sarama.ConsumerGroup
	TaskTopicName   string
	ResultTopicName string
	MX 				*sync.Mutex
	SolverMap       map[string]int
	Counter         int
}

type Heartbeat interface {
	Ping(string) error
}

func NewKafkaManager(heartbeat Heartbeat) *KafkaManager {
	var kafkaManager KafkaManager
	kafkaManager.TaskTopicName = "topic-solver"
	kafkaManager.ResultTopicName = "topic-result"
	kafkaManager.MX = &sync.Mutex{}
	kafkaManager.SolverMap = make(map[string]int)

	// Создаем тикер на одну секунду
	ticker := time.NewTicker(1 * time.Second)

	// Запускаем сердцебиение
	go func() {
		for {
			select {
			case <-ticker.C:
				err := heartbeat.Ping("")
				if err != nil {
					log.Printf("Ping was failed: %v", err.Error())
				}
			}
		}
	}()

	// Создание настроек для kafka продюсер
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	// Создание продюсера отправляющего задачи
	producer, err := sarama.NewAsyncProducer([]string{
		"kafka:9092",
		}, config)
	if err != nil {
		log.Printf("Can not create new producer: %v", err.Error())
	}
	kafkaManager.Produser = producer

	// Создаем поток с логом, для просмотра подтверждений сообщений
	go func() {
		for {
			select {
			case <-producer.Successes():
				log.Println("Task sent was successful")
			}
		}
	}()

	// Создание настроек для kafka консумер
	config = sarama.NewConfig()

	// Создание консумера, принимающего задачи
	consumer1, err := sarama.NewConsumerGroup([]string{
		"kafka:9092",
		}, "solver-group", config)
	if err != nil {
		log.Printf("Can not create new conumer: %v", err.Error())
	}

	consumer2, err := sarama.NewConsumerGroup([]string{
		"kafka:9092",
		}, "solver-group", config)
	if err != nil {
		log.Printf("Can not create new conumer: %v", err.Error())
	}

	// Запуск потока приема сообщений c задачами и отправки
	// их вычистителю
	mh1 := MessageHandler {
		Name: "Solver 1",
		ResultTopicName: kafkaManager.ResultTopicName,
		Producer: &producer,
	}

	mh2 := MessageHandler {
		Name: "Solver 2",
		ResultTopicName: kafkaManager.ResultTopicName,
		Producer: &producer,
	}

	go func() {
		for {
			err := consumer1.Consume(context.Background(), []string{kafkaManager.TaskTopicName}, &mh1)
			if err != nil {
				log.Printf("Error in consumer: %v", err.Error())
			}
		}
	}()

	go func() {
		for {
			err := consumer2.Consume(context.Background(), []string{kafkaManager.TaskTopicName}, &mh2)
			if err != nil {
				log.Printf("Error in consumer: %v", err.Error())
			}
		}
	}()



	return &kafkaManager
}

func (k *KafkaManager) Close() {
	(*k.Consumer1).Close()
	(*k.Consumer2).Close()
}

type MessageHandler struct{
	Name            string
	ResultTopicName string
	Producer        *sarama.AsyncProducer
}

func (k *MessageHandler) SendResultToOrchestrator(expression, userName, result string) error {
	// Создаем JSON для отправки в брокер сообщений
	jsonMessage, err := json.Marshal(Result{
		Expression:      expression,
		UserName:        userName,
		Result: result,
	})
	if err != nil {
		return err
	}

	// Создаем сообщение с JSON
	message := &sarama.ProducerMessage{
		Topic: k.ResultTopicName,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.ByteEncoder(jsonMessage),
	}

	// Отправляем сообщение
	(*k.Producer).Input() <- message
	return nil
}

func (k *MessageHandler) Setup(sarama.ConsumerGroupSession) error { return nil }
func (k *MessageHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (k *MessageHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// Итерируемся по каналу с сообщениями
		if err := k.processMessage(message); err != nil {
			// Если обработка сообщения не успешна
			log.Printf("Error processing message: %v", err)
			continue
		}
		// Если обработка успешна, фиксируем смещение
		session.MarkMessage(message, "")
	}
	return nil
}

func (k *MessageHandler) processMessage(message *sarama.ConsumerMessage) error {
	// Декодируем тело сообщения
	var task Task
	err := json.Unmarshal(message.Value, &task)
	if err != nil {
		return err
	}

	//log.Printf("Task: %v", task)
	//log.Printf("Name: %v", k.Name)
	time.Sleep(5 * time.Second)

	//k.SendResultToOrchestrator(task.Expression, task.UserName, "1234")
	log.Printf("MessageKey %v, Name %v, Data %v", message.Key, k.Name, task.FileData)
	//log.Printf("Commit result from solver: %v", k.Name)
	return nil
}


