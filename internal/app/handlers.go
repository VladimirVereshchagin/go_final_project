package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/VladimirVereshchagin/scheduler/internal/auth"
	"github.com/VladimirVereshchagin/scheduler/internal/models"
)

const defaultLimit = 50 // Default limit value

// writeJSONError sends an error in JSON format with the specified status code
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		log.Println("Error encoding JSON error:", err)
	}
}

// handleTask routes requests to the corresponding handlers depending on the method
func (a *App) handleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.addTaskHandler(w, r)
	case http.MethodGet:
		a.getTaskHandler(w, r)
	case http.MethodPut:
		a.editTaskHandler(w, r)
	case http.MethodDelete:
		a.deleteTaskHandler(w, r)
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// addTaskHandler handles adding a new task
func (a *App) addTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	var task models.Task
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		log.Println("Error reading JSON:", err)
		writeJSONError(w, http.StatusBadRequest, "Error reading JSON")
		return
	}

	if task.Title == "" {
		writeJSONError(w, http.StatusBadRequest, "Task title is required")
		return
	}

	id, err := a.TaskService.CreateTask(&task)
	if err != nil {
		log.Println("Error creating task:", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]any{
		"id": id,
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		log.Println("Error encoding JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Error encoding JSON")
	}
}

// getTaskHandler handles getting a task by ID
func (a *App) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	task, err := a.TaskService.GetTaskByID(id)
	if err != nil {
		log.Println("Task not found:", err)
		writeJSONError(w, http.StatusNotFound, "Task not found")
		return
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(task); err != nil {
		log.Println("Error encoding JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Error encoding JSON")
	}
}

// editTaskHandler handles editing a task
func (a *App) editTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	var task models.Task
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		log.Println("Error reading JSON:", err)
		writeJSONError(w, http.StatusBadRequest, "Error reading JSON")
		return
	}

	if task.ID == "" || task.Title == "" {
		writeJSONError(w, http.StatusBadRequest, "Task ID or title is required")
		return
	}

	if err := a.TaskService.UpdateTask(&task); err != nil {
		log.Println("Error updating task:", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]string{
		"message": "Task updated successfully",
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		log.Println("Error encoding JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Error encoding JSON")
	}
}

// deleteTaskHandler handles deleting a task
func (a *App) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	if err := a.TaskService.DeleteTask(id); err != nil {
		log.Println("Error deleting task:", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]string{
		"message": "Task deleted successfully",
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		log.Println("Error encoding JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Error encoding JSON")
	}
}

// handleTasks handles getting a list of tasks
func (a *App) handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	search := r.URL.Query().Get("search")
	limit := defaultLimit

	tasks, err := a.TaskService.ListTasks(search, limit)
	if err != nil {
		log.Println("Error getting task list:", err)
		writeJSONError(w, http.StatusInternalServerError, "Error getting task list")
		return
	}

	if tasks == nil {
		tasks = []*models.Task{}
	}

	response := map[string]any{
		"tasks": tasks,
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		log.Println("Error encoding JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Error encoding JSON")
	}
}

// handleDoneTask handles marking a task as done
func (a *App) handleDoneTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	if err := a.TaskService.MarkTaskDone(id); err != nil {
		log.Println("Error marking task as done:", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]string{
		"message": "Task marked as done",
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		log.Println("Error encoding JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Error encoding JSON")
	}
}

// handleNextDate handles calculating the next task date
func (a *App) handleNextDate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	nextDate, err := a.TaskService.CalculateNextDate(nowStr, dateStr, repeat)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]string{"next_date": nextDate}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		log.Println("Error encoding JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Error encoding JSON")
	}
}

// handleSignIn handles user authentication
func (a *App) handleSignIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var creds struct {
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&creds); err != nil {
		log.Println("Error reading JSON:", err)
		writeJSONError(w, http.StatusBadRequest, "Error reading JSON")
		return
	}

	pass := a.Config.Password

	if pass == "" {
		writeJSONError(w, http.StatusBadRequest, "Authentication not required")
		return
	}

	if creds.Password != pass {
		writeJSONError(w, http.StatusUnauthorized, "Incorrect password")
		return
	}

	tokenString, err := auth.GenerateToken(pass)
	if err != nil {
		log.Println("Error generating JWT token:", err)
		writeJSONError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	response := map[string]any{
		"token": tokenString,
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		log.Println("Error encoding JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Error encoding JSON")
	}
}
