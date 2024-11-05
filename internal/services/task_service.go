package services

import (
	"errors"
	"time"

	"github.com/VladimirVereshchagin/scheduler/internal/models"
	"github.com/VladimirVereshchagin/scheduler/internal/repository"
	"github.com/VladimirVereshchagin/scheduler/internal/timeutils"
)

// Константа для формата даты
const dateFormat = "20060102"

// TaskService предоставляет интерфейс для работы с задачами.
type TaskService interface {
	CreateTask(task *models.Task) (string, error)
	GetTaskByID(id string) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
	ListTasks(search string, limit int) ([]*models.Task, error)
	MarkTaskDone(id string) error
	CalculateNextDate(nowStr, dateStr, repeat string) (string, error)
}

// taskService реализует интерфейс TaskService.
type taskService struct {
	repo repository.TaskRepository // Репозиторий для работы с базой данных.
}

// NewTaskService создает новый сервис задач.
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

// CreateTask создает новую задачу и возвращает ее ID.
func (s *taskService) CreateTask(task *models.Task) (string, error) {
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if task.Date == "" {
		task.Date = now.Format(dateFormat)
	}

	dateParsed, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return "", errors.New("некорректный формат даты")
	}
	dateParsed = time.Date(dateParsed.Year(), dateParsed.Month(), dateParsed.Day(), 0, 0, 0, 0, time.UTC)

	if task.Repeat != "" {
		_, err := timeutils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return "", errors.New("некорректное правило повторения")
		}

		if dateParsed.Before(now) {
			nextDate, err := timeutils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return "", errors.New("некорректное правило повторения")
			}
			task.Date = nextDate
		}
	} else if dateParsed.Before(now) {
		task.Date = now.Format(dateFormat)
	}

	return s.repo.Create(task)
}

// GetTaskByID возвращает задачу по ее ID.
func (s *taskService) GetTaskByID(id string) (*models.Task, error) {
	return s.repo.GetByID(id)
}

// UpdateTask обновляет существующую задачу.
func (s *taskService) UpdateTask(task *models.Task) error {
	if task.ID == "" {
		return errors.New("не указан идентификатор задачи")
	}

	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if task.Date == "" {
		task.Date = now.Format(dateFormat)
	}

	dateParsed, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return errors.New("некорректный формат даты")
	}

	if task.Repeat != "" {
		_, err := timeutils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return errors.New("некорректное правило повторения")
		}

		if dateParsed.Before(now) {
			nextDate, err := timeutils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return errors.New("некорректное правило повторения")
			}
			task.Date = nextDate
		}
	} else if dateParsed.Before(now) {
		task.Date = now.Format(dateFormat)
	}

	return s.repo.Update(task)
}

// DeleteTask удаляет задачу по ее ID.
func (s *taskService) DeleteTask(id string) error {
	if id == "" {
		return errors.New("не указан идентификатор задачи")
	}
	return s.repo.Delete(id)
}

// ListTasks возвращает список задач с возможностью поиска и ограничения по количеству.
func (s *taskService) ListTasks(search string, limit int) ([]*models.Task, error) {
	return s.repo.List(search, limit)
}

// MarkTaskDone отмечает задачу как выполненную.
func (s *taskService) MarkTaskDone(id string) error {
	if id == "" {
		return errors.New("не указан идентификатор задачи")
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

// CalculateNextDate вычисляет следующую дату задачи на основе параметров.
func (s *taskService) CalculateNextDate(nowStr, dateStr, repeat string) (string, error) {
	if nowStr == "" || dateStr == "" || repeat == "" {
		return "", errors.New("отсутствуют параметры")
	}

	now, err := time.Parse(dateFormat, nowStr)
	if err != nil {
		return "", errors.New("некорректный параметр 'now'")
	}

	return timeutils.NextDate(now, dateStr, repeat)
}
