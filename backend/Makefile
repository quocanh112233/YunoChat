DB_DSN=postgres://postgres:123456@localhost:5432/yunochat?sslmode=disable

migrate-up:
	migrate -path db/migrations -database "$(DB_DSN)" -verbose up

migrate-down:
	migrate -path db/migrations -database "$(DB_DSN)" -verbose down

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go
