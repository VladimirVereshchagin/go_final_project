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
	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// Инициализация базы данных
	db, err := repository.NewDB(cfg.DBFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Инициализация репозиториев и сервисов
	taskRepo := repository.NewTaskRepository(db)
	taskService := services.NewTaskService(taskRepo)

	// Инициализация приложения
	application := app.NewApp(taskService, cfg)

	// Запуск сервера
	log.Printf("Запуск сервера на порту %s...\n", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, application.Router); err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
