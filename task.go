package main

// Task представляет задачу в планировщике
type Task struct {
	ID      string `json:"id"`                   // Уникальный идентификатор задачи
	Date    string `json:"date" db:"date"`       // Дата задачи
	Title   string `json:"title" db:"title"`     // Название задачи
	Comment string `json:"comment" db:"comment"` // Дополнительный комментарий к задаче
	Repeat  string `json:"repeat" db:"repeat"`   // Правило повторения задачи
}
