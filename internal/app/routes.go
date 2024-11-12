package app

import (
	"net/http"

	"github.com/VladimirVereshchagin/scheduler/internal/app/middleware"
	"github.com/VladimirVereshchagin/scheduler/internal/config"
	"github.com/VladimirVereshchagin/scheduler/internal/services"
)

// App represents the application structure with its configuration and dependencies
type App struct {
	Router      *http.ServeMux       // Router for handling HTTP requests
	TaskService services.TaskService // Service for task operations
	Config      *config.Config       // Application configuration
}

// NewApp creates a new application and registers the routes
func NewApp(taskService services.TaskService, cfg *config.Config) *App {
	app := &App{
		Router:      http.NewServeMux(), // Initialize router
		TaskService: taskService,        // Initialize task service
		Config:      cfg,                // Load configuration
	}
	app.registerRoutes() // Register routes
	return app
}

// registerRoutes registers routes for request handling
func (a *App) registerRoutes() {
	// Static files (frontend)
	webDir := "./web"
	a.Router.Handle("/", http.FileServer(http.Dir(webDir)))

	// API routes
	a.Router.HandleFunc("/api/nextdate", a.handleNextDate)                             // Calculate the next task date
	a.Router.HandleFunc("/api/task", middleware.Auth(a.handleTask, a.Config))          // Handle task operations (CRUD)
	a.Router.HandleFunc("/api/tasks", middleware.Auth(a.handleTasks, a.Config))        // Get list of tasks
	a.Router.HandleFunc("/api/task/done", middleware.Auth(a.handleDoneTask, a.Config)) // Mark task as done
	a.Router.HandleFunc("/api/signin", a.handleSignIn)                                 // User authentication
}
