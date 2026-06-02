APP_NAME=taskboard-api

DB_URL=postgres://postgres:postgres@localhost:5432/taskboard?sslmode=disable

run:
	go run ./cmd/server

build:
	go build -o bin/$(APP_NAME) ./cmd/server

test:
	go test ./...

test-unit:
	go test ./internal/service

test-integration:
	go test -tags=integration ./internal/app

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

generate:
	oapi-codegen -generate types -package api -o internal/api/openapi.gen.go api/openapi.yaml

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

fmt:
	go fmt ./...

lint:
	go vet ./...

clean:
	rm -rf bin