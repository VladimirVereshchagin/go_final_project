package services

import (
	"errors"
	"time"

	"github.com/VladimirVereshchagin/scheduler/internal/models"
	"github.com/VladimirVereshchagin/scheduler/internal/repository"
	"github.com/VladimirVereshchagin/scheduler/internal/timeutils"
)

// Constant for date format
const dateFormat = "20060102"

// TaskService provides an interface for task operations.
type TaskService interface {
	CreateTask(task *models.Task) (string, error)
	GetTaskByID(id string) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
	ListTasks(search string, limit int) ([]*models.Task, error)
	MarkTaskDone(id string) error
	CalculateNextDate(nowStr, dateStr, repeat string) (string, error)
}

// taskService implements the TaskService interface.
type taskService struct {
	repo repository.TaskRepository // Repository for interacting with the database.
}

// NewTaskService creates a new task service.
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

// CreateTask creates a new task and returns its ID.
func (s *taskService) CreateTask(task *models.Task) (string, error) {
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if task.Date == "" {
		task.Date = now.Format(dateFormat)
	}

	dateParsed, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return "", errors.New("invalid date format")
	}
	dateParsed = time.Date(dateParsed.Year(), dateParsed.Month(), dateParsed.Day(), 0, 0, 0, 0, time.UTC)

	if task.Repeat != "" {
		_, err := timeutils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return "", errors.New("invalid repeat rule")
		}

		if dateParsed.Before(now) {
			nextDate, err := timeutils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return "", errors.New("invalid repeat rule")
			}
			task.Date = nextDate
		}
	} else if dateParsed.Before(now) {
		task.Date = now.Format(dateFormat)
	}

	return s.repo.Create(task)
}

// GetTaskByID returns a task by its ID.
func (s *taskService) GetTaskByID(id string) (*models.Task, error) {
	return s.repo.GetByID(id)
}

// UpdateTask updates an existing task.
func (s *taskService) UpdateTask(task *models.Task) error {
	if task.ID == "" {
		return errors.New("task ID is required")
	}

	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if task.Date == "" {
		task.Date = now.Format(dateFormat)
	}

	dateParsed, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return errors.New("invalid date format")
	}

	if task.Repeat != "" {
		_, err := timeutils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return errors.New("invalid repeat rule")
		}

		if dateParsed.Before(now) {
			nextDate, err := timeutils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return errors.New("invalid repeat rule")
			}
			task.Date = nextDate
		}
	} else if dateParsed.Before(now) {
		task.Date = now.Format(dateFormat)
	}

	return s.repo.Update(task)
}

// DeleteTask deletes a task by its ID.
func (s *taskService) DeleteTask(id string) error {
	if id == "" {
		return errors.New("task ID is required")
	}
	return s.repo.Delete(id)
}

// ListTasks returns a list of tasks with optional search and limit parameters.
func (s *taskService) ListTasks(search string, limit int) ([]*models.Task, error) {
	return s.repo.List(search, limit)
}

// MarkTaskDone marks a task as done.
func (s *taskService) MarkTaskDone(id string) error {
	if id == "" {
		return errors.New("task ID is required")
	}

	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if task.Repeat == "" {
		return s.repo.Delete(id)
	}

	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	nextDate, err := timeutils.NextDate(now, task.Date, task.Repeat)
	if err != nil {
		return err
	}

	task.Date = nextDate
	return s.repo.Update(task)
}

// CalculateNextDate calculates the next task date based on the provided parameters.
func (s *taskService) CalculateNextDate(nowStr, dateStr, repeat string) (string, error) {
	if nowStr == "" || dateStr == "" || repeat == "" {
		return "", errors.New("missing parameters")
	}

	now, err := time.Parse(dateFormat, nowStr)
	if err != nil {
		return "", errors.New("invalid 'now' parameter")
	}

	return timeutils.NextDate(now, dateStr, repeat)
}
