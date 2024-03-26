package main

import (
	"os"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/frontend_server/internal/config"
)

func main() {
	// Задаем путь до конфигов
	os.Setenv("CONFIG_PATH", "./config/local.yaml")

	// Инициализируем конфиги
	cfg := config.MustLoad()
}
