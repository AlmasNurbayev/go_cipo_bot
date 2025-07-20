test:
	go test -v -count=1 ./tests/...

lint:
	golangci-lint run ./...

bot:
	go run cmd/bot/main.go -configEnv ./.env

updater:
	go run cmd/kofd_updater/main.go -configEnv "./.env" -firstDate "2025-06-01" -lastDate "2025-06-30" -bin "800727301256" 

migrate_up:
	go run cmd/migrator/main.go -typeTask "up" -dsn "postgres://postgres:postgres@localhost:5911/go_cipo_bot?sslmode=disable"