#!/bin/bash
set -euo pipefail
set +m

# Установить переменную TODO_PASSWORD, если она не установлена
export TODO_PASSWORD=${TODO_PASSWORD:-""}

# Запуск приложения в фоне
nohup ./app &> nohup.out &
APP_PID=$! # Сохраняем PID приложения

# Подождем немного, чтобы приложение успело запуститься
sleep 10

# Проверяем, что приложение запущено
if ps -p $APP_PID > /dev/null; then
    echo "Приложение запущено. PID: $APP_PID"
else
    echo "Не удалось запустить приложение. PID: $APP_PID"
    exit 1
fi

# Получение токена
export TOKEN=$(curl -X POST http://localhost:7540/api/signin -H "Content-Type: application/json" -d "{\"password\":\"$TODO_PASSWORD\"}" --silent -c cookies.txt | jq --raw-output .token)

if [[ -z "$TOKEN" ]]; then
    echo "Не удалось получить токен."
    kill $APP_PID >/dev/null 2>&1 || true
    exit 1
fi

# Запуск тестов
if go test ./tests -count=1; then
    echo "Тесты пройдены успешно"
else
    echo "Тесты не пройдены"
    kill $APP_PID >/dev/null 2>&1 || true
    exit 1
fi

# Остановка приложения
kill $APP_PID >/dev/null 2>&1 || true
wait $APP_PID 2>/dev/null || true
echo "Приложение остановлено. PID: $APP_PID"

# Очистка временных файлов
rm -f nohup.out cookies.txt

# Удаление базы данных после тестов
if [[ -f "scheduler.db" ]]; then
    rm "scheduler.db"
    echo "База данных scheduler.db удалена."
fi