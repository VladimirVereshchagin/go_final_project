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
var Token = os.Getenv("TOKEN")
