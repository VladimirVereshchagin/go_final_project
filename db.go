package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

// initDB инициализирует базу данных. Создаёт файл, если его нет.
func initDB() {
	// Получаем путь к базе данных
	dbPath := os.Getenv("TODO_DBFILE")
	if dbPath == "" {
		appPath, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		dbPath = filepath.Join(filepath.Dir(appPath), "scheduler.db")
	}

	// Проверяем наличие файла базы данных
	log.Printf("Проверка файла базы данных: %s", dbPath)
	_, err := os.Stat(dbPath)
	install := false
	if err != nil {
		install = true
		log.Println("Файл базы данных не найден. Создание новой базы.")
	}

	// Открываем базу данных
	db, err = sqlx.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Если базы нет, создаём таблицу
	if install {
		createTable(db)
	} else {
		log.Println("База данных уже существует.")
	}
}

// createTable создаёт таблицу и индекс, если их нет.
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
