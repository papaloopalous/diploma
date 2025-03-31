package router

import (
	"api/internal/handlers"
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

	router := mux.NewRouter()

	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	router.HandleFunc("/api/encryption-key", handlers.EncryptionKey).Methods("GET")

	router.HandleFunc("/api/login", authHandler.LogIN).Methods("POST")
	router.HandleFunc("/api/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/logout", authHandler.LogOUT).Methods("DELETE")

	router.HandleFunc("/", handlers.OutIndex)

	//router.HandleFunc("/api/createUser", handlers.CreateUser).Methods("POST")

	router.HandleFunc("/chat", handlers.OutChat)

	router.HandleFunc("/task", handlers.OutTask)

	router.HandleFunc("/upload-task", taskHandler.UploadFile)
	router.HandleFunc("/download-task", taskHandler.DownloadFile)

	return router
}
