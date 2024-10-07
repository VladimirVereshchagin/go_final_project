package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату задачи по правилу повторения
func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	// Парсим дату задачи
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("некорректный формат даты: %v", err)
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}

	// Повторение ежегодно
	if repeat == "y" {
		nextDate := date.AddDate(1, 0, 0)
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		return nextDate.Format("20060102"), nil
	}

	// Повторение через n дней
	if strings.HasPrefix(repeat, "d ") {
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", fmt.Errorf("некорректный формат правила 'd'")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("некорректное количество дней: %v", err)
		}

		nextDate := date.AddDate(0, 0, days)
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}
		return nextDate.Format("20060102"), nil
	}

	// Повторение по дням недели
	if strings.HasPrefix(repeat, "w ") {
		daysOfWeek := parseDaysOfWeek(repeat[2:])
		if daysOfWeek == nil {
			return "", fmt.Errorf("некорректный формат правила 'w'")
		}

		nextDate := date.AddDate(0, 0, 1)
		for {
			if nextDate.After(now) {
				weekday := int(nextDate.Weekday())
				if weekday == 0 {
					weekday = 7 // Воскресенье - 7
				}
				if daysOfWeek[weekday] {
					return nextDate.Format("20060102"), nil
				}
			}
			nextDate = nextDate.AddDate(0, 0, 1)
			if nextDate.Sub(now).Hours() > 24*365*5 {
				return "", fmt.Errorf("не удалось найти следующую дату по правилу 'w'")
			}
		}
	}

	// Повторение по дням месяца
	if strings.HasPrefix(repeat, "m ") {
		days, months, err := parseMonthRule(repeat[2:])
		if err != nil {
			return "", fmt.Errorf("некорректный формат правила 'm': %v", err)
		}

		nextDate := date.AddDate(0, 0, 1)
		for {
			if nextDate.After(now) && isValidDayMonth(nextDate, days, months) {
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

// parseDaysOfWeek парсит строку с днями недели
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
	return days
}

// parseMonthRule парсит правило по месяцам и дням
func parseMonthRule(s string) ([]int, []int, error) {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil, nil, fmt.Errorf("пустое правило 'm'")
	}

	days, err := parseDaysOfMonth(parts[0])
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

// parseDaysOfMonth парсит строку с днями месяца
func parseDaysOfMonth(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	var days []int
	for _, part := range parts {
		day, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || day < -2 || day == 0 || day > 31 {
			return nil, fmt.Errorf("некорректный день месяца: %v", part)
		}
		days = append(days, day)
	}
	return days, nil
}

// parseMonths парсит строку с месяцами
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

// isValidDayMonth проверяет, входит ли дата в указанные дни и месяцы
func isValidDayMonth(date time.Time, days []int, months []int) bool {
	dayValid := false
	day := date.Day()
	lastDay := getLastDayOfMonth(date)

	for _, d := range days {
		var targetDay int
		if d > 0 {
			targetDay = d
		} else if d == -1 {
			targetDay = lastDay
		} else if d == -2 {
			targetDay = lastDay - 1
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

	for _, m := range months {
		if int(date.Month()) == m {
			return true
		}
	}

	return false
}

// getLastDayOfMonth возвращает последний день месяца
func getLastDayOfMonth(date time.Time) int {
	year, month, _ := date.Date()
	location := date.Location()
	firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, location)
	return firstOfNextMonth.AddDate(0, 0, -1).Day()
}
