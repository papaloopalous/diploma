package main

import (
	_ "api/internal/load_config"
	"api/internal/router"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

func main() {

	apiPort := viper.GetString("api.port")

	router := router.CreateNewRouter()

	log.Println("Server is running on " + apiPort)
	err := http.ListenAndServe(apiPort, router)
	if err != nil {
		log.Fatal("Server error: ", err)
	}
}
