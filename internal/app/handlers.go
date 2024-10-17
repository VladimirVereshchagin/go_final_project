// internal/app/handlers.go

package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/VladimirVereshchagin/go_final_project/internal/models"
	"github.com/VladimirVereshchagin/go_final_project/internal/utils"
)

// writeJSONError отправляет ошибку в формате JSON с заданным статус-кодом
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		log.Println("Ошибка при кодировании JSON-ошибки:", err)
	}
}

// handleTask - маршрутизирует запросы на соответствующие обработчики в зависимости от метода
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
		// Метод не поддерживается
		writeJSONError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
	}
}

// addTaskHandler - обрабатывает добавление новой задачи
func (a *App) addTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	var task models.Task
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		// Ошибка при чтении тела запроса
		log.Println("Ошибка чтения JSON:", err)
		writeJSONError(w, http.StatusBadRequest, "Ошибка чтения JSON")
		return
	}

	// Проверка на наличие заголовка
	if task.Title == "" {
		writeJSONError(w, http.StatusBadRequest, "Не указан заголовок задачи")
		return
	}

	var err error

	// Создание задачи
	id, err := a.TaskService.CreateTask(&task)
	if err != nil {
		log.Println("Ошибка создания задачи:", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Возвращаем ID новой задачи
	response := map[string]interface{}{
		"id": id,
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(response)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Ошибка кодирования JSON")
	}
}

// getTaskHandler - обрабатывает получение задачи по ID
func (a *App) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		// Проверка на наличие ID
		writeJSONError(w, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	var err error

	// Поиск задачи по ID
	task, err := a.TaskService.GetTaskByID(id)
	if err != nil {
		log.Println("Задача не найдена:", err)
		writeJSONError(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	// Возвращаем найденную задачу
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(task)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Ошибка кодирования JSON")
	}
}

// editTaskHandler - обрабатывает редактирование задачи
func (a *App) editTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	var task models.Task
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		// Ошибка при чтении тела запроса
		log.Println("Ошибка чтения JSON:", err)
		writeJSONError(w, http.StatusBadRequest, "Ошибка чтения JSON")
		return
	}

	// Проверка на наличие ID и заголовка
	if task.ID == "" || task.Title == "" {
		writeJSONError(w, http.StatusBadRequest, "Не указан ID или заголовок задачи")
		return
	}

	var err error

	// Обновление задачи
	err = a.TaskService.UpdateTask(&task)
	if err != nil {
		log.Println("Ошибка обновления задачи:", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Успешное обновление
	response := map[string]string{}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(response)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Ошибка кодирования JSON")
	}
}

// deleteTaskHandler - обрабатывает удаление задачи
func (a *App) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Не указан идентификатор задачи")
		return
	}

	var err error

	// Удаление задачи
	err = a.TaskService.DeleteTask(id)
	if err != nil {
		log.Println("Ошибка удаления задачи:", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Успешное удаление
	response := map[string]string{}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(response)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Ошибка кодирования JSON")
	}
}

// handleTasks - обрабатывает получение списка задач
func (a *App) handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	search := r.URL.Query().Get("search")
	limit := 50

	var err error

	// Получаем список задач
	tasks, err := a.TaskService.ListTasks(search, limit)
	if err != nil {
		log.Println("Ошибка получения списка задач:", err)
		writeJSONError(w, http.StatusInternalServerError, "Ошибка получения списка задач")
		return
	}

	if tasks == nil {
		tasks = []*models.Task{}
	}

	// Возвращаем список задач
	response := map[string]interface{}{
		"tasks": tasks,
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(response)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Ошибка кодирования JSON")
	}
}

// handleDoneTask - обрабатывает отметку задачи как выполненной
func (a *App) handleDoneTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Не указан идентификатор задачи")
		return
	}

	var err error

	// Отметка задачи как выполненной
	err = a.TaskService.MarkTaskDone(id)
	if err != nil {
		log.Println("Ошибка при отметке задачи как выполненной:", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Успешное выполнение
	response := map[string]string{}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(response)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Ошибка кодирования JSON")
	}
}

// handleNextDate - обрабатывает вычисление следующей даты задачи
func (a *App) handleNextDate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var err error

	// Вычисляем следующую дату задачи
	nextDate, err := a.TaskService.CalculateNextDate(nowStr, dateStr, repeat)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Возвращаем следующую дату как простую строку
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}

// handleSignIn - обрабатывает аутентификацию пользователя
func (a *App) handleSignIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Чтение данных для входа
	var creds struct {
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&creds); err != nil {
		log.Println("Ошибка чтения JSON:", err)
		writeJSONError(w, http.StatusBadRequest, "Ошибка чтения JSON")
		return
	}

	pass := a.Config.Password

	if pass == "" {
		writeJSONError(w, http.StatusBadRequest, "Аутентификация не требуется")
		return
	}

	if creds.Password != pass {
		writeJSONError(w, http.StatusUnauthorized, "Неверный пароль")
		return
	}

	// Генерация JWT токена
	var err error
	tokenString, err := utils.GenerateToken(pass)
	if err != nil {
		log.Println("Ошибка генерации JWT токена:", err)
		writeJSONError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	// Возвращаем токен
	response := map[string]interface{}{
		"token": tokenString,
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(response)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		writeJSONError(w, http.StatusInternalServerError, "Ошибка кодирования JSON")
	}
}
