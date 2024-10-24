# syntax=docker/dockerfile:1.10

# Этап сборки
FROM --platform=$BUILDPLATFORM golang:1.22 AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Устанавливаем аргументы сборки для передачи переменных среды
ARG TARGETOS
ARG TARGETARCH

# Устанавливаем переменные окружения для кросс-компиляции
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV CGO_ENABLED=0

# Копируем файлы go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код приложения
COPY . .

# Собираем Go-приложение
RUN go build -o app ./cmd

# Финальный этап
FROM --platform=$TARGETPLATFORM ubuntu:22.04

# Устанавливаем рабочую директорию для финального контейнера
WORKDIR /app

# Копируем скомпилированное приложение из этапа сборки
COPY --from=builder /app/app .

# Копируем директорию web для фронтенда (если требуется)
COPY --from=builder /app/web ./web

# Открываем порт для доступа к приложению
EXPOSE 7540

# Запускаем скомпилированное приложение
CMD ["./app"]