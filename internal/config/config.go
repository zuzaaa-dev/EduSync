package config

import (
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
)

// Config хранит настройки приложения
type Config struct {
	ServerPort    string
	DatabaseURL   string
	LogLevel      string
	JWTSecret     []byte
	UrlParserRKSI string
}

// LoadConfig загружает конфигурацию из .env или переменных окружения
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные окружения")
	}

	cfg := &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		JWTSecret:     []byte(getEnv("JWT_SECRET", "default-secret-key")),
		UrlParserRKSI: getEnv("URL_PARSER_RKSI", ""),
	}

	// Формируем DatabaseURL из компонентов
	dbUser := getEnv("DB_USER", "")
	dbPass := getEnv("DB_PASSWORD", "")
	dbHost := getEnv("DB_HOST", "")
	dbPort := getEnv("DB_PORT", "")
	dbName := getEnv("DB_NAME", "")

	if dbUser == "" || dbHost == "" || dbName == "" {
		return nil, errors.New("необходимо заполнить DB_USER, DB_HOST и DB_NAME")
	}

	cfg.DatabaseURL = formatDatabaseURL(dbUser, dbPass, dbHost, dbPort, dbName)

	return cfg, nil
}

// formatDatabaseURL формирует URL для подключения к БД
func formatDatabaseURL(user, pass, host, port, dbname string) string {
	auth := user
	if pass != "" {
		auth += ":" + pass
	}

	hostPort := host
	if port != "" {
		hostPort += ":" + port
	}

	return "postgres://" + auth + "@" + hostPort + "/" + dbname + "?sslmode=disable"
}

// getEnv получает переменную окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
