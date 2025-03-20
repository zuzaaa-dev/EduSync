package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config хранит настройки приложения
type Config struct {
	ServerPort  string
	DatabaseURL string
	LogLevel    string
	JWTSecret   []byte
}

// LoadConfig загружает конфигурацию из .env или переменных окружения
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные окружения")
	}

	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		JWTSecret:   []byte(getEnv("JWT_SECRET", "default-secret-key")),
	}
}

// getEnv получает переменную окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
