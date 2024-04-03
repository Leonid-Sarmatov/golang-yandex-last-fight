package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"sort"

	_ "github.com/lib/pq"
	bcrypt "golang.org/x/crypto/bcrypt"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/config"
)

/*
GenerateHash создает хеш из строки
*/
func GenerateHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

/*
GenerateHash сравнивает хеш с возможным паролем
*/
func CompareHashes(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false
	}

	return true
}

/*
ConnectStringFromConfig создает строку подключения
к базе данных на оснофе конфигов
*/
func ConnectStringFromConfig(config *config.Config) string {
	return fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		config.PostgresConfig.Host,
		config.PostgresConfig.Port,
		config.PostgresConfig.User,
		config.PostgresConfig.Password,
		config.PostgresConfig.DBname,
	)
}

type Postgres struct {
	DB *sql.DB
}

type User struct {
	ID       int
	UserName string
	Password string
}

type TimeOfOperation struct {
	Addition       int `json:"addition"`
	Subtraction    int `json:"subtraction"`
	Division       int `json:"division"`
	Multiplication int `json:"multiplication"`
}

type Task struct {
	ID         int
	Expression string    `json:"expression"`
	Status     int       `json:"status"`
	Result     string    `json:"result"`
	UserName   string    `json:"user_name"`
	BeginTime  time.Time `json:"begin_time"`
	EndTime    time.Time `json:"end_time"`
}

type settingsTimes struct {
	ID              int
	Operation       string
	TimeOfOperation int
}

func NewPostgres(logger *slog.Logger, connectString string) (*Postgres, error) {
	// Пробуем создать соединение с базой данных
	db, err := sql.Open("postgres", connectString)
	if err != nil {
		logger.Error("Spawn connection to database was failed", err)
		return nil, err
	}

	// Если удалось, то добавляем соединение в возвращаемую структуру
	postgres := &Postgres{
		DB: db,
	}

	// Если по какой то причине в базе нет таблицы с запросами
	// на вычисленя, то создаем таблицу
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS task_table (
        id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY, 
        expression VARCHAR(255), 
        status BIGINT,
		result VARCHAR(255),
		user_name VARCHAR(255),
		time_begin TIMESTAMP,
		time_end TIMESTAMP
    );`)

	// Если таблицу создать не удалось, то возвращаем соединение
	// и ошибку создания таблицы
	if err != nil {
		return postgres, err
	}

	// Если по какой то причине в базе нет таблицы с настройками
	// времени вычисленя, то создаем таблицу
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS operation_table (
        id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY, 
        operation VARCHAR(255), 
        time INT
    );`)

	// Если таблицу создать не удалось, то возвращаем соединение
	// и ошибку создания таблицы
	if err != nil {
		return postgres, err
	}

	// Если по какой то причине в базе нет таблицы с настройками
	// времени вычисленя, то создаем таблицу
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS users_table (
        id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY, 
        user_name VARCHAR(255), 
        password VARCHAR(255)
    );`)

	// Если таблицу создать не удалось, то возвращаем соединение
	// и ошибку создания таблицы
	if err != nil {
		return postgres, err
	}

	logger.Info("Postgres init - OK")

	return postgres, nil
}

/*
СheckAccountExist проверяет аккаунт на существование
*/
func (p *Postgres) СheckAccountExist(userName string) (bool, error) {
	rows, err := p.DB.Query("SELECT * FROM users_table WHERE user_name=$1", userName)
	if err != nil {
		return false, err
	}

	users := make([]User, 0)
	for rows.Next() {
		var u User
		err = rows.Scan(&u.ID, &u.UserName, &u.Password)
		if err != nil {
			return false, err
		}

		users = append(users, u)
	}

	if len(users) != 0 {
		return true, nil
	}

	return false, nil
}

/*
CreateNewAccount создает аккаунт в таблице пользователей
*/
func (p *Postgres) CreateNewAccount(userName, password string) error {
	passwordHash, err := GenerateHash(password)
	if err != nil {
		return err
	}

	_, err = p.DB.Exec(`INSERT INTO users_table (user_name, password) VALUES ($1, $2)`, userName, passwordHash)
	if err != nil {
		return err
	}

	return nil
}

/*
СheckAccountPassword проверяет введенный клиентом пароль от аккаунта
*/
func (p *Postgres) СheckAccountPassword(userName, password string) (bool, error) {
	rows, err := p.DB.Query("SELECT password FROM users_table WHERE user_name=$1", userName)
	if err != nil {
		return false, err
	}

	ps := make([]string, 0)
	for rows.Next() {
		p := ""
		err = rows.Scan(&p)
		if err != nil {
			return false, err
		}
		ps = append(ps, p)
	}

	ok := CompareHashes(ps[0], password)
	if ok {
		return true, nil
	}

	return false, nil
}

/*
SaveTask сохраняет задачу в базу данных
*/
func (p *Postgres) SaveTask(userName, expression string) error {
	// Находим время подсчета выражения
	t, err := p.FindExecutionTime(expression)
	if err != nil {
		return err
	}

	// Записываем задачу в таблицу
	nowTime := time.Now()
	_, err = p.DB.Exec(`INSERT INTO task_table (
		expression, status, result, user_name, time_begin, time_end
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		expression, 0, "", userName,
		nowTime.Format("2006-01-02 15:04:05"),
		nowTime.Add(t).Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}

	return nil
}

/*
FindExecutionTime находит время выполнения арифметического выражения
*/
func (p *Postgres) FindExecutionTime(expression string) (time.Duration, error) {
	t, err := p.GetTimeOfOperation()
	if err != nil {
		return time.Second, err
	}
	operationTimeMap := map[string]int{
		"+": t.Addition,
		"-": t.Subtraction,
		"/": t.Division,
		"*": t.Multiplication,
	}
	// Создаем массив операций
	arrayOfOperation := make([]string, 0)
	for _, ch := range strings.Split(expression, "") {
		if ch == "+" || ch == "-" || ch == "/" || ch == "*" {
			arrayOfOperation = append(arrayOfOperation, ch)
		}
	}
	x := make([]int, 0)
	t1 := 0
	t2 := 0
	for i, val := range arrayOfOperation {
		if val == "*" || val == "/" {
			t1 += operationTimeMap[val]
			if i == len(arrayOfOperation)-1 {
				x = append(x, t1)
				t1 = 0
			}
			if i < len(arrayOfOperation)-1 &&
				(arrayOfOperation[i+1] == "+" || arrayOfOperation[i+1] == "-") {
				x = append(x, t1)
				t1 = 0
			}
		}
		if val == "+" || val == "-" {
			t2 += operationTimeMap[val]
		}
	}
	sort.Slice(x, func(i, j int) bool { return i > j })
	if len(x) != 0 {
		t2 += x[0]
	}
	return time.Duration(t2) * time.Second, nil
}

/*
GetTimeOfOperation получает время выполнения каждой операции
*/
func (p *Postgres) GetTimeOfOperation() (*TimeOfOperation, error) {
	var timeOfOperation TimeOfOperation

	rows, err := p.DB.Query("SELECT * FROM operation_table")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var s settingsTimes
		err = rows.Scan(&s.ID, &s.Operation, &s.TimeOfOperation)
		if err != nil {
			return nil, err
		}
		
		switch s.Operation {
		case "+":
			timeOfOperation.Addition = s.TimeOfOperation
		case "-":
			timeOfOperation.Subtraction = s.TimeOfOperation
		case "/":
			timeOfOperation.Division = s.TimeOfOperation
		case "*":
			timeOfOperation.Multiplication = s.TimeOfOperation
		}
	}

	return &timeOfOperation, nil
}

/*
GetListOfTask получает список со всеми задачами
*/
func (p *Postgres) GetListOfTask(userName string) ([]Task, error) {
	rows, err := p.DB.Query("SELECT * FROM task_table WHERE user_name=$1", userName)
	if err != nil {
		return nil, err
	}

	tasks := make([]Task, 0)
	for rows.Next() {
		var t Task
		err = rows.Scan(&t.ID, &t.Expression, &t.Status, &t.Result, &t.UserName, &t.BeginTime, &t.EndTime)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

/*
SetTimeOfOperation записывает время выполнения операций
*/
func (p *Postgres) SetTimeOfOperation(timeOfOperation TimeOfOperation) error {
	isIntable := false
	var query string

	times := map[string]int{
		"+": timeOfOperation.Addition,
		"-": timeOfOperation.Subtraction,
		"/": timeOfOperation.Division,
		"*": timeOfOperation.Multiplication,
	}

	for key, val := range times {
		err := p.DB.QueryRow(
			"SELECT EXISTS (SELECT 1 FROM operation_table WHERE operation = $1)", key).Scan(&isIntable)
		if err != nil {
			return err
		}
		if isIntable {
			query = `UPDATE operation_table SET time = $2 WHERE operation = $1;`
		} else {
			query = `INSERT INTO operation_table (operation, time) VALUES ($1, $2);`
		}
		_, err = p.DB.Exec(query, key, val)
		if err != nil {
			return err
		}
	}
	
	return nil
}
