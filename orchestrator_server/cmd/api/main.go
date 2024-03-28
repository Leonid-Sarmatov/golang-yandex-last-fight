package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/config"
	login "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/handlers/login"
	registration "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/handlers/registration"
	cors_headers "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/middlewares/cors_headers"
	jwt_manager "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/jwt"
	postgres "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/postgres"
)

func main() {
	// Задаем путь до конфигов
	os.Setenv("CONFIG_PATH", "./config/local.yaml")

	// Инициализируем конфиги
	cfg := config.MustLoad()

	// Инициализируем логгер
	logger := setupLogger(cfg.EnvMode)
	logger.Debug("Successful read configurations.", slog.Any("cfg", cfg))

	// Создаем структуру для работы с JWT токенами
	jwtManager := jwt_manager.NewJWTManager()

	// Создаем структуру для работы с базой данных
	postgres, err := postgres.NewPostgres(logger, postgres.ConnectStringFromConfig(cfg))
	if err != nil {
		panic(err)
	}

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

	// Эндпоинт для входа в аккаунт
	router.Post("/login", login.NewLoginHandler(logger, jwtManager, postgres))

	// Эндпоинт для входа в аккаунт
	router.Post("/registration", registration.NewRegistrationHandler(logger, postgres))

	// Эндпоинт принимающий выражение
	// Эндпоинт возвращающий список со всеми задачами
	// Эндпоинт принимающий список со временем выполнения для каждой операции
	// Эндпоинт возвращающий список с вычислителями и информацией о них

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
