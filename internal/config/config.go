package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	OpenAIKey string
	Port      int
}

var appConfig AppConfig

func Load() error {

	//// Channge when finlaise with team like .env, need to implement error handling or placeholder value will change
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	appConfig.OpenAIKey = os.Getenv("OPENAI_API_KEY")
	//fmt.Println("appConfig.OpenAIKey : ", appConfig.OpenAIKey)
	portStr := os.Getenv("PORT")
	if portStr == "" {
		log.Printf("Invalid port value, defaulting to 8081")

		portStr = "8081"

	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Invalid port value, defaulting to 8081")
		port = 8081
	}
	appConfig.Port = port

	if appConfig.OpenAIKey == "" {
		log.Println("Warning: OPENAI_API_KEY is not set.")
	}

	return nil
}

func Get() AppConfig {
	return appConfig
}
