package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() *sql.DB {
	// Получаем путь к базе данных из переменной окружения или используем стандартный путь
	dbPath := os.Getenv("TODO_DBFILE")
	if dbPath == "" {
		appPath, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		dbPath = filepath.Join(filepath.Dir(appPath), "scheduler.db")
	}

	// Проверяем, существует ли база данных
	log.Printf("Проверка существования файла базы данных: %s", dbPath)
	_, err := os.Stat(dbPath)
	install := false
	if err != nil {
		install = true
		log.Println("Файл базы данных не найден, база будет создана.")
	}

	// Открываем соединение с базой данных
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Если база данных не существует, создаём её
	if install {
		createTable(db)
	} else {
		log.Println("База данных существует, создание таблицы не требуется.")
	}

	return db
}

func createTable(db *sql.DB) {
	log.Println("Создание таблицы scheduler...")

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
