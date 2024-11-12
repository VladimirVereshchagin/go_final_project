#!/bin/bash
set -euo pipefail
set +m

# Set TODO_PASSWORD variable if not already set
export TODO_PASSWORD=${TODO_PASSWORD:-""}
export TODO_DBFILE="$(pwd)/test_data/test_scheduler.db"

# Create directory for test database
mkdir -p test_data

# Run the application in the background
nohup ./app &> nohup.out &
APP_PID=$! # Save the application PID

# Wait a bit to allow the application to start
sleep 10

# Check if the application is running
if ps -p $APP_PID > /dev/null; then
    echo "Application is running. PID: $APP_PID"
else
    echo "Failed to start the application. PID: $APP_PID"
    exit 1
fi

# Obtain token
export TOKEN=$(curl -X POST http://localhost:7540/api/signin -H "Content-Type: application/json" -d "{\"password\":\"$TODO_PASSWORD\"}" --silent -c cookies.txt | jq --raw-output .token)

if [[ -z "$TOKEN" ]]; then
    echo "Failed to obtain token."
    kill $APP_PID >/dev/null 2>&1 || true
    exit 1
fi

# Run tests
if go test ./tests -count=1; then
    echo "Tests passed successfully"
else
    echo "Tests failed"
    kill $APP_PID >/dev/null 2>&1 || true
    exit 1
fi

# Stop the application
kill $APP_PID >/dev/null 2>&1 || true
wait $APP_PID 2>/dev/null || true
echo "Application stopped. PID: $APP_PID"

# Clean up temporary files
rm -f nohup.out cookies.txt

# Delete the test database after tests
if [[ -f "$TODO_DBFILE" ]]; then
    rm "$TODO_DBFILE"
    echo "Test database $TODO_DBFILE deleted."
fi

# Remove test database directory if it is empty
rmdir test_data 2>/dev/null || true