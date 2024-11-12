package models

// Task represents a task in the scheduler
type Task struct {
	ID      string `json:"id"`                   // Unique identifier for the task
	Date    string `json:"date" db:"date"`       // Task date
	Title   string `json:"title" db:"title"`     // Task title
	Comment string `json:"comment" db:"comment"` // Additional comment for the task
	Repeat  string `json:"repeat" db:"repeat"`   // Task repetition rule
}
