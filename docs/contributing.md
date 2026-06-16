# Contributing to tak

## Development Setup

```bash
git clone https://github.com/mzner/tak.git
cd tak
go mod download
make build
make setup   # enable the pre-push hook (runs `make ci` before every push)
```

### Requirements

- Go 1.22+
- Git 2.20+
- tmux (for testing tmux features)
- golangci-lint (for `make lint`)

## Running Tests

```bash
make test              # unit tests (fast, no git/tmux needed)
make test-integration  # integration tests (needs git)
make lint              # linter
make test-all          # everything
make ci                # the exact GitHub Actions pipeline (build + test + integration + lint)
```

`make ci` mirrors `.github/workflows/ci.yml` step for step, so a green
`make ci` means a green CI run. The pre-push hook runs it automatically;
bypass in an emergency with `git push --no-verify`.

## Project Structure

```
tak/
├── main.go           # Entry point
├── cmd/              # Cobra CLI commands (thin wiring, no business logic)
├── internal/         # Domain packages (where the real logic lives)
│   ├── runner/       # CommandRunner interface for shelling out
│   ├── paths/        # Worktree path resolution + slugification
│   ├── config/       # YAML config loading and merging
│   ├── state/        # .tak/state.json cache management
│   ├── worktree/     # Git worktree operations
│   ├── tmux/         # Tmux window management
│   ├── shell/        # Shell hook generation
│   ├── hooks/        # Lifecycle hooks (copy, symlink, command)
│   └── doctor/       # Health checks
└── testdata/         # Test fixtures
```

### Key Design Principles

- **cmd/ is thin**: parse flags → call internal/ → format output. No logic.
- **internal/ packages are independent**: each has one job, tested in isolation.
- **CommandRunner for testability**: packages don't call os/exec directly. They accept a runner interface. Tests use FakeRunner.
- **Shell out, don't embed**: we call `git` and `tmux` binaries rather than using Go libraries. Simpler, debuggable, fewer deps.

## Adding a New Command

1. Create `cmd/mycommand.go`
2. Define a `cobra.Command` that parses flags and calls into `internal/` packages
3. Register it in `init()` with `rootCmd.AddCommand(myCmd)`
4. Add `ValidArgsFunction = completeWorktreeBranches` if it takes a branch argument
5. Add any new domain logic to the appropriate `internal/` package (or create a new one)
6. Write tests for the domain logic using FakeRunner
7. Add an integration test to `test_integration_test.go`
8. Update README command table and docs

## Adding a New Internal Package

1. Create `internal/mypkg/`
2. Add `doc.go` with package-level documentation
3. Implement the package with a service struct accepting `runner.CommandRunner` if it shells out
4. Write `*_test.go` using FakeRunner for any command execution

## Code Style

- Go standard formatting (`gofmt`)
- No comments unless the "why" is non-obvious
- Every exported function and type needs a doc comment
- Package-level `doc.go` in every `internal/` package
- Error messages: lowercase, no period, actionable
- Conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`

## Releasing

```bash
make release               # auto-bumps patch (v0.1.0 → v0.1.1)
make release VERSION=1.0.0 # explicit version
```

This runs CI, creates a git tag, and pushes to GitHub. CI then builds release binaries and publishes them automatically.

## Pull Requests

- One focused change per PR
- Tests required for new functionality
- Update docs if user-facing behavior changes
- Keep commits atomic (one logical change per commit)
