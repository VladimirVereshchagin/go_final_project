package repository

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/VladimirVereshchagin/go_final_project/internal/models"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const defaultLimit = 50 // Значение лимита по умолчанию

// TaskRepository - интерфейс для работы с задачами
type TaskRepository interface {
	Create(task *models.Task) (string, error)
	GetByID(id string) (*models.Task, error)
	Update(task *models.Task) error
	Delete(id string) error
	List(search string, limit int) ([]*models.Task, error)
}

// taskRepository - реализация интерфейса TaskRepository
type taskRepository struct {
	db *sqlx.DB
}

// NewTaskRepository - создаёт новый репозиторий задач
func NewTaskRepository(db *sqlx.DB) TaskRepository {
	return &taskRepository{db: db}
}

// NewDB - открывает или создаёт новую базу данных
func NewDB(dbPath string) (*sqlx.DB, error) {
	// Если путь к базе не указан, используем путь по умолчанию
	if dbPath == "" {
		dbPath = filepath.Join("data", "scheduler.db")
	}

	// Преобразуем путь к базе данных в абсолютный
	absDBPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}
	dbPath = absDBPath

	// Создаём директорию, если её нет
	dir := filepath.Dir(dbPath)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Открываем базу данных
	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		log.Printf("Ошибка открытия базы данных: %v", err)
		return nil, err
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		log.Printf("Ошибка соединения с базой данных: %v", err)
		return nil, err
	}

	// Проверяем наличие таблицы
	var exists int
	err = db.Get(&exists, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='scheduler'")
	if err != nil || exists == 0 {
		log.Println("Таблица scheduler не найдена. Создание новой таблицы.")
		createTable(db)
	} else {
		log.Println("База данных и таблица scheduler уже существуют.")
	}

	// Логирование пути к файлу базы данных после создания таблицы
	log.Printf("Используется файл базы данных: %s", dbPath)

	return db, nil
}

// createTable - создаёт таблицу задач, если её нет
func createTable(db *sqlx.DB) {
	log.Println("Создание таблицы scheduler...")

	// SQL-запрос для создания таблицы и индекса
	query := `
        CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT NOT NULL,
            title TEXT NOT NULL,
            comment TEXT,
            repeat TEXT DEFAULT '' NOT NULL
        );
        CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы: %v", err)
	}
	log.Println("Таблица и индекс успешно созданы.")
}

// Create - добавляет новую задачу в базу данных
func (r *taskRepository) Create(task *models.Task) (string, error) {
	query := `
        INSERT INTO scheduler (date, title, comment, repeat)
        VALUES (:date, :title, :comment, :repeat)
    `
	res, err := r.db.NamedExec(query, task)
	if err != nil {
		return "", err
	}

	// Получаем ID новой задачи
	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", id), nil
}

// GetByID - получает задачу по её ID
func (r *taskRepository) GetByID(id string) (*models.Task, error) {
	var task models.Task
	err := r.db.Get(&task, `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	task.ID = id
	return &task, nil
}

// Update - обновляет задачу в базе данных
func (r *taskRepository) Update(task *models.Task) error {
	query := `
        UPDATE scheduler
        SET date = :date, title = :title, comment = :comment, repeat = :repeat
        WHERE id = :id
    `
	result, err := r.db.NamedExec(query, task)
	if err != nil {
		return err
	}

	// Проверяем, обновлены ли строки
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}

// Delete - удаляет задачу по её ID
func (r *taskRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return err
	}

	// Проверяем, удалены ли строки
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}

// List - получает список задач, используя фильтрацию и ограничение
func (r *taskRepository) List(search string, limit int) ([]*models.Task, error) {
	var tasks []*models.Task
	var err error
	var query string
	var rows *sqlx.Rows

	if limit == 0 {
		limit = defaultLimit
	}

	params := map[string]interface{}{
		"limit": limit,
	}

	switch {
	case search == "":
		// Запрос без фильтрации
		query = `
            SELECT id, date, title, comment, repeat
            FROM scheduler
            ORDER BY date ASC
            LIMIT :limit
        `
		rows, err = r.db.NamedQuery(query, params)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var task models.Task
			err = rows.StructScan(&task)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, &task)
		}

	case isValidDate(search):
		// Фильтрация по дате
		date, _ := parseDate(search)
		params["date"] = date.Format("20060102")
		query = `
            SELECT id, date, title, comment, repeat
            FROM scheduler
            WHERE date = :date
            ORDER BY date ASC
            LIMIT :limit
        `
		rows, err = r.db.NamedQuery(query, params)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var task models.Task
			err = rows.StructScan(&task)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, &task)
		}

	default:
		// Фильтрация по заголовку или комментарию (нечувствительная к регистру Unicode)
		query = `
            SELECT id, date, title, comment, repeat
            FROM scheduler
            ORDER BY date ASC
            LIMIT :limit
        `
		rows, err = r.db.NamedQuery(query, params)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Фильтрация на стороне приложения
		searchLower := strings.ToLower(search)
		for rows.Next() {
			var task models.Task
			err = rows.StructScan(&task)
			if err != nil {
				return nil, err
			}

			titleLower := strings.ToLower(task.Title)
			commentLower := strings.ToLower(task.Comment)

			if strings.Contains(titleLower, searchLower) || strings.Contains(commentLower, searchLower) {
				tasks = append(tasks, &task)
				if len(tasks) >= limit {
					break
				}
			}
		}
	}

	return tasks, nil
}

// parseDate - парсит дату в формате "дд.мм.гггг"
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("02.01.2006", dateStr)
}

// isValidDate - проверяет, является ли строка датой в формате "дд.мм.гггг"
func isValidDate(dateStr string) bool {
	_, err := parseDate(dateStr)
	return err == nil
}
