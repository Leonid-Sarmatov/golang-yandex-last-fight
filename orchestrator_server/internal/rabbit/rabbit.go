package rabbit

import (
	"context"
	"encoding/json"

	//"time"

	//"log"
	"log/slog"
	//"strconv"

	amqp "github.com/rabbitmq/amqp091-go"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/config"
	postgres "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/postgres"
	grpc "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/grpc"
)

type RabbitManager struct {
	Connection     *amqp.Connection
	Channel        *amqp.Channel
	TaskExchange   string
	ResultExchange string
	KeyArray       []string
	KeyCounter     int
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

type GetterLivingSolvers interface {
	GetLivingSolvers() ([]*grpc.SolverInfo, error)
}

func NewRabbitManager(logger *slog.Logger, cfg *config.Config, sr SaverTaskResult, gls GetterLivingSolvers) *RabbitManager {
	var rb RabbitManager

	logger.Info("---NewRabbitManager---") // +cfg.RabbitConfig.Host+":"+cfg.RabbitConfig.Port+"/"

	// Устанавливаем соединение с сервером RabbitMQ "amqp://guest:guest@rabbitmq:5672/"
	connection, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		logger.Info("", err)
		return &rb
	}

	// Создаем канал для отправки задач
	channel, err := connection.Channel()
	if err != nil {
		logger.Info("", err)
		return &rb
	}

	// Создаем канал для приема ответов
	resultChannel, err := connection.Channel()
	if err != nil {
		logger.Info("", err)
		return &rb
	}

	//Объявляем exchange для отправки задач
	err = channel.ExchangeDeclare(
		cfg.RabbitConfig.TaskExchange, // Имя exchange
		"direct",                      // Тип exchange (headers)
		false,                         // durable
		false,                         // auto-deleted
		false,                         // internal
		false,                         // no-wait
		nil,                           // аргументы
	)
	if err != nil {
		logger.Info("", err)
		return &rb
	}

	//Объявляем exchange для приема ответов
	/*err = resultChannel.ExchangeDeclare(
		cfg.RabbitConfig.ResultExchange, // Имя exchange
		"direct",                      // Тип exchange (headers)
		false,                         // durable
		false,                         // auto-deleted
		false,                         // internal
		false,                         // no-wait
		nil,                           // аргументы
	)
	if err != nil {
		logger.Info("", err)
		return &rb
	}*/

	// Создаем очередь для приема сообщений
	q, err := resultChannel.QueueDeclare(
		cfg.RabbitConfig.ResultExchange, // Имя очереди
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // аргументы
	)
	if err != nil {
		
	}

	// Получаем сообщения из очереди
	msgs, err := resultChannel.Consume(
		q.Name, // queue
		"",     // consumer
		true,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		
	}

	// Создаем поток для записи ответов в базу данных
	go func() {
		for {
			select {
			case msg := <-msgs:
				var res Result
				err := json.Unmarshal(msg.Body, &res)
				if err != nil {
					logger.Info("Decoding JSON was failed", err)
				}
				err = sr.SaveTaskResult(res.UserName, res.Expression, res.Result)
				if err != nil {
					logger.Info("Save result was failed", err)
				}
			}
		}
	}()

	// Создаем массив с ключами для распределения задач по вычислителям 
	keys := make([]string, 0) 
	// Ждем пока появится хоть один вычислитель
	for arr, err := gls.GetLivingSolvers(); err != nil || len(arr) == 0; {

	}
	// Создаем список с ключами всех зарегистрировавшихся вычислителей
	arr, _ := gls.GetLivingSolvers()
	for _, solver := range arr {
		keys = append(keys, solver.Key)
	}

	rb.KeyArray = keys
	rb.Connection = connection
	rb.Channel = channel
	rb.TaskExchange = cfg.RabbitConfig.TaskExchange
	rb.ResultExchange = cfg.RabbitConfig.ResultExchange
	return &rb
}

func (rb *RabbitManager) Close() {
	rb.Connection.Close()
	rb.Channel.Close()
}

func (rb *RabbitManager) SendTaskToSolver(userName, expression string, gto GetterTimeOfOperation, gls GetterLivingSolvers) error {
	solversArray, err := gls.GetLivingSolvers()
	if err != nil {
		return err 
	}

	if rb.KeyCounter >= len(solversArray) {
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
			"worker": solversArray[rb.KeyCounter].Key, // Маркируем сообщение
		},
	}

	err = rb.Channel.PublishWithContext(
		context.Background(), rb.TaskExchange,
		solversArray[rb.KeyCounter].Key, false, false, message,
	)

	rb.KeyCounter += 1
	return nil
}
