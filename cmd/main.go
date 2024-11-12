package main

import (
	"log"
	"net/http"

	"github.com/VladimirVereshchagin/scheduler/internal/app"
	"github.com/VladimirVereshchagin/scheduler/internal/config"
	"github.com/VladimirVereshchagin/scheduler/internal/repository"
	"github.com/VladimirVereshchagin/scheduler/internal/services"

	_ "modernc.org/sqlite"
)

func main() {
	// Loading configuration
	cfg := config.LoadConfig()

	// Initializing the database
	db, err := repository.NewDB(cfg.DBFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initializing repositories and services
	taskRepo := repository.NewTaskRepository(db)
	taskService := services.NewTaskService(taskRepo)

	// Initializing the application
	application := app.NewApp(taskService, cfg)

	// Starting the server
	log.Printf("Starting server on port %s...\n", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, application.Router); err != nil {
		log.Fatal("Server startup error: ", err)
	}
}
