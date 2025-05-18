package router

import (
	"api/internal/encryption"
	"api/internal/handlers"
	"api/internal/healthcheck"
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/middleware"
	"api/internal/repo"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// tokenRepo хранит глобальный репозиторий для работы с JWT токенами
var tokenRepo = repo.NewTokenRepo()

var coef int // коэффициент для таймаута соединения с микросервисами

// Адреса микросервисов
var userAddr, chatAddr, sessionAddr, taskAddr, loggerAddr string

// init инициализирует секретный ключ для JWT токенов
func init() {
	tokenRepo.SetData("secret jwt key")
	coef = viper.GetInt("api.timeout")
	userAddr = viper.GetString("user.addr")
	chatAddr = viper.GetString("chat.addr")
	sessionAddr = viper.GetString("session.addr")
	taskAddr = viper.GetString("task.addr")
	loggerAddr = viper.GetString("logger.addr")
}

// CreateNewRouter создает и настраивает роутер приложения
// Устанавливает соединения с микросервисами, инициализирует обработчики и настраивает маршруты
func CreateNewRouter() *mux.Router {
	// Создаем контекст с таймаутом для установки соединений
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(coef)*time.Second)
	defer cancel()

	// Устанавливаем соединения с микросервисами
	userConn, err := grpc.DialContext(ctx, userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}

	chatConn, err := grpc.DialContext(ctx, chatAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to chat service: %v", err)
	}

	sessionConn, err := grpc.DialContext(ctx, sessionAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to session service: %v", err)
	}

	taskConn, err := grpc.DialContext(ctx, taskAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to task service: %v", err)
	}

	loggerConn, err := grpc.DialContext(ctx, loggerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Printf("failed to connect to logger service: %v", err)
	}

	// Инициализируем проверку здоровья сервисов
	healthChecker := healthcheck.NewHealthChecker(5 * time.Second)
	healthChecker.AddConnection("user-service", userConn)
	healthChecker.AddConnection("task-service", taskConn)
	healthChecker.AddConnection("chat-service", chatConn)
	healthChecker.AddConnection("session-service", sessionConn)
	healthChecker.AddConnection("logger-service", loggerConn)

	loggergrpc.LC = loggergrpc.NewLogClient(loggerConn)

	// Создаем репозитории для работы с данными
	userRepo := repo.NewUserRepo(userConn)
	sessionRepo := repo.NewSessionRepo(sessionConn)
	taskRepo := repo.NewTaskRepo(taskConn)
	chatRepo := repo.NewChatRepo(chatConn)

	// Создаем обработчики запросов
	authHandler := &handlers.AuthHandler{
		User:    userRepo,
		Token:   tokenRepo,
		Session: sessionRepo,
	}

	taskHandler := &handlers.TaskHandler{
		User:  userRepo,
		Tasks: taskRepo,
	}

	userHandler := &handlers.UserHandler{
		User: userRepo,
	}

	middlewareHandler := &middleware.MiddlewareHandler{
		User:    userRepo,
		Session: sessionRepo,
		Token:   tokenRepo,
	}

	chatHandler := &handlers.ChatHandler{
		User: userRepo,
		Chat: chatRepo,
	}

	// Создаем основной роутер
	router := mux.NewRouter()

	// Настраиваем раздачу статических файлов
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// Маршруты для шифрования и аутентификации
	router.HandleFunc("/api/key-exchange", authHandler.EncryptionKey).Methods("POST")
	router.HandleFunc("/api/crypto-params", encryption.GetCryptoParams).Methods("GET")

	router.HandleFunc("/api/login", authHandler.LogIN).Methods("POST")
	router.HandleFunc("/api/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/logout", authHandler.LogOUT).Methods("DELETE")

	// Маршруты для всех авторизованных пользователей
	userRouter := router.NewRoute().Subrouter()
	userRouter.Use(middlewareHandler.CheckAny)
	userRouter.HandleFunc("/api/fill-profile", userHandler.FillProfile).Methods("POST")
	userRouter.HandleFunc("/api/get-profile", userHandler.GetProfile).Methods("GET")
	userRouter.HandleFunc("/api/get-tasks", taskHandler.OutAllTasks).Methods("GET")
	userRouter.HandleFunc("/api/download-task", taskHandler.DownloadTask).Methods("GET")
	userRouter.HandleFunc("/ws", chatHandler.HandleConnection).Methods("GET")
	userRouter.HandleFunc("/api/create-chat-room", chatHandler.CreateRoom).Methods("POST")

	// Маршруты только для студентов
	studentRouter := router.NewRoute().Subrouter()
	studentRouter.Use(middlewareHandler.CheckStudent)
	studentRouter.HandleFunc("/api/get-teachers", userHandler.OutAllTeachers).Methods("GET")
	studentRouter.HandleFunc("/api/get-my-teachers", userHandler.OutMyTeachers).Methods("GET")
	studentRouter.HandleFunc("/api/send-request", userHandler.AddRequest).Methods("POST")
	studentRouter.HandleFunc("/api/get-student-requests", userHandler.OutRequests).Methods("GET")
	studentRouter.HandleFunc("/api/upload-solution", taskHandler.AddSolution).Methods("POST")
	studentRouter.HandleFunc("/api/add-rating", userHandler.AddRating).Methods("POST")
	studentRouter.HandleFunc("/api/cancel-request", userHandler.CancelRequest).Methods("POST")

	// Маршруты только для преподавателей
	teacherRouter := router.NewRoute().Subrouter()
	teacherRouter.Use(middlewareHandler.CheckTeacher)
	teacherRouter.HandleFunc("/api/get-students", userHandler.OutAllStudents).Methods("GET")
	teacherRouter.HandleFunc("/api/get-teacher-requests", userHandler.OutRequests).Methods("GET")
	teacherRouter.HandleFunc("/api/confirm", userHandler.ConfirmRequest).Methods("POST")
	teacherRouter.HandleFunc("/api/deny", userHandler.DenyRequest).Methods("POST")
	teacherRouter.HandleFunc("/api/upload-task", taskHandler.CreateTask).Methods("POST")
	teacherRouter.HandleFunc("/api/download-solution", taskHandler.DownloadTask).Methods("GET")
	teacherRouter.HandleFunc("/api/add-grade", taskHandler.AddGrade).Methods("POST")

	// Маршруты для статических страниц
	router.HandleFunc("/", handlers.OutIndex)
	router.HandleFunc("/register", handlers.OutRegister)
	router.HandleFunc("/login", handlers.OutLogin)
	router.HandleFunc("/fill-profile", handlers.OutFillProfile)
	router.HandleFunc("/main", handlers.OutMain)
	router.HandleFunc("/task", handlers.OutTask)
	router.HandleFunc("/chat", handlers.OutChat)

	return router
}
