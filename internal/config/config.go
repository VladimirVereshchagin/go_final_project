package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config - структура для хранения конфигурационных данных
type Config struct {
	Port     string // Порт для запуска сервера
	DBFile   string // Файл базы данных
	Password string // Пароль для аутентификации
}

// LoadConfig - загружает конфигурацию из .env файла или системных переменных
func LoadConfig() *Config {
	// Загружаем переменные окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Println("Не удалось загрузить .env файл, используем системные переменные")
	}

	port := getEnv("TODO_PORT", "7540")
	if !isValidPort(port) {
		log.Fatalf("Недопустимый порт: %s", port)
	}

	password := os.Getenv("TODO_PASSWORD")
	if password == "" {
		log.Println("Внимание: Пароль для авторизации не задан. Доступ будет без аутентификации")
	}

	return &Config{
		Port:     port,
		DBFile:   getEnv("TODO_DBFILE", "data/scheduler.db"),
		Password: password,
	}
}

// getEnv - получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// isValidPort - проверяет, что порт находится в допустимом диапазоне
func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p > 0 && p <= 65535
}
