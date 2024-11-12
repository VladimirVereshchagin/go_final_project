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
		title:   "Call at 16:00",
		comment: "Discuss plans",
		repeat:  "d 5",
	}

	todo := addTask(t, task)

	body, err := requestJSON("api/task", nil, http.MethodGet)
	assert.NoError(t, err)

	// Change: map[string]any instead of map[string]string to support any value types in response
	var m map[string]any
	err = json.Unmarshal(body, &m)
	assert.NoError(t, err)

	// Change: expecting an error instead of checking for its absence (assert.True instead of assert.False)
	_, ok := m["error"]
	assert.True(t, ok, "Expected error when no ID is provided")

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
		title:   "Order pizza",
		comment: "at 17:00",
		repeat:  "",
	}

	id := addTask(t, tsk)

	tbl := []fulltask{
		{"", task{"20240129", "Test", "", ""}},
		{"abc", task{"20240129", "Test", "", ""}},
		{"7645346343", task{"20240129", "Test", "", ""}},
		{id, task{"20240129", "", "", ""}},
		{id, task{"20240192", "Qwerty", "", ""}},
		{id, task{"28.01.2024", "Title", "", ""}},
		{id, task{"20240212", "Title", "", "ooops"}},
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

		// Change: assert.True instead of assert.False to check for error presence
		errVal, ok := m["error"]
		assert.True(t, ok && len(fmt.Sprint(errVal)) > 0, "Expected error for value %v", v)
	}

	updateTask := func(newVals map[string]any) {
		mupd, err := postJSON("api/task", newVals, http.MethodPut)
		assert.NoError(t, err)

		// Change: check for error with output if present
		if errVal, ok := mupd["error"]; ok && fmt.Sprint(errVal) != "" {
			t.Errorf("Unexpected error: %v", errVal)
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
			t.Errorf("Date cannot be earlier than today")
		}
	}

	updateTask(map[string]any{
		"id":      id,
		"date":    now.Format(`20060102`),
		"title":   "Order khinkali",
		"comment": "at 18:00",
		"repeat":  "d 7",
	})
}
