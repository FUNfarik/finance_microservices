# Используем официальный go образ
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o main .

# Финальный этап - минимальный образ
FROM alpine:latest

# Устанавливаем SSL сертификаты для HTTPS запросов
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем скомпилированное приложение
COPY --from=builder /app/main .

# Открываем порт
EXPOSE 8001

# Запускаем приложение
CMD ["./main"]