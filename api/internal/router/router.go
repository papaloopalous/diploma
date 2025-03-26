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

	authHandler := &handlers.AuthHandler{
		User:    userRepo,
		Token:   tokenRepo,
		Session: sessionRepo,
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

	return router
}
