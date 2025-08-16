# TODO - переписать

# Используем официальный образ Go
FROM golang:1.24-alpine AS builder
# Устанавливаем рабочую директорию в контейнере
WORKDIR /app

COPY go.mod ./
RUN go mod download && go mod verify && go mod tidy

# Копируем файлы проекта
COPY . .

# Собираем приложение
RUN go build -o BOT ./cmd/bot/main.go
RUN go build -o MIGRATOR ./cmd/migrator/main.go
RUN go build -o KOFD_UPDATER ./cmd/kofd_updater/main.go

# Используем stage 2: минимальный контейнер
FROM alpine:3.21.3 AS final
WORKDIR /app/

# Добавляем необходимые зависимости
# tzdata - для установки временной зоны
# curl - для проверки доступности сервиса
RUN apk add --no-cache tzdata curl

# Копируем бинарники, миграции и конфиги из builder-образа
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/BOT .
COPY --from=builder /app/MIGRATOR .
COPY --from=builder /app/KOFD_UPDATER .

CMD ["sh", "-c", "./MIGRATOR -typeTask up -dsn $DSN && exec ./BOT"]