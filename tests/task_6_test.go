package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTask(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	now := time.Now()

	task := task{
		date:    now.Format(`20060102`),
		title:   "Созвон в 16:00",
		comment: "Обсуждение планов",
		repeat:  "d 5",
	}

	todo := addTask(t, task)

	body, err := requestJSON("api/task", nil, http.MethodGet)
	assert.NoError(t, err)

	// Изменение: map[string]any вместо map[string]string, чтобы поддерживать любые типы значений в ответе
	var m map[string]any
	err = json.Unmarshal(body, &m)
	assert.NoError(t, err)

	// Изменение: ожидается наличие ошибки вместо проверки её отсутствия (assert.True вместо assert.False)
	_, ok := m["error"]
	assert.True(t, ok, "Ожидается ошибка при отсутствии ID")

	body, err = requestJSON("api/task?id="+todo, nil, http.MethodGet)
	assert.NoError(t, err)
	var taskResp map[string]string
	err = json.Unmarshal(body, &taskResp)
	assert.NoError(t, err)

	assert.Equal(t, todo, taskResp["id"])
	assert.Equal(t, task.date, taskResp["date"])
	assert.Equal(t, task.title, taskResp["title"])
	assert.Equal(t, task.comment, taskResp["comment"])
	assert.Equal(t, task.repeat, taskResp["repeat"])
}

type fulltask struct {
	id string
	task
}

func TestEditTask(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	now := time.Now()

	tsk := task{
		date:    now.Format(`20060102`),
		title:   "Заказать пиццу",
		comment: "в 17:00",
		repeat:  "",
	}

	id := addTask(t, tsk)

	tbl := []fulltask{
		{"", task{"20240129", "Тест", "", ""}},
		{"abc", task{"20240129", "Тест", "", ""}},
		{"7645346343", task{"20240129", "Тест", "", ""}},
		{id, task{"20240129", "", "", ""}},
		{id, task{"20240192", "Qwerty", "", ""}},
		{id, task{"28.01.2024", "Заголовок", "", ""}},
		{id, task{"20240212", "Заголовок", "", "ooops"}},
	}
	for _, v := range tbl {
		m, err := postJSON("api/task", map[string]any{
			"id":      v.id,
			"date":    v.date,
			"title":   v.title,
			"comment": v.comment,
			"repeat":  v.repeat,
		}, http.MethodPut)
		assert.NoError(t, err)

		// Изменение: assert.True вместо assert.False для проверки наличия ошибки
		errVal, ok := m["error"]
		assert.True(t, ok && len(fmt.Sprint(errVal)) > 0, "Ожидается ошибка для значения %v", v)
	}

	updateTask := func(newVals map[string]any) {
		mupd, err := postJSON("api/task", newVals, http.MethodPut)
		assert.NoError(t, err)

		// Изменение: проверка ошибки с выводом при её наличии
		if errVal, ok := mupd["error"]; ok && fmt.Sprint(errVal) != "" {
			t.Errorf("Неожиданная ошибка: %v", errVal)
			return
		}

		var task Task
		err = db.Get(&task, `SELECT * FROM scheduler WHERE id=?`, id)
		assert.NoError(t, err)

		assert.Equal(t, id, strconv.FormatInt(task.ID, 10))
		assert.Equal(t, newVals["title"], task.Title)
		if _, is := newVals["comment"]; !is {
			newVals["comment"] = ""
		}
		if _, is := newVals["repeat"]; !is {
			newVals["repeat"] = ""
		}
		assert.Equal(t, newVals["comment"], task.Comment)
		assert.Equal(t, newVals["repeat"], task.Repeat)

		now := time.Now().Format(`20060102`)
		if task.Date < now {
			t.Errorf("Дата не может быть меньше сегодняшней")
		}
	}

	updateTask(map[string]any{
		"id":      id,
		"date":    now.Format(`20060102`),
		"title":   "Заказать хинкали",
		"comment": "в 18:00",
		"repeat":  "d 7",
	})
}
