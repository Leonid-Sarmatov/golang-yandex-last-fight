package rabbit

import (
	"context"
	"encoding/json"
	//"time"

	//"log"
	"log/slog"
	"strconv"

	//"time"

	amqp "github.com/rabbitmq/amqp091-go"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/config"
	postgres "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/postgres"
)

type RabbitManager struct {
	Connection   *amqp.Connection
	Channel      *amqp.Channel
	TaskExchange string
	KeyArray     []string
	KeyCounter   int
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

	logger.Info("---NewRabbitManager---") // +cfg.RabbitConfig.Host+":"+cfg.RabbitConfig.Port+"/"

	// Устанавливаем соединение с сервером RabbitMQ "amqp://guest:guest@rabbitmq:5672/"
	connection, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		logger.Info("", err)
		return &rb
	}
    
	// Создаем канал
	channel, err := connection.Channel()
	if err != nil {
		logger.Info("", err)
		return &rb
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
		logger.Info("", err)
		return &rb
	}
	logger.Info("&&&&&&&&&&&&&", cfg.RabbitConfig.TaskExchange)
	
	keys := make([]string, 0)//[]string{"key1", "key2"}
	for i := 1; i <= cfg.RabbitConfig.QuantitySolvers; i += 1 {
		keys = append(keys, "Solver "+strconv.Itoa(i))
	}
	rb.KeyArray = keys

	/*go func() {
		for {
			if rb.KeyCounter >= len(rb.KeyArray) {
				rb.KeyCounter = 0
			}
			logger.Info("1111!")
			body, err := json.Marshal(Task{Expression: "2+2", UserName: "user"})
			if err != nil {
				continue
			}
		
			logger.Info("2222!")
			message := amqp.Publishing{
				ContentType: "text/plain", Body: body,
				Headers: amqp.Table{
					"worker": keys[rb.KeyCounter], // Маркируем сообщение
				},
			}
		
			logger.Info("3333!")
			err = rb.Channel.PublishWithContext(
				context.Background(), cfg.RabbitConfig.TaskExchange, 
				keys[rb.KeyCounter], false, false, message,
			)

			if err != nil {
				logger.Info("SSSSSSEEEEEENNNDDDD!", err)
			}
			rb.KeyCounter += 1

			time.Sleep(1 * time.Second)
		}
	}()*/

	rb.Connection = connection
	rb.Channel = channel
	rb.TaskExchange = cfg.RabbitConfig.TaskExchange
	//rb.Re = cfg.RabbitConfig.TaskExchange
	return &rb
}

func (rb *RabbitManager) Close() {
	rb.Connection.Close()
	rb.Channel.Close()
}

func (rb *RabbitManager) SendTaskToSolver(userName, expression string, gto GetterTimeOfOperation) error {
	if rb.KeyCounter >= len(rb.KeyArray) {
		rb.KeyCounter = 0
	}

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
			"worker": rb.KeyArray[rb.KeyCounter], // Маркируем сообщение
		},
	}

	err = rb.Channel.PublishWithContext(
		context.Background(), rb.TaskExchange, 
		rb.KeyArray[rb.KeyCounter], false, false, message,
	)

	rb.KeyCounter += 1
	return nil
}