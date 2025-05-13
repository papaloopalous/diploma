package main

import (
	"api/internal/router"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env was not found")
	}
}

func main() {

	apiPort := os.Getenv("API_PORT")

	router := router.CreateNewRouter()

	log.Println("Server is running on " + apiPort)
	err := http.ListenAndServe(":"+apiPort, router)
	if err != nil {
		log.Fatal("Server error: ", err)
	}
}
