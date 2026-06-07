# Getting Started

## Prerequisites

- Git 2.20+ (for worktree support)
- tmux (optional, for `tak open` and `-o` flag)
- Go 1.22+ (only if building from source)

## Installation

### Homebrew (macOS/Linux)

```bash
brew install mzner/tap/tak
```

### From source

```bash
go install github.com/mzner/tak@latest
```

### Build locally

```bash
git clone https://github.com/mzner/tak.git
cd tak
make build
# Binary at ./bin/tak
```

## First Steps

### 1. Initialize tak in your repo

```bash
cd ~/projects/my-repo
tak init
```

This creates:
- `.tak.yml` — your config file
- `.tak/` — state directory (auto-added to .gitignore)

### 2. Set up shell integration

Add to your `~/.zshrc` (or `~/.bashrc`):

```bash
eval "$(tak shell-init zsh)"
```

Reload your shell: `source ~/.zshrc`

### 3. Create your first worktree

```bash
tak add feature/my-feature -o
```

This:
1. Creates a worktree at `../my-repo--feature--my-feature`
2. Opens a tmux window named `feature-my-feature`
3. You're now working in the worktree

### 4. Navigate between worktrees

```bash
tak ls          # see all worktrees
tak cd feature/my-feature   # jump to it
tak open feature/my-feature # switch tmux window
```

### 5. Pin important worktrees

```bash
tak pin  # pins the current worktree
```

Pinned worktrees are never removed by `tak gc`.

### 6. Clean up when done

```bash
tak doctor          # see what's stale
tak gc --dry-run    # preview cleanup
tak gc              # actually clean up
```
