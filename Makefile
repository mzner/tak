.PHONY: build install dev test test-integration lint test-all ci setup clean release help

# Resolve golangci-lint via PATH, falling back to GOPATH/bin. Git hooks run
# with a minimal environment that often lacks GOPATH/bin on PATH, so the bare
# command name isn't enough.
GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null || echo "$(shell go env GOPATH)/bin/golangci-lint")

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
	@test -x "$(GOLANGCI_LINT)" || { echo "golangci-lint not found. Install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; exit 1; }
	"$(GOLANGCI_LINT)" run ./...

test-all: lint test

# Mirror the GitHub Actions pipeline exactly so a green `make ci` means a
# green CI run. The pre-push hook (see `make setup`) runs this before pushing.
ci:
	@test -x "$(GOLANGCI_LINT)" || { echo "golangci-lint not found. Install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; exit 1; }
	go build ./...
	go test ./...
	go test -tags=integration -count=1 ./...
	"$(GOLANGCI_LINT)" run ./...

# Point git at the version-controlled .githooks/ directory so the pre-push
# hook is active. Run once after cloning.
setup:
	git config core.hooksPath .githooks
	@echo "Git hooks enabled (.githooks). Pushes now run 'make ci' first."

# Create a release: runs CI, tags, pushes to all remotes. Usage:
#   make release VERSION=0.2.0
#   make release              (auto-bumps patch from last tag)
release:
	@if [ -n "$$(git status --porcelain)" ]; then echo "error: working tree is dirty"; exit 1; fi
	@$(MAKE) ci
	$(eval LAST_TAG := $(shell git tag -l 'v*' | sort -V | tail -1))
	$(eval NEW_TAG := $(if $(VERSION),v$(VERSION),$(shell \
		if [ -z "$(LAST_TAG)" ]; then echo "v0.1.0"; \
		else echo "$(LAST_TAG)" | awk -F. '{printf "%s.%s.%d", $$1, $$2, $$3+1}'; \
		fi)))
	@echo "Releasing $(NEW_TAG) (previous: $(or $(LAST_TAG),none))"
	@read -p "Continue? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	git tag -a $(NEW_TAG) -m "Release $(NEW_TAG)"
	git push origin main --tags

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
	@echo "  release            - Run CI, tag, and push (VERSION=x.y.z or auto-bump)"
	@echo "  clean              - Remove build artifacts"
	@echo "  help               - Show this help message"
