package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	//"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	config "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/config"
	get_list_of_solvers "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/handlers/get_list_of_solvers"
	get_list_of_task "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/handlers/get_list_of_task"
	login "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/handlers/login"
	registration "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/handlers/registration"
	send_task "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/handlers/send_task"
	send_time_of_operation "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/handlers/send_time_of_operation"
	jwt_manager "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/jwt"
	rabbit "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/rabbit"
	cors_headers "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/middlewares/cors_headers"
	validate_token "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/middlewares/validate_token"
	postgres "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/postgres"
	grpc "github.com/Leonid-Sarmatov/golang-yandex-last-fight/orchestrator_server/internal/grpc"
)

func main() {
	// Задаем путь до конфигов   /golang_yandex_last_figth/orchestrator_server
	os.Setenv("CONFIG_PATH", "./orchestrator_server/config/local.yaml")

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

	// Создаем структуру для работы с брокером сообщений
	rabbitManager := rabbit.NewRabbitManager(logger, cfg, postgres)
	//defer rabbitManager.Close()
	//fuck(logger)

	// Создаем GRPC сервер
	grpcManager := grpc.NewGRPCManager(logger, cfg)

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

	router.Route("/api", func(r chi.Router) {
		// Подключаем middleware для проверки токена запроса
		r.Use(validate_token.ValidateJWTToken(logger, jwtManager))

		// Эндпоинт принимающий выражение
		r.Post("/sendTask", send_task.NewSendTaskHandler(logger, rabbitManager, postgres, postgres))

		// Эндпоинт возвращающий список со всеми задачами
		r.Get("/getListOfTask", get_list_of_task.NewGetListOfTaskHandler(logger, postgres))

		// Эндпоинт принимающий список со временем выполнения для каждой операции
		r.Post("/sendTimeOfOperations", send_time_of_operation.NewSendTimeOfOperationsHandler(logger, postgres))

		// Эндпоинт возвращающий список с вычислителями и информацией о них
		r.Get("/getListOfSolvers", get_list_of_solvers.NewGetListOfSolversHandler(logger, grpcManager))
	})

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

	// Создаем канал с сигналом об остановки сервиса
	//osSignalsChan := make(chan os.Signal, 1)
	//signal.Notify(osSignalsChan, os.Interrupt)

	// Ждем сигнал об остановке (Ctrl + C в терминале)
	//<-osSignalsChan
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

func fuck(logger *slog.Logger) {
	resp, err := http.Get("http://frontend_server:8081/registration")
	if err != nil || resp.StatusCode != http.StatusOK {
		// Если не удалочь отправить успешный запрос или отказано
		// в получении задачи то ждем две секунды, 
		// и пытаемся отправить запрос повторно
		logger.Info("[ERROR]: Can not connect to orkestrator", err.Error())

	} else {

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Info("[error]:", err.Error())
			return
		}

		logger.Info("html:", string(body))
	}
}
