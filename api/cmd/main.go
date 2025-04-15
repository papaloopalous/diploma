package main

import (
	"api/internal/router"
	"fmt"
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

	fmt.Println("Server is running on " + apiPort)
	http.ListenAndServe(":"+apiPort, router)
}
