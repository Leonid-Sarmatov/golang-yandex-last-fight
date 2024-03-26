package main

import (
	"log/slog"
	"os"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/frontend_server/internal/config"
	login "github.com/Leonid-Sarmatov/golang-yandex-last-fight/frontend_server/internal/handlers/login"
	account "github.com/Leonid-Sarmatov/golang-yandex-last-fight/frontend_server/internal/handlers/account"
	cors_headers "github.com/Leonid-Sarmatov/golang-yandex-last-fight/frontend_server/internal/middlewares/cors_headers"
)

func main() {
	// Задаем путь до конфигов
	os.Setenv("CONFIG_PATH", "./config/local.yaml")

	// Инициализируем конфиги
	cfg := config.MustLoad()

	// Инициализируем логгер
	logger := setupLogger(cfg.EnvMode)
	logger.Debug("Successful read configurations.", slog.Any("cfg", cfg))

	// Инициализируем роутер
	router := chi.NewRouter()

	// Подключаем готовый middleware для логирования запросов
	router.Use(middleware.Logger)

	// Подключаем готовый middleware, который отлавливает возможные паники,
	// что бы избежать падение приложения
	router.Use(middleware.Recoverer)

	// Подключаем свой moddleware, который подключает CORS заголовки
	// что бы исключить возможные неполадки со стороны браузера
	router.Use(cors_headers.AddCorsHeaders())

	// Подключаем обработчик сайта для входа в аккаунт и регистрации
	router.Post("/login", login.NewLoginSiteHandler(logger, cfg))

	// Подключаем обработчик сайта для входа в аккаунт и регистрации
	router.Post("/account", account.NewAccountSiteHandler(logger, cfg))

	// Создаем сервер
	server := &http.Server{
		Addr:         cfg.HTTPServerConfig.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServerConfig.RequestTimeout,
		WriteTimeout: cfg.HTTPServerConfig.RequestTimeout,
		IdleTimeout:  cfg.HTTPServerConfig.ConnectionTimeout,
	}

	// Запускаем сервер
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Server was stoped")
	}
}

/*
setupLogger инициализирует логер
*/
func setupLogger(envMode string) *slog.Logger {
	var logger *slog.Logger

	switch envMode {
	case "local":
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "dev":
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prodaction":
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return logger
}
