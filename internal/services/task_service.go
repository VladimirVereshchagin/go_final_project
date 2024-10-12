package services

import (
	"errors"
	"time"

	"github.com/VladimirVereshchagin/go_final_project/internal/models"
	"github.com/VladimirVereshchagin/go_final_project/internal/repository"
	"github.com/VladimirVereshchagin/go_final_project/internal/utils"
)

// TaskService - интерфейс для работы с задачами
type TaskService interface {
	CreateTask(task *models.Task) (string, error)
	GetTaskByID(id string) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
	ListTasks(search string, limit int) ([]*models.Task, error)
	MarkTaskDone(id string) error
	CalculateNextDate(nowStr, dateStr, repeat string) (string, error)
}

// taskService - реализация TaskService
type taskService struct {
	repo repository.TaskRepository // Репозиторий для работы с базой данных
}

// NewTaskService - создаёт новый сервис задач
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

// CreateTask - создаёт новую задачу
func (s *taskService) CreateTask(task *models.Task) (string, error) {
	// Устанавливаем текущую дату, если не указана
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	// Проверяем формат даты
	dateParsed, err := time.Parse("20060102", task.Date)
	if err != nil {
		return "", errors.New("некорректный формат даты")
	}
	dateParsed = time.Date(dateParsed.Year(), dateParsed.Month(), dateParsed.Day(), 0, 0, 0, 0, time.UTC)

	// Обработка правила повторения задачи
	if task.Repeat != "" {
		_, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return "", errors.New("некорректное правило повторения")
		}

		// Если задача повторяется и дата меньше текущей, пересчитываем дату
		if dateParsed.Before(now) {
			nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return "", errors.New("некорректное правило повторения")
			}
			task.Date = nextDate
		}
	} else {
		// Если дата меньше текущей, устанавливаем текущую дату
		if dateParsed.Before(now) {
			task.Date = now.Format("20060102")
		}
	}

	return s.repo.Create(task)
}

// GetTaskByID - получает задачу по её ID
func (s *taskService) GetTaskByID(id string) (*models.Task, error) {
	return s.repo.GetByID(id)
}

// UpdateTask - обновляет задачу
func (s *taskService) UpdateTask(task *models.Task) error {
	// Валидация задачи (аналогична CreateTask)
	if task.ID == "" {
		return errors.New("не указан идентификатор задачи")
	}

	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	// Проверка даты и повторения
	dateParsed, err := time.Parse("20060102", task.Date)
	if err != nil {
		return errors.New("некорректный формат даты")
	}

	// Обработка повторения
	if task.Repeat != "" {
		_, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return errors.New("некорректное правило повторения")
		}

		// Пересчёт даты
		if dateParsed.Before(now) {
			nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return errors.New("некорректное правило повторения")
			}
			task.Date = nextDate
		}
	} else if dateParsed.Before(now) {
		task.Date = now.Format("20060102")
	}

	return s.repo.Update(task)
}

// DeleteTask - удаляет задачу по ID
func (s *taskService) DeleteTask(id string) error {
	if id == "" {
		return errors.New("не указан идентификатор задачи")
	}
	return s.repo.Delete(id)
}

// ListTasks - возвращает список задач
func (s *taskService) ListTasks(search string, limit int) ([]*models.Task, error) {
	return s.repo.List(search, limit)
}

// MarkTaskDone - отмечает задачу как выполненную
func (s *taskService) MarkTaskDone(id string) error {
	if id == "" {
		return errors.New("не указан идентификатор задачи")
	}

	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if task.Repeat == "" {
		// Если задача одноразовая, удаляем её
		return s.repo.Delete(id)
	} else {
		// Если повторяющаяся, вычисляем следующую дату
		now := time.Now().UTC()
		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}

		task.Date = nextDate
		return s.repo.Update(task)
	}
}

// CalculateNextDate - вычисляет следующую дату задачи
func (s *taskService) CalculateNextDate(nowStr, dateStr, repeat string) (string, error) {
	if nowStr == "" || dateStr == "" || repeat == "" {
		return "", errors.New("отсутствуют параметры")
	}

	// Парсим текущую дату
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		return "", errors.New("некорректный параметр 'now'")
	}

	// Вычисляем следующую дату задачи
	return utils.NextDate(now, dateStr, repeat)
}
