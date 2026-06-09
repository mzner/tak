.PHONY: build install dev test test-integration lint test-all ci setup clean help

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

# Mirror the GitHub Actions pipeline exactly so a green `make ci` means a
# green CI run. The pre-push hook (see `make setup`) runs this before pushing.
ci:
	go build ./...
	go test ./...
	go test -tags=integration -count=1 ./...
	golangci-lint run ./...

# Point git at the version-controlled .githooks/ directory so the pre-push
# hook is active. Run once after cloning.
setup:
	git config core.hooksPath .githooks
	@echo "Git hooks enabled (.githooks). Pushes now run 'make ci' first."

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
	@echo "  ci                 - Run the full CI pipeline locally"
	@echo "  setup              - Enable git pre-push hook (run once after clone)"
	@echo "  clean              - Remove build artifacts"
	@echo "  help               - Show this help message"
