APP        ?= go-web-template
BINARY     ?= go-web-template

.PHONY: run build test lint tidy swagger clean docker-up docker-down docker-build

run: ## run the API (dev config by default)
	go run ./cmd/api

build: ## build the API binary into bin/
	go build -o bin/$(BINARY) ./cmd/api

test: ## run unit tests
	go test ./... -race

lint: ## static checks
	go vet ./...
	gofmt -l .

tidy: ## go mod tidy
	go mod tidy

swagger: ## regenerate OpenAPI docs into docs/
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs

docker-up: ## start mysql + redis (+ app when configured)
	docker compose up -d

docker-down: ## stop the local stack
	docker compose down

docker-build: ## build the application image
	docker build -t $(APP):latest .

clean:
	rm -rf bin logs
