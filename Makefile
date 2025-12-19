lint:
	golangci-lint run ./...
check-format:
	test -z $$(go fmt ./...)
codegen:
	go generate ./...
LDFLAGS := -w -s
build:
	go build -ldflags "$(LDFLAGS)" -o bin/anti-bruteforce cmd/anti-bruteforce/main.go
build-ab-admin:
	go build -ldflags "$(LDFLAGS)" -o bin/ab-admin cmd/ab-admin/main.go
.PHONY: lint check-format codegen build build-ab-admin

test:
	go test -cover -v -race -count=10 ./internal/ratelimiting/leakybucket/
up:
	docker compose -f deployments/docker-compose.yaml --env-file deployments/postgres.env --profile prod up --build
down:
	docker compose -f deployments/docker-compose.yaml --env-file deployments/postgres.env --profile prod down
.PHONY: test up down
