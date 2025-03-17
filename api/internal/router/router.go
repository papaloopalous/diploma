package router

import (
	"api/internal/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func CreateNewRouter() *mux.Router {
	router := mux.NewRouter()

	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	router.HandleFunc("/api/encryption-key", handlers.EncryptionKey).Methods("GET")

	router.HandleFunc("/api/login", handlers.AuthHandler).Methods("POST")
	router.HandleFunc("/api/register", handlers.AuthHandler).Methods("POST")

	router.HandleFunc("/", handlers.OutIndex)

	return router
}
