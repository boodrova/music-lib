package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	APIPort    string
	SwaggerAPI string
}

func LoadConfig() *Config {
	// Загружаем переменные из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("ERROR: Failed to load .env file. Ensure environment variables are set.")
	}

	// Читаем переменные окружения
	config := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		APIPort:    os.Getenv("API_PORT"),
		SwaggerAPI: os.Getenv("SWAGGER_API_URL"),
	}

	// Проверяем обязательные параметры
	missingVars := []string{}
	if config.DBHost == "" {
		missingVars = append(missingVars, "DB_HOST")
	}
	if config.DBPort == "" {
		missingVars = append(missingVars, "DB_PORT")
	}
	if config.DBUser == "" {
		missingVars = append(missingVars, "DB_USER")
	}
	if config.DBPassword == "" {
		missingVars = append(missingVars, "DB_PASSWORD")
	}
	if config.DBName == "" {
		missingVars = append(missingVars, "DB_NAME")
	}
	if config.APIPort == "" {
		missingVars = append(missingVars, "API_PORT")
	}

	if len(missingVars) > 0 {
		log.Fatalf("ERROR: Missing required environment variables: %v", missingVars)
	}

	return config
}
