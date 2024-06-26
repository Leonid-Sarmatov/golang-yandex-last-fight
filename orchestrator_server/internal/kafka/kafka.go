package kafka

import (
	//"context"
	//"encoding/json"
	//"fmt"
	//"log/slog"
	//"strconv"
	//"time"

	//"github.com/IBM/sarama"

	//config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/config"
	//postgres "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/postgres"
)
/*
type MyMessage struct {
	Instructions string `json:"instructions"`
	FileName     string `json:"fileName"`
	FileData     string `json:"fileData"`
}

type KafkaManager struct {
	Produser         *sarama.AsyncProducer
	Admin            *sarama.ClusterAdmin
	TaskTopicName    string
	ResultTopicName  string
	PartitionCounter int
	PartitionNum     int
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

func NewKafkaManager(logger *slog.Logger, cfg *config.Config, sr SaverTaskResult) *KafkaManager {
	var kafkaManager KafkaManager
	kafkaManager.TaskTopicName = cfg.KafkaConfig.TaskTopicName
	kafkaManager.ResultTopicName = cfg.KafkaConfig.ResultTopicName
	kafkaManager.PartitionNum = cfg.KafkaConfig.PartitionNum

	// Создание настроек
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	// Создание продюсера отправляющего задачи
	producer, err := sarama.NewAsyncProducer([]string{
		cfg.KafkaConfig.Host + ":" + cfg.KafkaConfig.Port,
	}, config)
	if err != nil {
		logger.Info("Can not create new producer", err.Error())
	} // PartitionNum

	// Создаем админ клиент
	admin, err := sarama.NewClusterAdmin([]string{
		cfg.KafkaConfig.Host + ":" + cfg.KafkaConfig.Port,
	}, config)
	if err != nil {
		logger.Info("Error creating Kafka admin client", err.Error())
	}

	// Создаем топик с нужным количеством разделов
	err = admin.CreateTopic(cfg.KafkaConfig.TaskTopicName, &sarama.TopicDetail{
		NumPartitions:     2,//int32(cfg.KafkaConfig.PartitionNum),
		ReplicationFactor: 1,
	}, false)
	if err != nil {
		logger.Info("Error creating Kafka topic", err.Error())
	}

	go func() {
		for {
			for i := 0; i < 2; i += 1 {
				jsonMessage := MyMessage{
					Instructions: "write",
					FileName:     "data_1.txt",
					FileData:     strconv.Itoa(int(time.Now().Unix())),
				}

				jsonData, err := json.Marshal(jsonMessage)
				if err != nil {
					//log.Printf("[ERROR] Encoding to JSON was failed: %v\n", err)
				}

				message := &sarama.ProducerMessage{
					Topic: kafkaManager.TaskTopicName,
					Partition: int32(i),
					Key:   sarama.StringEncoder("key"+strconv.Itoa(i)),
					Value: sarama.ByteEncoder(jsonData),
				}

				fmt.Println("key"+strconv.Itoa(i))
				//fmt.Println("key"+string(i))

				producer.Input() <- message
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// Создаем поток с логом, для просмотра подтверждений сообщений
	go func() {
		for {
			select {
			case <-producer.Successes():
				logger.Info("Task sent was successful")
			}
		}
	}()

	// Создание настроек
	config = sarama.NewConfig()
	//config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	// Создание консумера, принимающего решенные задачи
	consumer, err := sarama.NewConsumerGroup([]string{
		cfg.KafkaConfig.Host + ":" + cfg.KafkaConfig.Port,
	}, "orchestrator-group", config)
	if err != nil {
		logger.Info("Can not create new conumer", err.Error())
	}

	// Создаем кастомный приемник сообщений
	handler := &MessageHandler{}
	handler.Saver = sr

	// Запуск потока приема сообщений с задачами
	go func() {
		for {
			err := consumer.Consume(context.Background(), []string{"topic-result"}, handler)
			if err != nil {
				logger.Info("Error in consumer: ", err.Error())
			}
		}
	}()

	// Заполняем поля структуры
	kafkaManager.Produser = &producer
	kafkaManager.Admin = &admin
	logger.Info("Kafka init - OK")
	return &kafkaManager
}

func (k *KafkaManager) Close() {
	(*k.Produser).Close()
	(*k.Admin).Close()
}

func (k *KafkaManager) SendTaskToSolver(userName, expression string, gto GetterTimeOfOperation) error {
	// Вычисляем номер партиции
	if k.PartitionCounter >= k.PartitionNum {
		k.PartitionCounter = 0
	}

	// Берем время выполнения для каждой операции
	timeOfOperation, err := gto.GetTimeOfOperation(userName)
	if err != nil {
		return err
	}

	// Создаем JSON для отправки в брокер сообщений
	jsonMessage, err := json.Marshal(Task{
		Expression:      expression,
		UserName:        userName,
		TimeOfOperation: *timeOfOperation,
	})
	if err != nil {
		return err
	}

	// Создаем сообщение с JSON
	message := &sarama.ProducerMessage{
		Topic:     k.TaskTopicName,
		Partition: int32(k.PartitionCounter),
		Key:       sarama.StringEncoder("key-"+strconv.Itoa(k.PartitionCounter)),
		Value:     sarama.ByteEncoder(jsonMessage),//+strconv.Itoa(k.PartitionCounter)
	}
	fmt.Printf("^^^^^^^^^Partition counter %v, Partition num %v\n", k.PartitionCounter, k.PartitionNum)
	k.PartitionCounter += 1

	// Отправляем сообщение
	(*k.Produser).Input() <- message
	return nil
}

// Структура для кастомного приемника сообщений
type MessageHandler struct {
	Saver SaverTaskResult
}

func (h *MessageHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *MessageHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h *MessageHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// Итерируемся по каналу с сообщениями
		if err := h.processMessage(message); err != nil {
			// Если обработка сообщения не успешна
			continue
		}
		// Если обработка успешна, фиксируем смещение
		session.MarkMessage(message, "")
	}
	return nil
}

func (h *MessageHandler) processMessage(message *sarama.ConsumerMessage) error {
	// Декодируем тело сообщения
	var res Result
	err := json.Unmarshal(message.Value, &res)
	if err != nil {
		return err
	}
	// Пробуем сохранить его в базе данных
	err = h.Saver.SaveTaskResult(res.UserName, res.Expression, res.Result)
	if err != nil {
		return err
	}
	return nil
}*/
