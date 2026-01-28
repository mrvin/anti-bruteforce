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
	mkdir -p reports
	go test -race -count=10 ./internal/ratelimiting/fixedwindow/ -coverprofile=reports/coverage.out
coverage:
	go tool cover -func reports/coverage.out | grep "total:" | \
	awk '{print ((int($$3) > 86) != 1) }'
report:
	go tool cover -html=reports/coverage.out -o reports/cover.html
run:
	docker compose -f deployments/docker-compose.yaml --env-file configs/anti-bruteforce.env --profile prod up --build
down:
	docker compose -f deployments/docker-compose.yaml --env-file configs/anti-bruteforce.env --profile prod down
.PHONY: test run down
