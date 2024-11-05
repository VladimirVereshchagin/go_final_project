package app

import (
	"net/http"

	"github.com/VladimirVereshchagin/scheduler/internal/app/middleware"
	"github.com/VladimirVereshchagin/scheduler/internal/config"
	"github.com/VladimirVereshchagin/scheduler/internal/services"
)

// App представляет структуру приложения с его настройками и зависимостями
type App struct {
	Router      *http.ServeMux       // Маршрутизатор для обработки HTTP-запросов
	TaskService services.TaskService // Сервис для работы с задачами
	Config      *config.Config       // Конфигурация приложения
}

// NewApp - создаёт новое приложение и регистрирует маршруты
func NewApp(taskService services.TaskService, cfg *config.Config) *App {
	app := &App{
		Router:      http.NewServeMux(), // Инициализация маршрутизатора
		TaskService: taskService,        // Инициализация сервиса задач
		Config:      cfg,                // Загрузка конфигурации
	}
	app.registerRoutes() // Регистрация маршрутов
	return app
}

// registerRoutes - регистрирует маршруты для обработки запросов
func (a *App) registerRoutes() {
	// Статические файлы (фронтенд)
	webDir := "./web"
	a.Router.Handle("/", http.FileServer(http.Dir(webDir)))

	// API маршруты
	a.Router.HandleFunc("/api/nextdate", a.handleNextDate)                             // Расчёт следующей даты задачи
	a.Router.HandleFunc("/api/task", middleware.Auth(a.handleTask, a.Config))          // Работа с задачей (CRUD)
	a.Router.HandleFunc("/api/tasks", middleware.Auth(a.handleTasks, a.Config))        // Получение списка задач
	a.Router.HandleFunc("/api/task/done", middleware.Auth(a.handleDoneTask, a.Config)) // Отметка задачи как выполненной
	a.Router.HandleFunc("/api/signin", a.handleSignIn)                                 // Аутентификация пользователя
}
