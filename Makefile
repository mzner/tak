.PHONY: build install dev test test-integration lint test-all clean help

build:
	go build -o bin/tak .

install:
	go install .

dev:
	go run . $(ARGS)

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

lint:
	golangci-lint run ./...

test-all: lint test

clean:
	rm -rf bin/
	go clean

help:
	@echo "Available targets:"
	@echo "  build              - Build the tak binary"
	@echo "  install            - Install tak binary"
	@echo "  dev                - Run tak in dev mode"
	@echo "  test               - Run unit tests"
	@echo "  test-integration   - Run integration tests"
	@echo "  lint               - Run linters"
	@echo "  test-all           - Run lint and tests"
	@echo "  clean              - Remove build artifacts"
	@echo "  help               - Show this help message"
