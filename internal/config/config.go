// internal/config/config.go
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	NestAPIBase string
	OpenAIKey   string
}

var AppConfig *Config

func LoadConfig() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println(" .env file not found, relying on environment variables")
	}

	AppConfig = &Config{
		Port:        getEnv("PORT", "8089"),
		NestAPIBase: getEnv("NEST_API_URL", "https://api.ultahost.dev"),
		OpenAIKey:   getEnv("OPENAI_KEY", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// --- New helpers for integer envs (used for memory/queue sizing) ---

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

// Public convenience so other packages can read config ints with defaults.
func Int(key string, def int) int {
	return getEnvInt(key, def)
}
