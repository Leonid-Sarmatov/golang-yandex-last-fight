package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

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
	Connection     *amqp.Connection
	Channel        *amqp.Channel
	TaskExchange   string
	ResultExchange string
	Key            string
	KeyCounter     int
	KeyMax         int
	Expression string
}

type Heartbeat interface {
	Ping(string, string) error
}

func NewRabbitManager(key string, heartbeat Heartbeat) *RabbitManager {
	var rb RabbitManager
	rb.ResultExchange = "result_exchange"

	// Создаем тикер на одну секунду
	ticker := time.NewTicker(1 * time.Second)

	// Запускаем сердцебиение
	go func() {
		for {
			select {
			case <-ticker.C:
				err := heartbeat.Ping(key, rb.Expression)
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

	// Создаем канал для чтения задач
	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Ошибка при создании канала: %s", err)
	}

	// Создаем канал для отправки ответов
	resultChannel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Ошибка при создании канала: %s", err)
	}

	// Объявляем exchange для получения задач
	err = channel.ExchangeDeclare(
		"task_exchange", // Имя exchange
		"direct",        // Тип exchange (может быть direct, fanout, topic, headers)
		false,           // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // аргументы
	)
	if err != nil {
		log.Fatalf("Ошибка при объявлении exchange: %s", err)
	}

	// Объявляем exchange для отправки ответов
	/*err = resultChannel.ExchangeDeclare(
		rb.ResultExchange, // Имя exchange
		"direct",        // Тип exchange (может быть direct, fanout, topic, headers)
		false,           // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // аргументы
	)
	if err != nil {
		log.Fatalf("Ошибка при объявлении exchange: %s", err)
	}*/

	// Объявляем очередь для приема задач rb.ResultExchange
	q, err := channel.QueueDeclare("queue", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Объявляем очередь для отправки ответов
	resultQueue, err := resultChannel.QueueDeclare(rb.ResultExchange, false, false, false, false, nil)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Привязываем очередь к exchange
	err = channel.QueueBind(q.Name, key, "task_exchange", false, nil)
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

	log.Printf(q.Name)
	go func() {
		for d := range msgs {
			log.Printf("Body: %s, Key: %v", d.Body, key)
			//d.Ack(false)
			// Здесь должна быть логика обработки сообщения
			var task Task
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				log.Printf("Encoding to JSON was failed %v", err)
			}

			rb.Expression = task.Expression

			// Подтверждаем получение сообщения
			d.Ack(false)

			res, err := Calculate(task)
			var body []byte
			if err != nil {
				body, err = json.Marshal(Result{
					Expression: task.Expression,
					UserName:   task.UserName,
					Result:     "ERROR",
				})
				if err != nil {
					log.Printf("Encoding to JSON was failed %v", err)
				}
			} else {
				body, err = json.Marshal(Result{
					Expression: task.Expression,
					UserName:   task.UserName,
					Result:     strconv.FormatFloat(res, 'f', -1, 64),
				})
				if err != nil {
					log.Printf("Encoding to JSON was failed %v", err)
				}
			}

			message := amqp.Publishing{
				ContentType: "text/plain",
				Body:        body,
			}

			err = resultChannel.PublishWithContext(
				context.Background(), "",
				resultQueue.Name, false, false, message,
			)

			if err != nil {
				log.Printf("Send to orchestrator was failed %v", err)
			}

			rb.Expression = ""
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

func Calculate(task Task) (float64, error) {
	// Получаем выражение
	expression := task.Expression

	// Пролучаем время операций
	timesMap := make(map[string]int)
	timesMap["+"] = task.Addition
	timesMap["-"] = task.Subtraction
	timesMap["/"] = task.Division
	timesMap["*"] = task.Multiplication

	// Создаем синхронизатор
	wg := sync.WaitGroup{}

	// Создаем канал с ошибками, при передачи ошибки, нужно прервать выполнение вычисления
	errChan := make(chan error, 1)

	// Создаем массив чисел в виде строк
	stringArrayOfNumber := strings.FieldsFunc(expression, func(r rune) bool {
		return r == '+' || r == '-' || r == '/' || r == '*'
	})

	// Преобразуем числа из строкового представления в числовое
	arrayOfNumber := make([]float64, 0)
	for _, val := range stringArrayOfNumber {
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0.0, err
		}
		arrayOfNumber = append(arrayOfNumber, f)
	}

	// Создаем массив операций
	arrayOfOperation := make([]string, 0)
	for _, ch := range strings.Split(expression, "") {
		if ch == "+" || ch == "-" || ch == "/" || ch == "*" {
			arrayOfOperation = append(arrayOfOperation, ch)
		}
	}

	// Группа операций первого приоритета, пример:
	// В выражении 0+1*2+3+4-5*6/7 -> 1*2 и 5*6/7 являются группами
	// в этом выражении 1*2 и 5*6/7 будут сначала выполнены в отдельных потоках
	// и их результаты будут положены в: 0+_+3+4-_
	// после чего будет найдено финальное значение выражения

	// Счетчик количества групп состоящийх из операций второго приоритета
	groupCounter := 1
	for i := 0; i < len(arrayOfOperation); i += 1 {
		// Ищем операции второго приоритета
		if arrayOfOperation[i] == "+" || arrayOfOperation[i] == "-" {
			groupCounter += 1
		}
	}

	counter := 0
	operations := make([]string, 0) // Список операций группы первого приоритета
	numbers := make([]float64, 0)   // Список чисел группы первого приоритета
	begin := false                  // Нашли начало группы первого приоритета

	// Массив в который будут отправлены числа,
	// над которыми будут выполняться операции второго приоритета
	groupResultArray := make([]float64, groupCounter)
	// Массив с операциями второго приоритета
	groupOperatinArray := make([]string, groupCounter-1)
	for i := 0; i < len(arrayOfOperation); i += 1 {
		//fmt.Printf(" --- Итерация: %v --- \n", i)
		// Если нашли операцию второго приоритета, значит
		// записываем число с операцией в массив, либо, если мы до этого нашли
		// группу первого приоритета, ее надо вычислить и отправить ее
		// результат в массив вместо числа
		if arrayOfOperation[i] == "+" || arrayOfOperation[i] == "-" {
			if begin {
				// Если мы нашли операцию второго приоритета,
				// а до этого была операция первого приоритета,
				// то записываем в число в список группы первого приоритета
				numbers = append(numbers, arrayOfNumber[i])
				// Записываем операцию второго приоритета в массив
				groupOperatinArray[counter] = arrayOfOperation[i]
				// запускаем вычисление группы в отдельной горутине,
				// передав ей массивы со значениями и операциями,
				// а так же индекс, куда надо положить результат вычисления группы
				wg.Add(1)
				go func(numbers []float64, operations []string, counter int) {
					defer wg.Done()
					x, err := FirstPriority(numbers, operations, timesMap)
					if err != nil {
						errChan <- err
					}
					groupResultArray[counter] = x
				}(numbers, operations, counter)

				// Очищаем массивы для поиска следующей группы операций первого приоритета
				operations = make([]string, 0)
				numbers = make([]float64, 0)
				begin = false
				counter += 1
			} else {
				// Если найдена операция второго приоритета,
				// а до этого не быт открыт набор в группу первого приоритета,
				// то просто записываем текущий знак и текущее число
				groupResultArray[counter] = arrayOfNumber[i]
				groupOperatinArray[counter] = arrayOfOperation[i]
				counter += 1
			}

			// Если мы на конце массива с операциями, надо доподнительно записать
			// крайнее в выражении число
			if i == len(arrayOfOperation)-1 {
				groupResultArray[counter] = arrayOfNumber[i+1]
			}
		}

		// Если нашли операцию первого приоритета, создаем
		// группу первого приоритета, содержащую только операции первого порядка
		// Такая группа выполняется в отдельной горутине, однако внутри себя
		// группа может распраралелиться еще
		if arrayOfOperation[i] == "/" || arrayOfOperation[i] == "*" {
			if !begin {
				begin = true
			}
			// Добавляем число, после которого идет оператор первого приоритета
			numbers = append(numbers, arrayOfNumber[i])
			// Добавляем оператор после этого числа
			operations = append(operations, arrayOfOperation[i])

			// Если мы на конце, то есть выражение заканчивается произведением/делением
			// то добавляем крайнее число и запускаем подсчет
			if i == len(arrayOfOperation)-1 {
				numbers = append(numbers, arrayOfNumber[i+1])
				wg.Add(1)
				go func(numbers []float64, operations []string, counter int) {
					defer wg.Done()
					x, err := FirstPriority(numbers, operations, timesMap)
					if err != nil {
						errChan <- err
					}
					groupResultArray[counter] = x
				}(numbers, operations, counter)
			}
		}
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Отлавливаем ошибки, возникающие в процессе вычислений
	for i := range errChan {
		//log.Printf("Error %v", i)
		return 0.0, i
	}

	// Ждем пока посчитаются все операции первого приоритета
	// после чего получается массив чисел, над которыми остается
	// совершать только сложения и вычитания, то есть операции второго приоритета
	//wg.Wait()
	return SecondPriority(groupResultArray, groupOperatinArray, timesMap), nil
	//return 0.0, nil
}

/*
FirstPriority
*/
func FirstPriority(arrayOfNumber []float64, arrayOfOperation []string, timesMap map[string]int) (float64, error) {
	res := arrayOfNumber[0]
	for i := 0; i < len(arrayOfOperation); i += 1 {
		switch arrayOfOperation[i] {
		case "*":
			res *= arrayOfNumber[i+1]
			time.Sleep(time.Duration(timesMap["*"]) * time.Second)
		case "/":
			if arrayOfNumber[i+1] != 0.0 {
				res /= arrayOfNumber[i+1]
				time.Sleep(time.Duration(timesMap["/"]) * time.Second)
			} else {
				return 0.0, fmt.Errorf("Division by zero: %v / %v", res, arrayOfNumber[i+1])
			}
		}
	}
	return res, nil
}

/*
SecondPriority
*/
func SecondPriority(arrayOfNumber []float64, arrayOfOperation []string, timesMap map[string]int) float64 {
	res := arrayOfNumber[0]
	for i := 0; i < len(arrayOfOperation); i += 1 {
		switch arrayOfOperation[i] {
		case "+":
			res += arrayOfNumber[i+1]
			time.Sleep(time.Duration(timesMap["+"]) * time.Second)
		case "-":
			res -= arrayOfNumber[i+1]
			time.Sleep(time.Duration(timesMap["-"]) * time.Second)
		}
	}
	return res
}
