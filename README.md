# tak

> Dutch for "branch". Git worktree manager with pinning, tmux integration, and lifecycle tools.

tak makes git worktrees easy to create, navigate, and clean up. Pin long-lived worktrees, jump between them with tmux, and garbage-collect stale ones.

## Install

```bash
brew install mzner/tap/tak
```

Or build from source:

```bash
go install github.com/mzner/tak@latest
```

## Quick Start

```bash
# Initialize tak in your repo
tak init

# Create a worktree and open it in tmux
tak add feature/auth -t

# Pin it so gc won't clean it up
tak pin

# List all worktrees
tak ls

# Jump to a worktree
tak cd feature/auth

# Health check
tak doctor

# Clean up merged branches
tak gc --merged
```

## Shell Integration

Add to your shell rc file for `tak cd` to work:

```bash
# .zshrc or .bashrc
eval "$(tak shell-init zsh)"
```

```fish
# config.fish
tak shell-init fish | source
```

## Commands

| Command | Description |
|---------|-------------|
| `tak add <branch> [-t] [--pin]` | Create a worktree (`-t` opens tmux, `--pin` pins it) |
| `tak rm [branch...] [--force]` | Remove worktree(s) — interactive multi-select if no arg |
| `tak ls` | List all worktrees with status |
| `tak cd [branch]` | Change to a worktree directory — interactive if no arg |
| `tak open [branch]` | Open/switch to tmux window — interactive if no arg |
| `tak pin [branch]` | Pin a worktree (no arg = current) |
| `tak unpin [branch]` | Unpin a worktree |
| `tak doctor` | Health check all worktrees |
| `tak gc [--merged] [--dry-run]` | Clean up stale worktrees |
| `tak init` | Initialize tak in a repo |
| `tak shell-init <shell>` | Print shell hook |

## Configuration

### Per-repo: `.tak.yml`

```yaml
worktree_base: ""         # empty = sibling dirs (default)
branch_prefix: ""         # auto-prepend to branch names
pins:
  - feature/auth
```

### Global: `~/.config/tak/config.yml`

```yaml
worktree_base: ~/worktrees   # override default for all repos
repos:
  web: ~/projects/web
  ocis: ~/projects/ocis
```

## How It Works

- Worktrees are created as sibling directories by default: `~/projects/web` -> `~/projects/web--feature--auth`
- Pins are stored in `.tak.yml` (recoverable config, not ephemeral state)
- State cache (`.tak/state.json`) is rebuilt automatically if deleted
- All git/tmux interaction is via shell commands - no heavy dependencies

## Contributing

See [docs/contributing.md](docs/contributing.md).

## License

MIT
