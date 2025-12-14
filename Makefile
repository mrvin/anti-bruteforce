build:
	go build -ldflags '-w -s' -o bin/anti-bruteforce
test:
	go test -cover -v -race -count=10 ./internal/ratelimiting/leakybucket/
build-anti-bruteforce:
	docker compose -f deployments/docker-compose.yaml --env-file deployments/postgres.env build anti-bruteforce
up:
	docker compose -f deployments/docker-compose.yaml --env-file deployments/postgres.env up
down:
	docker compose -f deployments/docker-compose.yaml --env-file deployments/postgres.env down
lint:
	golangci-lint run ./...
codegen:
	go generate ./...
.PHONY: build test build-anti-bruteforce up down lint codegen
