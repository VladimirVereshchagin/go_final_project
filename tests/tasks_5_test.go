package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func addTask(t *testing.T, task task) string {
	ret, err := postJSON("api/task", map[string]any{
		"date":    task.date,
		"title":   task.title,
		"comment": task.comment,
		"repeat":  task.repeat,
	}, http.MethodPost)
	assert.NoError(t, err)
	assert.NotNil(t, ret["id"])
	id := fmt.Sprint(ret["id"])
	assert.NotEmpty(t, id)
	return id
}

func getTasks(t *testing.T, search string) []map[string]string {
	url := "api/tasks"
	if len(search) > 0 {
		url += "?search=" + search
	}
	body, err := requestJSON(url, nil, http.MethodGet)
	assert.NoError(t, err)

	var m map[string][]map[string]string
	err = json.Unmarshal(body, &m)
	assert.NoError(t, err)
	return m["tasks"]
}

func TestTasks(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	_, err := db.Exec("DELETE FROM scheduler")
	assert.NoError(t, err)

	tasks := getTasks(t, "")
	assert.NotNil(t, tasks)
	assert.Empty(t, tasks)

	now := time.Now()
	addTask(t, task{
		date:    now.Format(`20060102`),
		title:   "Watch a movie",
		comment: "with popcorn",
		repeat:  "",
	})
	now = now.AddDate(0, 0, 1)
	date := now.Format(`20060102`)
	addTask(t, task{
		date:    date,
		title:   "Go to the pool",
		comment: "",
		repeat:  "",
	})
	addTask(t, task{
		date:    date,
		title:   "Pay utilities",
		comment: "",
		repeat:  "d 30",
	})
	tasks = getTasks(t, "")
	assert.Equal(t, 3, len(tasks))

	now = now.AddDate(0, 0, 2)
	date = now.Format(`20060102`)
	addTask(t, task{
		date:    date,
		title:   "Swim",
		comment: "Pool with a coach",
		repeat:  "d 7",
	})
	addTask(t, task{
		date:    date,
		title:   "Call the MC",
		comment: "Figure out the hot water issue",
		repeat:  "",
	})
	addTask(t, task{
		date:    date,
		title:   "Meet with Vasya",
		comment: "at 18:00",
		repeat:  "",
	})

	tasks = getTasks(t, "")
	assert.Equal(t, 6, len(tasks))

	tasks = getTasks(t, "MC")
	assert.Equal(t, 1, len(tasks))
	tasks = getTasks(t, now.Format(`02.01.2006`))
	assert.Equal(t, 3, len(tasks))
}
