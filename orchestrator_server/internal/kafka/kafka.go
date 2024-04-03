package kafka

import (
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/config"
	postgres "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/postgres"
)

type KafkaManager struct {
	Produser *sarama.AsyncProducer
	TaskTopicName    string
	ResultTopicName    string
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
	GetTimeOfOperation() (*postgres.TimeOfOperation, error)
}

func NewKafkaManager(logger *slog.Logger, cfg *config.Config) *KafkaManager {
	var kafkaManager KafkaManager

	// Создание настроек
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	// Создание продюсера отправляющего задачи
	producer, err := sarama.NewAsyncProducer([]string{
		cfg.KafkaConfig.Host + ":" + cfg.KafkaConfig.Port,
	}, config)
	if err != nil {
		logger.Info("Can not create new producer", err.Error())
	}

	// Создание консумера, принимающего решенные задачи


	kafkaManager.Produser = &producer
	kafkaManager.TaskTopicName = cfg.KafkaConfig.TopicName
	logger.Info("Kafka init - OK")
	return &kafkaManager
}

func (k *KafkaManager) SendTaskToSolver(userName, expression string, gto GetterTimeOfOperation) error {
	// Берем время выполнения для каждой операции
	timeOfOperation, err := gto.GetTimeOfOperation()
	if err != nil {
		return err
	}

	// Создаем JSON для отправки в брокер сообщений
	jsonMessage, err := json.Marshal(Task{
		Expression:      expression,
		UserName:        userName,
		TimeOfOperation: *timeOfOperation,
	})

	// Создаем сообщение с JSON
	message := &sarama.ProducerMessage{
		Topic: k.TaskTopicName,
		Key:   sarama.StringEncoder("task"),
		Value: sarama.ByteEncoder(jsonMessage),
	}

	// Отправляем сообщение
	(*k.Produser).Input() <- message
	return nil
}
