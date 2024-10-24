# Этап сборки
FROM --platform=$BUILDPLATFORM golang:1.22 AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Устанавливаем аргументы сборки для передачи переменных окружения
ARG TARGETOS
ARG TARGETARCH

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app ./cmd

# Финальный этап
FROM --platform=$TARGETPLATFORM ubuntu:22.04

# Устанавливаем необходимые зависимости
RUN apt-get update && apt-get install -y libsqlite3-0

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем исполняемый файл из этапа сборки
COPY --from=builder /app/app .

# Копируем директорию web (для фронтенда)
COPY --from=builder /app/web ./web

# Открываем порт
EXPOSE 7540

# Команда для запуска приложения
CMD ["./app"]