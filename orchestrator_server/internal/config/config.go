package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	EnvMode          string `yaml:"environment_mode" env-default:"local"`
	PostgresConfig   `yaml:"postgres"`
	RabbitConfig      `yaml:"rabbitmq"`
	HTTPServerConfig `yaml:"http_server"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" env-default:"postgres"`
	Port     string `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	DBname   string `yaml:"dbname" env-default:"database"`
}

type RabbitConfig struct {
	Host            string `yaml:"host" env-default:"kafka"`
	Port            string `yaml:"port" env-default:"9092"`
	TaskExchange    string `yaml:"task-exchange" env-default:"task-exchange"`
}

type HTTPServerConfig struct {
	Address           string        `yaml:"address" env-default:"localhost:8080"`
	RequestTimeout    time.Duration `yaml:"request_timeout" env-default:"4s"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	// Проверяем задан ли путь до конфиг-файла
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is empty")
	}

	// Проверяем существует ли конфиг-файл
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file is not exist: %v", configPath)
	}

	// Считываем конфиги
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Reading config was failed: %v", err)
	}
	return &cfg
}
