package tests

import "os"

// Port — порт, на котором работает приложение.
var Port = 7540

// DBFile — путь к файлу базы данных, который может быть переопределён переменной окружения TODO_DBFILE.
var DBFile = func() string {
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		return envDBFile
	}
	return "../scheduler.db"
}()

// FullNextDate — флаг, который определяет, включена ли полная обработка следующей даты.
var FullNextDate = true

// Search — флаг, включающий или отключающий функциональность поиска.
var Search = true

// Token — токен для авторизации, который может быть установлен через переменную окружения TOKEN.
var Token = func() string {
	if envToken := os.Getenv("TOKEN"); envToken != "" {
		return envToken
	}
	return ""
}()
