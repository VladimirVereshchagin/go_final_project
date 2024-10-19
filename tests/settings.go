package tests

import "os"

var Port = 7540

var DBFile = func() string {
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		return envDBFile
	}
	return "../scheduler.db"
}()

var FullNextDate = true
var Search = true

// Проверяем наличие токена в переменных окружения, если не найден, токен пустой
var Token = func() string {
	if envToken := os.Getenv("TOKEN"); envToken != "" {
		return envToken
	}
	return ""
}()
