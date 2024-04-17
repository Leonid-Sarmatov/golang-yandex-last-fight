package rabbit

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/config"
	postgres "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/postgres"
)

type RabbitManager struct {
	Connection   *amqp.Connection
	Channel      *amqp.Channel
	TaskExchange string
	KeyCounter   int
	KeyMax       int
}

type Task struct {
	Expression               string `json:"expression"`
	UserName                 string `json:"user_name"`
	postgres.TimeOfOperation `json:"time_of_operation"`
}

type Result struct {
	Expression string `json:"expression"`
	UserName   string `json:"user_name"`
	Result     string `json:"result"`
}

type GetterTimeOfOperation interface {
	GetTimeOfOperation(userName string) (*postgres.TimeOfOperation, error)
}

type SaverTaskResult interface {
	SaveTaskResult(userName, expression, result string) error
}

func NewRabbitManager(logger *slog.Logger, cfg *config.Config, sr SaverTaskResult) *RabbitManager {
	var rb RabbitManager

	// Устанавливаем соединение с сервером RabbitMQ
	connection, err := amqp.Dial("amqp://guest:guest@"+cfg.RabbitConfig.Host+":"+cfg.RabbitConfig.Port+"/")
	if err != nil {
		logger.Info("Ошибка при установке соединения с RabbitMQ", err)
	}

	// Создаем канал
	channel, err := connection.Channel()
	if err != nil {
		logger.Info("Ошибка при создании канала", err)
	}

	err = channel.ExchangeDeclare(
		cfg.RabbitConfig.TaskExchange, // Имя exchange
		"direct",     // Тип exchange (headers)
		false,         // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // аргументы
	)
	if err != nil {
		logger.Info("Ошибка при объявлении exchange", err)
	}

	keys := []string{"key1", "key2"}

	go func() {
		for {
			for _, val := range keys {
				body, err := json.Marshal(Task{Expression: "2+2", UserName: "user"})
				if err != nil {
					logger.Info("Encoding to JSON was failed", err)
				}

				message := amqp.Publishing{
					ContentType: "text/plain",
					Body:        body,
					Headers: amqp.Table{
						"worker": val, // Маркируем сообщение
					},
				}
		
				err = channel.PublishWithContext(
					context.Background(), cfg.RabbitConfig.TaskExchange, 
					val, false, false, message,
				)

				logger.Info("Send OK", message.Body)
				time.Sleep(1 * time.Second)
			}
		}
	}()

	rb.Connection = connection
	rb.Channel = channel
	return &rb
}

func (rb *RabbitManager) Close() {
	rb.Connection.Close()
	rb.Channel.Close()
}

func (rb *RabbitManager) SendTaskToSolver(userName, expression string, gto GetterTimeOfOperation) error {
	times, err := gto.GetTimeOfOperation(userName)
	if err != nil {
		return err
	}

	body, err := json.Marshal(Task{Expression: expression, UserName: userName, TimeOfOperation: *times})
	if err != nil {
		return err
	}

	message := amqp.Publishing{
		ContentType: "text/plain", Body: body,
		Headers: amqp.Table{
			"worker": strconv.Itoa(rb.KeyCounter), // Маркируем сообщение
		},
	}

	err = rb.Channel.PublishWithContext(
		context.Background(), rb.TaskExchange, 
		strconv.Itoa(rb.KeyCounter), false, false, message,
	)

	return nil
}