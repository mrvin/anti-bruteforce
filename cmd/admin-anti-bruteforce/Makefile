build:
	go build -ldflags '-w -s' -o ../../bin/admin-anti-bruteforce
lint:
	golangci-lint run ./...
codegen:
	go generate ./...
.PHONY: build lint codegen
