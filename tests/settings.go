package tests

import "os"

// Port - the port on which the application runs.
var Port = 7540

// DBFile - path to the database file for testing.
// Use the value of the TODO_DBFILE environment variable if set.
// Otherwise, use the default test database.
var DBFile = func() string {
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		return envDBFile
	}
	return "test_data/test_scheduler.db"
}()

// FullNextDate - flag that determines whether full next date processing is enabled.
var FullNextDate = true

// Search - flag that enables or disables search functionality.
var Search = true

// Token - authorization token that can be set via the TOKEN environment variable.
var Token = func() string {
	if envToken := os.Getenv("TOKEN"); envToken != "" {
		return envToken
	}
	return ""
}()
