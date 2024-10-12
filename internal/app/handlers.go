package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/VladimirVereshchagin/go_final_project/internal/models"
	"github.com/VladimirVereshchagin/go_final_project/internal/utils"
)

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
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

// addTaskHandler - обрабатывает добавление новой задачи
func (a *App) addTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		// Ошибка при чтении тела запроса
		log.Println("Ошибка чтения JSON:", err)
		http.Error(w, `{"error":"Ошибка чтения JSON"}`, http.StatusBadRequest)
		return
	}

	// Проверка на наличие заголовка
	if task.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	// Создание задачи
	id, err := a.TaskService.CreateTask(&task)
	if err != nil {
		log.Println("Ошибка создания задачи:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Возвращаем ID новой задачи
	response := map[string]interface{}{
		"id": id,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		http.Error(w, `{"error":"Ошибка кодирования JSON"}`, http.StatusInternalServerError)
	}
}

// getTaskHandler - обрабатывает получение задачи по ID
func (a *App) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		// Проверка на наличие ID
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	// Поиск задачи по ID
	task, err := a.TaskService.GetTaskByID(id)
	if err != nil {
		log.Println("Задача не найдена:", err)
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	// Возвращаем найденную задачу
	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		http.Error(w, `{"error":"Ошибка кодирования JSON"}`, http.StatusInternalServerError)
	}
}

// editTaskHandler - обрабатывает редактирование задачи
func (a *App) editTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		// Ошибка при чтении тела запроса
		log.Println("Ошибка чтения JSON:", err)
		http.Error(w, `{"error":"Ошибка чтения JSON"}`, http.StatusBadRequest)
		return
	}

	// Проверка на наличие ID и заголовка
	if task.ID == "" || task.Title == "" {
		http.Error(w, `{"error":"Не указан ID или заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	// Обновление задачи
	err = a.TaskService.UpdateTask(&task)
	if err != nil {
		log.Println("Ошибка обновления задачи:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Успешное обновление
	w.Write([]byte(`{}`))
}

// deleteTaskHandler - обрабатывает удаление задачи
func (a *App) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	// Удаление задачи
	err := a.TaskService.DeleteTask(id)
	if err != nil {
		log.Println("Ошибка удаления задачи:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Успешное удаление
	w.Write([]byte(`{}`))
}

// handleTasks - обрабатывает получение списка задач
func (a *App) handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	search := r.URL.Query().Get("search")
	limit := 50

	// Получаем список задач
	tasks, err := a.TaskService.ListTasks(search, limit)
	if err != nil {
		log.Println("Ошибка получения списка задач:", err)
		http.Error(w, `{"error":"Ошибка получения списка задач"}`, http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []*models.Task{}
	}

	// Возвращаем список задач
	response := map[string]interface{}{
		"tasks": tasks,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		http.Error(w, `{"error":"Ошибка кодирования JSON"}`, http.StatusInternalServerError)
		return
	}
}

// handleDoneTask - обрабатывает отметку задачи как выполненной
func (a *App) handleDoneTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	// Отметка задачи как выполненной
	err := a.TaskService.MarkTaskDone(id)
	if err != nil {
		log.Println("Ошибка при отметке задачи как выполненной:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Успешное выполнение
	w.Write([]byte(`{}`))
}

// handleNextDate - обрабатывает вычисление следующей даты задачи
func (a *App) handleNextDate(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Вычисляем следующую дату задачи
	nextDate, err := a.TaskService.CalculateNextDate(nowStr, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем следующую дату
	w.Write([]byte(nextDate))
}

// handleSignIn - обрабатывает аутентификацию пользователя
func (a *App) handleSignIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	// Чтение данных для входа
	var creds struct {
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Println("Ошибка чтения JSON:", err)
		http.Error(w, `{"error":"Ошибка чтения JSON"}`, http.StatusBadRequest)
		return
	}

	pass := a.Config.Password

	if pass == "" {
		http.Error(w, `{"error":"Аутентификация не требуется"}`, http.StatusBadRequest)
		return
	}

	// Проверка пароля
	if creds.Password != pass {
		http.Error(w, `{"error":"Неверный пароль"}`, http.StatusUnauthorized)
		return
	}

	// Генерация JWT токена
	tokenString, err := utils.GenerateToken(pass)
	if err != nil {
		log.Println("Ошибка генерации JWT токена:", err)
		http.Error(w, `{"error":"Ошибка генерации токена"}`, http.StatusInternalServerError)
		return
	}

	// Возвращаем токен
	response := map[string]interface{}{
		"token": tokenString,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		http.Error(w, `{"error":"Ошибка кодирования JSON"}`, http.StatusInternalServerError)
		return
	}
}
