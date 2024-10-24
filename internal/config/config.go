package config

import (
	"log"
	"os"

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
		// Если .env файл не найден, выводим сообщение и используем системные переменные
		log.Println("Не удалось загрузить .env файл, используем системные переменные")
	}

	return &Config{
		Port:     getEnv("TODO_PORT", "7540"),           // Загрузка порта с дефолтным значением 7540
		DBFile:   getEnv("TODO_DBFILE", "scheduler.db"), // Загрузка пути к базе данных
		Password: os.Getenv("TODO_PASSWORD"),            // Загрузка пароля для аутентификации
	}
}

// getEnv - получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		// Если переменная окружения не установлена, возвращаем значение по умолчанию
		return defaultValue
	}
	return value
}
