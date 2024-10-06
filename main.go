package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные из .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Не удалось загрузить .env файл, используем системные переменные")
	}

	// Инициализируем базу данных
	db := initDB()
	defer db.Close()

	// Определяем директорию с веб-ресурсами
	webDir := "./web"

	// Обрабатываем запросы для фронтенда
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Обработчик для /api/nextdate
	http.HandleFunc("/api/nextdate", nextDateHandler)

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
