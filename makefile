test:
	go test -v -count=1 ./cmd/... ./internal/...

lint:
	golangci-lint run ./cmd/... ./internal/...

bot:
	go run cmd/bot/main.go -configEnv ./.env

# Профиль заливки истории. Дефолты в .env рассчитаны на прогон по cron и для
# больших периодов малы: HTTP_LONG_TIMEOUT тратится на каждую активную кассу,
# поэтому потолок прогона нужен с запасом. Переменные окружения перекрывают .env
BACKFILL_ENV = POSTGRES_TIMEOUT=600s HTTP_LONG_TIMEOUT=120s HTTP_TIMEOUT=15s NATS_STARTUP_TIMEOUT=30s

updater:
	$(BACKFILL_ENV) go run cmd/kofd_updater/main.go -configEnv "./.env" -firstDate "2025-07-01" -lastDate "2025-07-26" -bin "800727301256"

updater2023-01:
	$(BACKFILL_ENV) go run cmd/kofd_updater/main.go -configEnv "./.env" -firstDate "2023-01-01" -lastDate "2023-01-31" -bin "800727301256"


updater32:
	go run cmd/kofd_updater/main.go -configEnv "./.env" -days 32 -bin "800727301256" 

updater10:
	go run cmd/kofd_updater/main.go -configEnv "./.env" -days 10 -bin "800727301256" 

updater1:
	go run cmd/kofd_updater/main.go -configEnv "./.env" -days 1 -bin "800727301256" 

migrate_up:
	go run cmd/migrator/main.go -typeTask "up" -dsn "postgres://postgres:postgres@localhost:5911/go_cipo_bot?sslmode=disable"
