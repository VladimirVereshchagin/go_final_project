package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

type Task struct {
	ID      int64  `db:"id"`
	Date    string `db:"date"`
	Title   string `db:"title"`
	Comment string `db:"comment"`
	Repeat  string `db:"repeat"`
}

func count(db *sqlx.DB) (int, error) {
	var count int
	err := db.Get(&count, `SELECT count(id) FROM scheduler`)
	return count, err
}

func openDB(t *testing.T) *sqlx.DB {
	dbfile := os.Getenv("TODO_DBFILE")
	if dbfile == "" {
		dbfile = "data/scheduler.db" // или путь по умолчанию
	}

	// Преобразуем путь к базе данных в абсолютный
	absDBPath, err := filepath.Abs(dbfile)
	if err != nil {
		t.Fatalf("Не удалось получить абсолютный путь к базе данных: %v", err)
	}

	// Создаём директорию, если её нет
	dir := filepath.Dir(absDBPath)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		t.Fatalf("Не удалось создать директорию для базы данных: %v", err)
	}

	// Открываем базу данных
	db, err := sqlx.Open("sqlite", absDBPath)
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		t.Fatalf("Ошибка соединения с базой данных: %v", err)
	}

	return db
}

func TestDB(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	before, err := count(db)
	assert.NoError(t, err)

	today := time.Now().Format(`20060102`)

	res, err := db.Exec(`INSERT INTO scheduler (date, title, comment, repeat)
        VALUES (?, 'Todo', 'Комментарий', '')`, today)
	assert.NoError(t, err)

	id, err := res.LastInsertId()
	assert.NoError(t, err)

	var task Task
	err = db.Get(&task, `SELECT * FROM scheduler WHERE id=?`, id)
	assert.NoError(t, err)
	assert.Equal(t, id, task.ID)
	assert.Equal(t, `Todo`, task.Title)
	assert.Equal(t, `Комментарий`, task.Comment)

	_, err = db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	assert.NoError(t, err)

	after, err := count(db)
	assert.NoError(t, err)

	assert.Equal(t, before, after)
}
