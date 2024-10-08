# Этап сборки
FROM golang:1.20 AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=linux go build -o app .

# Финальный этап
FROM ubuntu:latest

# Устанавливаем необходимые зависимости
RUN apt-get update && apt-get install -y libsqlite3-0

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем исполняемый файл из этапа сборки
COPY --from=builder /app/app .

# Копируем директорию web (если она нужна для фронтенда)
COPY --from=builder /app/web ./web

# Открываем порт
EXPOSE 7540

# Команда для запуска приложения
CMD ["./app"]