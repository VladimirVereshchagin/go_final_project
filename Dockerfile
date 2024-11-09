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
FROM ubuntu:22.04

# Добавляем метаданные для образа
LABEL org.opencontainers.image.source="https://github.com/VladimirVereshchagin/scheduler"
LABEL maintainer="Vladimir Vereshchagin <vlvereschagin06@gmail.com>"

# Устанавливаем рабочую директорию для финального контейнера
WORKDIR /app

# Копируем скомпилированное приложение из этапа сборки
COPY --from=builder /app/app .

# Копируем директорию web для фронтенда
COPY --from=builder /app/web ./web

# Устанавливаем переменную окружения для порта
ENV TODO_PORT=7540

# Запускаем скомпилированное приложение
CMD ["./app"]