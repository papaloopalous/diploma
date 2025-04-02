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

	router.HandleFunc("/", handlers.OutIndex)
	router.HandleFunc("/register.html", handlers.OutRegister)
	router.HandleFunc("/login.html", handlers.OutLogin)
	router.HandleFunc("/fill-profile.html", handlers.OutFillProfile)
	router.HandleFunc("/main.html", handlers.OutMain)
	router.HandleFunc("/task.html", handlers.OutTask)

	userRouter := router.NewRoute().Subrouter()
	userRouter.Use(middlewareHandler.CheckTeacher)
	userRouter.HandleFunc("/api/fill-profile", userHandler.FillProfile).Methods("POST")

	userRouter.HandleFunc("/api/get-profile", userHandler.GetProfile).Methods("GET")

	//router.HandleFunc("/api/createUser", handlers.CreateUser).Methods("POST")

	router.HandleFunc("/chat", handlers.OutChat)

	router.HandleFunc("/task", handlers.OutTask)

	router.HandleFunc("/upload-task", taskHandler.CreateTask)
	router.HandleFunc("/download-task", taskHandler.DownloadFile)

	return router
}
