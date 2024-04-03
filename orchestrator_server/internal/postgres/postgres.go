package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

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
	ID int
	Expression string
	Status     int       `json:"status"`
	Result     string    `json:"result"`
	BeginTime  time.Time `json:"beginTime"`
	EndTime    time.Time `json:"endTime"`
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
	return nil
}

func (p *Postgres) GetTimeOfOperation() (*TimeOfOperation, error) {
	var timeOfOperation TimeOfOperation
	return &timeOfOperation, nil
}

func (p *Postgres) GetListOfTask() ([]Task, error) {
	listOfTask := make([]Task, 0)
	return listOfTask, nil
}

func (p *Postgres) SetTimeOfOperation(timeOfOperation TimeOfOperation) error {
	return nil
}
