package main

import (
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/router"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		loggergrpc.LC.Log("api", "ERROR", ".env was not found", nil)
	}
}

func main() {

	apiPort := os.Getenv("API_PORT")

	router := router.CreateNewRouter()

	fmt.Println("Server is running on " + apiPort)
	http.ListenAndServe(":"+apiPort, router)
}
