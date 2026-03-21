include .env
export

migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" -verbose up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" -verbose down

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)
