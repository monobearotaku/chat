export POSTGRES_HOST = localhost
export POSTGRES_PORT = 5432
export POSTGRES_USER = some-handsome-man
export POSTGRES_PASSWORD = some-handsome-password
export POSTGRES_DB = chat

run:
	go run cmd/main/main.go

client:
	go run cmd/client/main.go