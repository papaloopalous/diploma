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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var tokenRepo = repo.NewTokenRepo()

func init() {
	tokenRepo.SetData("biba")
}

func CreateNewRouter() *mux.Router {
	// userConn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	log.Fatalf("failed to connect to user service: %v", err)
	// }

	// sessionConn, err := grpc.NewClient("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	log.Fatalf("failed to connect to session service: %v", err)
	// }

	// taskConn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	log.Fatalf("failed to connect to task service: %v", err)
	// }

	// loggerConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	log.Printf("failed to connect to logger service: %v", err)
	// }

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userConn, err := grpc.DialContext(ctx, "localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}

	chatConn, err := grpc.DialContext(ctx, "localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to chat service: %v", err)
	}

	sessionConn, err := grpc.DialContext(ctx, "localhost:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to session service: %v", err)
	}

	taskConn, err := grpc.DialContext(ctx, "localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect to task service: %v", err)
	}

	loggerConn, err := grpc.DialContext(ctx, "localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Printf("failed to connect to logger service: %v", err)
	}

	healthChecker := healthcheck.NewHealthChecker(5 * time.Second)
	healthChecker.AddConnection("user-service", userConn)
	healthChecker.AddConnection("task-service", taskConn)
	healthChecker.AddConnection("chat-service", chatConn)
	healthChecker.AddConnection("session-service", sessionConn)
	healthChecker.AddConnection("logger-service", loggerConn)

	loggergrpc.LC = loggergrpc.NewLogClient(loggerConn)

	userRepo := repo.NewUserRepo(userConn)
	sessionRepo := repo.NewSessionRepo(sessionConn)
	taskRepo := repo.NewTaskRepo(taskConn)
	chatRepo := repo.NewChatRepo(chatConn)

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

	router := mux.NewRouter()

	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	router.HandleFunc("/api/key-exchange", authHandler.EncryptionKey).Methods("POST")
	router.HandleFunc("/api/crypto-params", encryption.GetCryptoParams).Methods("GET")

	router.HandleFunc("/api/login", authHandler.LogIN).Methods("POST")
	router.HandleFunc("/api/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/logout", authHandler.LogOUT).Methods("DELETE")

	//for all users
	userRouter := router.NewRoute().Subrouter()
	userRouter.Use(middlewareHandler.CheckAny)
	userRouter.HandleFunc("/api/fill-profile", userHandler.FillProfile).Methods("POST")
	userRouter.HandleFunc("/api/get-profile", userHandler.GetProfile).Methods("GET")
	userRouter.HandleFunc("/api/get-tasks", taskHandler.OutAllTasks).Methods("GET")
	userRouter.HandleFunc("/api/download-task", taskHandler.DownloadTask).Methods("GET")
	userRouter.HandleFunc("/ws", chatHandler.HandleConnection).Methods("GET")
	userRouter.HandleFunc("/api/create-chat-room", chatHandler.CreateRoom).Methods("POST")

	//for students
	studentRouter := router.NewRoute().Subrouter()
	studentRouter.Use(middlewareHandler.CheckStudent)
	studentRouter.HandleFunc("/api/get-teachers", userHandler.OutAllTeachers).Methods("GET")
	studentRouter.HandleFunc("/api/get-my-teachers", userHandler.OutMyTeachers).Methods("GET")
	studentRouter.HandleFunc("/api/send-request", userHandler.AddRequest).Methods("POST")
	studentRouter.HandleFunc("/api/get-student-requests", userHandler.OutRequests).Methods("GET")
	studentRouter.HandleFunc("/api/upload-solution", taskHandler.AddSolution).Methods("POST")
	studentRouter.HandleFunc("/api/add-rating", userHandler.AddRating).Methods("POST")
	studentRouter.HandleFunc("/api/cancel-request", userHandler.CancelRequest).Methods("POST")

	//for teachers
	teacherRouter := router.NewRoute().Subrouter()
	teacherRouter.Use(middlewareHandler.CheckTeacher)
	teacherRouter.HandleFunc("/api/get-students", userHandler.OutAllStudents).Methods("GET")
	teacherRouter.HandleFunc("/api/get-teacher-requests", userHandler.OutRequests).Methods("GET")
	teacherRouter.HandleFunc("/api/confirm", userHandler.ConfirmRequest).Methods("POST")
	teacherRouter.HandleFunc("/api/deny", userHandler.DenyRequest).Methods("POST")
	teacherRouter.HandleFunc("/api/upload-task", taskHandler.CreateTask).Methods("POST")
	teacherRouter.HandleFunc("/api/download-solution", taskHandler.DownloadTask).Methods("GET")
	teacherRouter.HandleFunc("/api/add-grade", taskHandler.AddGrade).Methods("POST")

	//static
	router.HandleFunc("/", handlers.OutIndex)
	router.HandleFunc("/register", handlers.OutRegister)
	router.HandleFunc("/login", handlers.OutLogin)
	router.HandleFunc("/fill-profile", handlers.OutFillProfile)
	router.HandleFunc("/main", handlers.OutMain)
	router.HandleFunc("/task", handlers.OutTask)
	router.HandleFunc("/chat", handlers.OutChat)

	return router
}
