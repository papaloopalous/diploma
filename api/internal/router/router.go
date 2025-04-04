package router

import (
	"api/internal/handlers"
	"api/internal/middleware"
	"api/internal/repo"
	"net/http"

	"github.com/gorilla/mux"
)

var tokenRepo = repo.NewTokenRepo()

func init() {
	tokenRepo.SetData("biba")
}

func CreateNewRouter() *mux.Router {
	userRepo := repo.NewUserRepo()
	sessionRepo := repo.NewSessionRepo()
	taskRepo := repo.NewTaskRepo()

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

	router := mux.NewRouter()

	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	router.HandleFunc("/api/encryption-key", handlers.EncryptionKey).Methods("GET")

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
