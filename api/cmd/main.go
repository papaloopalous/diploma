package main

import (
	"api/internal/router"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile("./config/config.yaml")
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
}

func main() {

	apiPort := viper.GetString("api.port")

	router := router.CreateNewRouter()

	log.Println("Server is running on " + apiPort)
	err := http.ListenAndServe(apiPort, router)
	if err != nil {
		log.Fatal("Server error: ", err)
	}
}
