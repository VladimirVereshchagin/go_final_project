package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	// Загружаем переменные из .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Не удалось загрузить .env файл, используем системные переменные")
	}

	// Инициализируем базу данных
	db = initDB()
	defer db.Close()

	// Определяем директорию с веб-ресурсами
	webDir := "./web"

	// Обрабатываем запросы для фронтенда
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Обработчик для /api/nextdate
	http.HandleFunc("/api/nextdate", nextDateHandler)

	// Обработчик для /api/task
	http.HandleFunc("/api/task", taskHandler)

	// Получаем порт из переменной окружения TODO_PORT или используем 7540
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Запускаем сервер
	log.Printf("Запуск сервера на порту %s...\n", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}

// Обработчик для маршрута /api/nextdate
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, "Отсутствуют параметры", http.StatusBadRequest)
		return
	}

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "Некорректный параметр 'now'", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, nextDate)
}

// Обработчик для /api/task
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		addTaskHandler(w, r)
	default:
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

// Обработчик для добавления задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	var task struct {
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	// Десериализуем JSON-запрос
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, `{"error":"Ошибка чтения JSON"}`, http.StatusBadRequest)
		return
	}

	// Проверяем обязательное поле title
	if task.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Устанавливаем текущую дату, если поле date не указано
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	// Парсим дату задачи
	dateParsed, err := time.Parse("20060102", task.Date)
	if err != nil {
		http.Error(w, `{"error":"Некорректный формат даты"}`, http.StatusBadRequest)
		return
	}
	dateParsed = time.Date(dateParsed.Year(), dateParsed.Month(), dateParsed.Day(), 0, 0, 0, 0, time.UTC)

	// Обрабатываем правило повторения
	if task.Repeat != "" {
		if dateParsed.Before(now) {
			// Вычисляем следующую дату после now
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
		// Если дата равна или после сегодняшней, оставляем её без изменений
	} else {
		// Если дата меньше сегодняшней, устанавливаем сегодняшнюю дату
		if dateParsed.Before(now) {
			task.Date = now.Format("20060102")
		}
	}

	// SQL-запрос с именованными параметрами
	query := `
        INSERT INTO scheduler (date, title, comment, repeat)
        VALUES (:date, :title, :comment, :repeat)
    `

	// Выполняем запрос с именованными параметрами
	res, err := db.Exec(query,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
	)
	if err != nil {
		log.Println("Ошибка базы данных:", err)
		http.Error(w, `{"error":"Ошибка базы данных"}`, http.StatusInternalServerError)
		return
	}

	// Получаем ID вставленной записи
	id, err := res.LastInsertId()
	if err != nil {
		log.Println("Ошибка получения ID задачи:", err)
		http.Error(w, `{"error":"Ошибка получения ID задачи"}`, http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	response := map[string]interface{}{
		"id": id,
	}
	json.NewEncoder(w).Encode(response)
}
