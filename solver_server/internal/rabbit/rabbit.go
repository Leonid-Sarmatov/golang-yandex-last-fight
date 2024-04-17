package rabbit

import (
	"time"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
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

type RabbitManager struct {
	Connection   *amqp.Connection
	Channel      *amqp.Channel
	TaskExchange string
	Key          string
	KeyCounter   int
	KeyMax       int
}

type Heartbeat interface {
	Ping(string) error
}

func NewRabbitManager(key string, heartbeat Heartbeat) *RabbitManager {
	var rb RabbitManager

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

	// Устанавливаем соединение с сервером RabbitMQ
	connection, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Ошибка при установке соединения с RabbitMQ: %s", err)
	}

	// Создаем канал
	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Ошибка при создании канала: %s", err)
	}

	// Объявляем exchange
	err = channel.ExchangeDeclare(
		"task-exchange", // Имя exchange
		"direct",      // Тип exchange (может быть direct, fanout, topic, headers)
		false,         // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // аргументы
	)
	if err != nil {
		log.Fatalf("Ошибка при объявлении exchange: %s", err)
	}

	// Объявляем очередь
	q, err := channel.QueueDeclare("my_queue", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Привязываем очередь к exchange
	err = channel.QueueBind(q.Name, key, "task-exchange", false, nil)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Устанавливаем prefetch_count
	err = channel.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Получаем сообщения из очереди
	msgs, err := channel.Consume(
		q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("%s", err)
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			// Здесь должна быть логика обработки сообщения

			// Подтверждаем получение сообщения
			d.Ack(false)
		}
	}()

	rb.Key = key
	rb.Connection = connection
	rb.Channel = channel
	return &rb
}

func (rb *RabbitManager) Close() {
	rb.Connection.Close()
	rb.Channel.Close()
}