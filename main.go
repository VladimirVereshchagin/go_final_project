package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	http.HandleFunc("/api/task", auth(taskHandler))
	http.HandleFunc("/api/tasks", auth(tasksHandler))
	http.HandleFunc("/api/task/done", auth(doneTaskHandler))
	http.HandleFunc("/api/signin", signHandler)

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

// auth middleware function
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if pass == "" {
			// Если пароль не задан, аутентификация не требуется
			next(w, r)
			return
		}

		// Получаем токен из куки
		cookie, err := r.Cookie("token")
		if err != nil {
			// Токен отсутствует
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value

		// Парсим и проверяем токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверяем метод подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Возвращаем ключ (пароль)
			return []byte(pass), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Проверяем claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Проверяем хэш пароля
		if claims["hash"] != generatePasswordHash(pass) {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Токен валиден
		next(w, r)
	})
}

// signHandler handles the /api/sign endpoint
func signHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	// Читаем тело запроса
	var creds struct {
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Println("Ошибка чтения JSON:", err)
		http.Error(w, `{"error":"Ошибка чтения JSON"}`, http.StatusBadRequest)
		return
	}

	// Получаем пароль из переменной окружения
	pass := os.Getenv("TODO_PASSWORD")

	if pass == "" {
		// Пароль не задан, аутентификация не требуется
		http.Error(w, `{"error":"Аутентификация не требуется"}`, http.StatusBadRequest)
		return
	}

	// Сравниваем пароли
	if creds.Password != pass {
		http.Error(w, `{"error":"Неверный пароль"}`, http.StatusUnauthorized)
		return
	}

	// Создаём JWT токен
	token := jwt.New(jwt.SigningMethodHS256)

	// Устанавливаем claims
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = "user"                               // Subject
	claims["exp"] = time.Now().Add(8 * time.Hour).Unix() // Истекает через 8 часов
	claims["hash"] = generatePasswordHash(pass)          // Хэш пароля для проверки

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(pass))
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

// generatePasswordHash generates a hash of the password
func generatePasswordHash(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// Остальной код остаётся без изменений

// taskHandler обрабатывает запросы для /api/task
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r) // Обработка добавления новой задачи
	case http.MethodGet:
		getTaskHandler(w, r) // Обработка получения задачи по ID
	case http.MethodPut:
		editTaskHandler(w, r) // Обработка редактирования задачи
	case http.MethodDelete:
		deleteTaskHandler(w, r) // Обработка удаления задачи
	default:
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

// doneTaskHandler обрабатывает отметку задачи как выполненной
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Некорректный идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	var task Task
	err = db.Get(&task, `SELECT * FROM scheduler WHERE id = ?`, id)
	if err != nil {
		log.Println("Задача не найдена:", err)
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	if task.Repeat == "" {
		// Если задача одноразовая, удаляем её
		_, err = db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
		if err != nil {
			log.Println("Ошибка удаления задачи:", err)
			http.Error(w, `{"error":"Ошибка удаления задачи"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// Если задача повторяющаяся, вычисляем следующую дату
		now := time.Now().UTC()
		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Println("Ошибка вычисления следующей даты:", err)
			http.Error(w, `{"error":"Ошибка вычисления следующей даты"}`, http.StatusInternalServerError)
			return
		}

		// Обновляем дату задачи
		_, err = db.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, nextDate, id)
		if err != nil {
			log.Println("Ошибка обновления задачи:", err)
			http.Error(w, `{"error":"Ошибка обновления задачи"}`, http.StatusInternalServerError)
			return
		}
	}

	// Возвращаем пустой JSON при успешном выполнении
	w.Write([]byte(`{}`))
}

// deleteTaskHandler обрабатывает удаление задачи
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Некорректный идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	// Проверяем существование задачи
	var task Task
	err = db.Get(&task, `SELECT id FROM scheduler WHERE id = ?`, id)
	if err != nil {
		log.Println("Задача не найдена:", err)
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	// Удаляем задачу
	_, err = db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		log.Println("Ошибка удаления задачи:", err)
		http.Error(w, `{"error":"Ошибка удаления задачи"}`, http.StatusInternalServerError)
		return
	}

	// Возвращаем пустой JSON при успешном удалении
	w.Write([]byte(`{}`))
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
