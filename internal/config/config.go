package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config - structure for storing configuration data
type Config struct {
	Port     string // Port for server startup
	DBFile   string // Database file
	Password string // Password for authentication
}

// LoadConfig loads configuration from .env file or system variables
func LoadConfig() *Config {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env file, using system variables")
	}

	port := getEnv("TODO_PORT", "7540")
	if !isValidPort(port) {
		log.Fatalf("Invalid port: %s", port)
	}

	password := os.Getenv("TODO_PASSWORD")
	if password == "" {
		log.Println("Warning: Authentication password is not set. Access will be without authentication")
	}

	dbFile := getEnv("TODO_DBFILE", "data/scheduler.db")
	// Create the directory for the database if it does not exist
	err = os.MkdirAll(filepath.Dir(dbFile), os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating directory for database: %v", err)
	}

	return &Config{
		Port:     port,
		DBFile:   dbFile,
		Password: password,
	}
}

// getEnv gets the value of an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// isValidPort checks if the port is within the valid range
func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p > 0 && p <= 65535
}
