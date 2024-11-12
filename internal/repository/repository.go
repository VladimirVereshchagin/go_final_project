package repository

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/VladimirVereshchagin/scheduler/internal/models"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const defaultLimit = 50 // Default limit value

// TaskRepository - interface for task operations
type TaskRepository interface {
	Create(task *models.Task) (string, error)
	GetByID(id string) (*models.Task, error)
	Update(task *models.Task) error
	Delete(id string) error
	List(search string, limit int) ([]*models.Task, error)
}

// taskRepository - implementation of the TaskRepository interface
type taskRepository struct {
	db *sqlx.DB
}

// NewTaskRepository - creates a new task repository
func NewTaskRepository(db *sqlx.DB) TaskRepository {
	return &taskRepository{db: db}
}

// NewDB - opens or creates a new database
func NewDB(dbPath string) (*sqlx.DB, error) {
	// If the database path is not provided, use the default path
	if dbPath == "" {
		dbPath = filepath.Join("data", "scheduler.db")
	}

	// Convert the database path to an absolute path
	absDBPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}
	dbPath = absDBPath

	// Create the directory if it does not exist
	dir := filepath.Dir(dbPath)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Open the database
	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil, err
	}

	// Check the connection
	if err := db.Ping(); err != nil {
		log.Printf("Error connecting to the database: %v", err)
		return nil, err
	}

	// Check if the table exists
	var exists int
	err = db.Get(&exists, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='scheduler'")
	if err != nil || exists == 0 {
		log.Println("Table 'scheduler' not found. Creating a new table.")
		createTable(db)
	} else {
		log.Println("Database and 'scheduler' table already exist.")
	}

	// Log the database file path after creating the table
	log.Printf("Using database file: %s", dbPath)

	return db, nil
}

// createTable - creates the task table if it does not exist
func createTable(db *sqlx.DB) {
	log.Println("Creating 'scheduler' table...")

	// SQL query to create the table and index
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
		log.Fatalf("Error creating table: %v", err)
	}
	log.Println("Table and index successfully created.")
}

// Create - adds a new task to the database
func (r *taskRepository) Create(task *models.Task) (string, error) {
	query := `
        INSERT INTO scheduler (date, title, comment, repeat)
        VALUES (:date, :title, :comment, :repeat)
    `
	res, err := r.db.NamedExec(query, task)
	if err != nil {
		return "", err
	}

	// Get the ID of the new task
	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", id), nil
}

// GetByID - retrieves a task by its ID
func (r *taskRepository) GetByID(id string) (*models.Task, error) {
	var task models.Task
	err := r.db.Get(&task, `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	task.ID = id
	return &task, nil
}

// Update - updates a task in the database
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

	// Check if any rows were updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// Delete - deletes a task by its ID
func (r *taskRepository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return err
	}

	// Check if any rows were deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// List - retrieves a list of tasks with filtering and limitation
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
		// Query without filtering
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
		// Filtering by date
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
		// Filtering by title or comment (case-insensitive Unicode)
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

		// Application-side filtering
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

// parseDate - parses a date in the format "dd.mm.yyyy"
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("02.01.2006", dateStr)
}

// isValidDate - checks if the string is a date in the format "dd.mm.yyyy"
func isValidDate(dateStr string) bool {
	_, err := parseDate(dateStr)
	return err == nil
}
