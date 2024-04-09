package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

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

type Result struct {
	Expression string `json:"expression"`
	UserName   string `json:"user_name"`
	Result     string `json:"result"`
}

type KafkaManager struct {
	Name string
	Produser        sarama.AsyncProducer
	TaskTopicName   string
	ResultTopicName string
}

type Heartbeat interface {
	Ping(string) error
}

func NewKafkaManager(name string, heartbeat Heartbeat) *KafkaManager {
	var kafkaManager KafkaManager
	kafkaManager.TaskTopicName = "topic-solver"
	kafkaManager.ResultTopicName = "topic-result"
	kafkaManager.Name = name

	// Создаем тикер на одну секунду
	ticker := time.NewTicker(1 * time.Second)

	// Запускаем сердцебиение
	go func() {
		for {
			select {
			case <-ticker.C:
				err := heartbeat.Ping(kafkaManager.Name)
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
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	// Создание консумера, принимающего задачи
	consumer, err := sarama.NewConsumerGroup([]string{
		"kafka:9092",
		}, "solver-group", config)
	if err != nil {
		log.Printf("Can not create new conumer: %v", err.Error())
	}

	// Запуск потока приема сообщений c задачами и отправки
	// их вычистителю
	go func() {
		for {
			err := consumer.Consume(context.Background(), []string{kafkaManager.TaskTopicName}, &kafkaManager)
			if err != nil {
				log.Printf("Error in consumer: %v", err.Error())
			}
		}
	}()

	return &kafkaManager
}

func (k *KafkaManager) SendResultToOrchestrator(expression, userName, result string) error {
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
	k.Produser.Input() <- message
	return nil
}

func (k *KafkaManager) Setup(sarama.ConsumerGroupSession) error { return nil }
func (k *KafkaManager) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (k *KafkaManager) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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

func (k *KafkaManager) processMessage(message *sarama.ConsumerMessage) error {
	// Декодируем тело сообщения
	var task Task
	err := json.Unmarshal(message.Value, &task)
	if err != nil {
		return err
	}

	log.Printf("Task: %v", task)
	log.Printf("Name: %v", k.Name)
	time.Sleep(3 * time.Second)

	k.SendResultToOrchestrator(task.Expression, task.UserName, "1234")
	return nil
}


