#!/bin/bash
set -euo pipefail

# Установить переменную TODO_PASSWORD, если она не установлена
export TODO_PASSWORD=${TODO_PASSWORD:-""}

# Получение токена
export TOKEN=$(curl -X POST http://localhost:7540/api/signin -H "Content-Type: application/json" -d "{\"password\":\"$TODO_PASSWORD\"}" --silent -c cookies.txt \
                | jq --raw-output .token)

# Запуск тестов
go test ./tests -count=1