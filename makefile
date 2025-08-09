test:
	go test -v -count=1 ./tests/...

lint:
	golangci-lint run ./...

bot:
	go run cmd/bot/main.go -configEnv ./.env

updater:
	go run cmd/kofd_updater/main.go -configEnv "./.env" -firstDate "2025-07-01" -lastDate "2025-07-26" -bin "800727301256" 

updater30:
	go run cmd/kofd_updater/main.go -configEnv "./.env" -days 10 -bin "800727301256" 

updater10:
	go run cmd/kofd_updater/main.go -configEnv "./.env" -days 10 -bin "800727301256" 

updater1:
	go run cmd/kofd_updater/main.go -configEnv "./.env" -days 1 -bin "800727301256" 

migrate_up:
	go run cmd/migrator/main.go -typeTask "up" -dsn "postgres://postgres:postgres@localhost:5911/go_cipo_bot?sslmode=disable"