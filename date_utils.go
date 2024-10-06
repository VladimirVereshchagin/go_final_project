package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату выполнения задачи на основе правила повторения
func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	// Парсим дату задачи
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("некорректный формат даты: %v", err)
	}

	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}

	// Правило ежегодного повторения
	if repeat == "y" {
		nextDate := date.AddDate(1, 0, 0) // Добавляем как минимум один год
		for {
			// Проверяем, существует ли дата (учёт 29 февраля)
			if nextDate.Month() != date.Month() || nextDate.Day() != date.Day() {
				// Дата не существует в этом году, корректируем на 1 марта
				nextDate = time.Date(nextDate.Year(), time.March, 1, 0, 0, 0, 0, nextDate.Location())
			}

			if nextDate.After(now) {
				return nextDate.Format("20060102"), nil
			}
			nextDate = nextDate.AddDate(1, 0, 0)
		}
	}

	// Правило повторения через определённое количество дней
	if strings.HasPrefix(repeat, "d ") {
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", fmt.Errorf("некорректный формат правила повторения 'd'")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("некорректное количество дней: %v", err)
		}

		nextDate := date.AddDate(0, 0, days) // Добавляем как минимум один период
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}

		return nextDate.Format("20060102"), nil
	}

	// Правило повторения по дням недели
	if strings.HasPrefix(repeat, "w ") {
		daysOfWeek := parseDaysOfWeek(repeat[2:])
		if daysOfWeek == nil {
			return "", fmt.Errorf("некорректный формат правила повторения 'w'")
		}

		nextDate := date.AddDate(0, 0, 1) // Начинаем со следующего дня
		for {
			weekday := int(nextDate.Weekday())
			if weekday == 0 {
				weekday = 7 // В Go воскресенье — это 0, приводим к 7
			}
			if daysOfWeek[weekday] && nextDate.After(now) {
				return nextDate.Format("20060102"), nil
			}
			nextDate = nextDate.AddDate(0, 0, 1)
			if nextDate.Sub(now).Hours() > 24*365*5 {
				return "", fmt.Errorf("не удалось найти следующую дату по правилу 'w'")
			}
		}
	}

	// Правило повторения по дням месяца
	if strings.HasPrefix(repeat, "m ") {
		days, months, err := parseMonthRule(repeat[2:])
		if err != nil {
			return "", fmt.Errorf("некорректный формат правила повторения 'm': %v", err)
		}

		nextDate := date.AddDate(0, 0, 1) // Начинаем со следующего дня
		for {
			if isValidDayMonth(nextDate, days, months) && nextDate.After(now) {
				return nextDate.Format("20060102"), nil
			}
			nextDate = nextDate.AddDate(0, 0, 1)
			if nextDate.Sub(now).Hours() > 24*365*5 {
				return "", fmt.Errorf("не удалось найти следующую дату по правилу 'm'")
			}
		}
	}

	return "", fmt.Errorf("неподдерживаемое правило повторения")
}

// parseDaysOfWeek парсит дни недели из строки и возвращает мапу
func parseDaysOfWeek(s string) map[int]bool {
	parts := strings.Split(s, ",")
	days := make(map[int]bool)
	for _, part := range parts {
		day, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || day < 1 || day > 7 {
			return nil
		}
		days[day] = true
	}
	if len(days) == 0 {
		return nil
	}
	return days
}

// parseMonthRule парсит правило повторения по месяцам
func parseMonthRule(s string) ([]int, []int, error) {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil, nil, fmt.Errorf("пустое правило 'm'")
	}

	daysStr := parts[0]
	days, err := parseDaysOfMonth(daysStr)
	if err != nil {
		return nil, nil, err
	}

	var months []int
	if len(parts) > 1 {
		months, err = parseMonths(parts[1])
		if err != nil {
			return nil, nil, err
		}
	}
	return days, months, nil
}

// parseDaysOfMonth парсит дни месяца из строки
func parseDaysOfMonth(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	var days []int
	for _, part := range parts {
		day, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || (day < -2 || day == 0 || day > 31) {
			return nil, fmt.Errorf("некорректный день месяца: %v", part)
		}
		days = append(days, day)
	}
	if len(days) == 0 {
		return nil, fmt.Errorf("не указаны дни месяца")
	}
	return days, nil
}

// parseMonths парсит месяцы из строки
func parseMonths(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	var months []int
	for _, part := range parts {
		month, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || month < 1 || month > 12 {
			return nil, fmt.Errorf("некорректный месяц: %v", part)
		}
		months = append(months, month)
	}
	return months, nil
}

// isValidDayMonth проверяет, соответствует ли дата заданным дням и месяцам
func isValidDayMonth(date time.Time, days []int, months []int) bool {
	dayValid := false
	day := date.Day()
	lastDay := getLastDayOfMonth(date)

	for _, d := range days {
		var targetDay int
		if d > 0 && d <= 31 {
			targetDay = d
		} else if d == -1 {
			targetDay = lastDay
		} else if d == -2 {
			targetDay = lastDay - 1
		} else {
			continue // Некорректный день
		}

		if day == targetDay {
			dayValid = true
			break
		}
	}

	if !dayValid {
		return false
	}

	if len(months) == 0 {
		return true
	}

	monthValid := false
	month := int(date.Month())
	for _, m := range months {
		if m == month {
			monthValid = true
			break
		}
	}

	return monthValid
}

// getLastDayOfMonth возвращает последний день месяца для заданной даты
func getLastDayOfMonth(date time.Time) int {
	year, month, _ := date.Date()
	location := date.Location()
	firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, location)
	return firstOfNextMonth.AddDate(0, 0, -1).Day()
}
