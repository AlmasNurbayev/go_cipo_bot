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
RUN go build -o SERVER ./cmd/server/main.go
RUN go build -o MIGRATOR ./cmd/migrator/main.go
RUN go build -o PARSER ./cmd/parser/main.go
RUN go build -o SEEDER ./cmd/seeder/main.go
RUN go build -o CLEARDB ./cmd/cleardb/main.go


# Используем stage 2: минимальный контейнер
FROM alpine:3.21.3 AS final
WORKDIR /app/

# Добавляем необходимые зависимости
# tzdata - для установки временной зоны
# curl - для проверки доступности сервиса
RUN apk add --no-cache tzdata curl

# Копируем бинарники, миграции и конфиги из builder-образа
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/seeds ./seeds
COPY --from=builder /app/SERVER .
COPY --from=builder /app/MIGRATOR .
COPY --from=builder /app/PARSER .
COPY --from=builder /app/SEEDER .
COPY --from=builder /app/CLEARDB .

CMD ["sh", "-c", "./MIGRATOR -typeTask up -dsn $DSN && ./SEEDER -typeTask up -dsn $DSN && exec ./SERVER"]