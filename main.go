package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Загружаем переменные окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Println("Не удалось загрузить .env файл, используем системные переменные")
	}

	// Инициализируем базу данных
	initDB()
	defer db.Close()

	// Определяем директорию с веб-ресурсами
	webDir := "./web"

	// Регистрируем обработку статических файлов
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Регистрируем API обработчики
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/tasks", tasksHandler)

	// Устанавливаем порт из переменной окружения или по умолчанию
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Запускаем HTTP-сервер
	log.Printf("Запуск сервера на порту %s...\n", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}

// nextDateHandler обрабатывает запросы на получение следующей даты задачи
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Проверка на наличие всех параметров
	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, "Отсутствуют параметры", http.StatusBadRequest)
		return
	}

	// Парсим текущую дату
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "Некорректный параметр 'now'", http.StatusBadRequest)
		return
	}

	// Вычисляем следующую дату
	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем результат
	w.Write([]byte(nextDate))
}

// taskHandler обрабатывает запросы для /api/task
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r) // Обработка добавления новой задачи
	case http.MethodGet:
		getTaskHandler(w, r) // Обработка получения задачи по ID
	case http.MethodPut:
		editTaskHandler(w, r) // Обработка редактирования задачи
	default:
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

// tasksHandler возвращает список задач с возможностью поиска
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	search := r.URL.Query().Get("search")
	limit := 50

	var (
		tasks []Task
		err   error
		rows  *sqlx.Rows
	)

	// Поиск задач в зависимости от наличия параметра search
	if search == "" {
		// Выбираем все задачи, если параметр поиска отсутствует
		query := `
            SELECT id, date, title, comment, repeat
            FROM scheduler
            ORDER BY date ASC
            LIMIT :limit
        `
		rows, err = db.NamedQuery(query, map[string]interface{}{
			"limit": limit,
		})
	} else {
		// Проверка, является ли строка датой
		date, parseErr := time.Parse("02.01.2006", search)
		if parseErr == nil {
			// Поиск по дате
			dateStr := date.Format("20060102")
			query := `
                SELECT id, date, title, comment, repeat
                FROM scheduler
                WHERE date = :date
                ORDER BY date ASC
                LIMIT :limit
            `
			rows, err = db.NamedQuery(query, map[string]interface{}{
				"date":  dateStr,
				"limit": limit,
			})
		} else {
			// Поиск по заголовку или комментарию
			searchPattern := "%" + search + "%"
			query := `
                SELECT id, date, title, comment, repeat
                FROM scheduler
                WHERE title LIKE :search OR comment LIKE :search
                ORDER BY date ASC
                LIMIT :limit
            `
			rows, err = db.NamedQuery(query, map[string]interface{}{
				"search": searchPattern,
				"limit":  limit,
			})
		}
	}

	if err != nil {
		log.Println("Ошибка базы данных:", err)
		http.Error(w, `{"error":"Ошибка базы данных"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Чтение строк и создание списка задач
	for rows.Next() {
		var tempTask struct {
			ID      int64  `db:"id"`
			Date    string `db:"date"`
			Title   string `db:"title"`
			Comment string `db:"comment"`
			Repeat  string `db:"repeat"`
		}
		err := rows.StructScan(&tempTask)
		if err != nil {
			log.Println("Ошибка чтения данных:", err)
			http.Error(w, `{"error":"Ошибка чтения данных"}`, http.StatusInternalServerError)
			return
		}
		// Преобразование ID в строку и добавление задачи в список
		task := Task{
			ID:      strconv.FormatInt(tempTask.ID, 10),
			Date:    tempTask.Date,
			Title:   tempTask.Title,
			Comment: tempTask.Comment,
			Repeat:  tempTask.Repeat,
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		log.Println("Ошибка чтения строк:", err)
		http.Error(w, `{"error":"Ошибка чтения данных"}`, http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []Task{}
	}

	// Возвращаем список задач в формате JSON
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

// addTaskHandler обрабатывает добавление новой задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	// Декодируем запрос в структуру задачи
	var task struct {
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Println("Ошибка чтения JSON:", err)
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

	// Если дата не указана, используем текущую
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	// Проверка и обработка даты
	dateParsed, err := time.Parse("20060102", task.Date)
	if err != nil {
		http.Error(w, `{"error":"Некорректный формат даты"}`, http.StatusBadRequest)
		return
	}
	dateParsed = time.Date(dateParsed.Year(), dateParsed.Month(), dateParsed.Day(), 0, 0, 0, 0, time.UTC)

	// Обработка правила повторения задачи
	if task.Repeat != "" {
		// Проверяем корректность правила повторения
		_, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
			return
		}

		if dateParsed.Before(now) {
			// Если задача повторяется, вычисляем следующую дату
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	} else {
		// Если дата меньше текущей, используем сегодняшнюю
		if dateParsed.Before(now) {
			task.Date = now.Format("20060102")
		}
	}

	// SQL-запрос для вставки задачи
	query := `
		INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (:date, :title, :comment, :repeat)
	`

	// Выполнение запроса
	res, err := db.NamedExec(query, map[string]interface{}{
		"date":    task.Date,
		"title":   task.Title,
		"comment": task.Comment,
		"repeat":  task.Repeat,
	})
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

	// Отправляем успешный ответ с id задачи
	response := map[string]interface{}{
		"id": strconv.FormatInt(id, 10), // Преобразуем ID в строку
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		http.Error(w, `{"error":"Ошибка кодирования JSON"}`, http.StatusInternalServerError)
	}
}

// getTaskHandler обрабатывает получение задачи по идентификатору
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Некорректный идентификатор"}`, http.StatusBadRequest)
		return
	}

	var task Task
	err = db.Get(&task, `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id)
	if err != nil {
		log.Println("Задача не найдена:", err)
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	// Преобразуем ID в строку
	task.ID = strconv.FormatInt(id, 10)

	// Отправляем задачу в формате JSON
	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
		http.Error(w, `{"error":"Ошибка кодирования JSON"}`, http.StatusInternalServerError)
	}
}

// editTaskHandler обрабатывает редактирование задачи
func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	// Декодируем запрос в структуру задачи
	var task struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Println("Ошибка чтения JSON:", err)
		http.Error(w, `{"error":"Ошибка чтения JSON"}`, http.StatusBadRequest)
		return
	}

	// Проверяем обязательные поля
	if task.ID == "" {
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Некорректный идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	// Проверяем существование задачи
	var existingTask Task
	err = db.Get(&existingTask, `SELECT id FROM scheduler WHERE id = ?`, id)
	if err != nil {
		log.Println("Задача не найдена:", err)
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Если дата не указана, используем текущую
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	// Проверка и обработка даты
	dateParsed, err := time.Parse("20060102", task.Date)
	if err != nil {
		http.Error(w, `{"error":"Некорректный формат даты"}`, http.StatusBadRequest)
		return
	}
	dateParsed = time.Date(dateParsed.Year(), dateParsed.Month(), dateParsed.Day(), 0, 0, 0, 0, time.UTC)

	// Обработка правила повторения задачи
	if task.Repeat != "" {
		// Проверяем корректность правила повторения
		_, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
			return
		}

		if dateParsed.Before(now) {
			// Если задача повторяется, вычисляем следующую дату
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	} else {
		// Если дата меньше текущей, используем сегодняшнюю
		if dateParsed.Before(now) {
			task.Date = now.Format("20060102")
		}
	}

	// Обновляем задачу в базе данных
	query := `
		UPDATE scheduler
		SET date = :date, title = :title, comment = :comment, repeat = :repeat
		WHERE id = :id
	`
	result, err := db.NamedExec(query, map[string]interface{}{
		"id":      id,
		"date":    task.Date,
		"title":   task.Title,
		"comment": task.Comment,
		"repeat":  task.Repeat,
	})
	if err != nil {
		log.Println("Ошибка обновления задачи:", err)
		http.Error(w, `{"error":"Ошибка базы данных"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Ошибка получения количества затронутых строк:", err)
		http.Error(w, `{"error":"Ошибка базы данных"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	// Возвращаем пустой JSON при успешном обновлении
	w.Write([]byte(`{}`))
}
