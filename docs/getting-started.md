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

Add these lines to your `~/.zshrc` (or `~/.bashrc`):

```bash
eval "$(tak shell-init zsh)"
source <(tak completion zsh)
```

Without this, `tak cd` can only print paths (can't change your directory), and tab completion won't work.

Reload your shell: `source ~/.zshrc`

### 3. Create your first worktree

```bash
# New branch from main (default)
tak add feature/my-feature

# New branch from a specific base
tak add feature/my-feature -f develop

# Existing remote branch (checks it out, no new branch)
tak add feature/existing-branch

# Create and open in tmux
tak add feature/my-feature -o

# Create and pin (protect from gc)
tak add feature/my-feature -p
```

### 4. Navigate between worktrees

```bash
tak ls                          # see all worktrees
tak cd feature/my-feature       # jump to it
tak open feature/my-feature     # switch/create tmux window
tak info feature/my-feature     # show details (base, ahead/behind, age)
tak exec feature/my-feature -- git status  # run command without cd'ing
```

### 5. Pin important worktrees

```bash
tak pin                   # pin current worktree
tak pin feature/long-lived  # pin by name
```

Pinned worktrees are never removed by `tak gc`.

### 6. Clean up when done

```bash
tak rm feature/my-feature   # remove worktree + branch
tak doctor                  # see what's stale
tak gc -n                   # preview cleanup (dry run)
tak gc -m                   # clean up merged + broken worktrees
```

### 7. Register repos for cross-repo access

```bash
tak repo add                    # register current repo
tak repo add ~/projects/api     # register another repo
tak repo ls                     # see registered repos

# Then from anywhere:
tak ls web                      # list web's worktrees
tak cd web:feature/auth         # jump to web's worktree
```

## Optional: Tmux Layout

Configure pane layouts so `tak open` sets up your dev environment automatically:

```bash
tak layout   # interactive wizard
```

Or edit `.tak.yml` directly:

```yaml
tmux:
  layout: main-vertical
  panes:
    - name: editor
      command: $EDITOR
    - name: dev
      command: pnpm dev
    - name: shell
      command: ""
```

## Optional: Lifecycle Hooks

Automate setup for new worktrees (copy config, install deps):

```yaml
# .tak.yml
hooks:
  post_create:
    - type: copy
      from: .env
    - type: symlink
      from: node_modules
    - type: command
      command: pnpm install
```

Every `tak add` runs these automatically after creating the worktree.
