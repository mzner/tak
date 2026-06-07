package paths

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugifyBranch(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{name: "simple slash", branch: "feature/auth", want: "feature--auth"},
		{name: "multiple slashes", branch: "feature/auth/totp", want: "feature--auth--totp"},
		{name: "no slash", branch: "main", want: "main"},
		{name: "uppercase", branch: "Feature/Auth", want: "feature--auth"},
		{name: "spaces", branch: "my branch", want: "my-branch"},
		{name: "special chars", branch: "fix/bug#123", want: "fix--bug-123"},
		{name: "dots", branch: "release/v1.0.0", want: "release--v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SlugifyBranch(tt.branch)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTmuxSlug(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{name: "simple slash", branch: "feature/auth", want: "feature-auth"},
		{name: "multiple slashes", branch: "feature/auth/totp", want: "feature-auth-totp"},
		{name: "no slash", branch: "main", want: "main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TmuxSlug(tt.branch)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name         string
		branch       string
		repoRoot     string
		worktreeBase string
		want         string
	}{
		{
			name:         "default sibling dir",
			branch:       "feature/auth",
			repoRoot:     "/Users/dev/projects/web",
			worktreeBase: "",
			want:         "/Users/dev/projects/web--feature--auth",
		},
		{
			name:         "configured base",
			branch:       "feature/auth",
			repoRoot:     "/Users/dev/projects/web",
			worktreeBase: "/Users/dev/worktrees",
			want:         "/Users/dev/worktrees/web--feature--auth",
		},
		{
			name:         "simple branch default",
			branch:       "hotfix",
			repoRoot:     "/home/user/repos/ocis",
			worktreeBase: "",
			want:         "/home/user/repos/ocis--hotfix",
		},
		{
			name:         "simple branch configured base",
			branch:       "hotfix",
			repoRoot:     "/home/user/repos/ocis",
			worktreeBase: "/tmp/worktrees",
			want:         "/tmp/worktrees/ocis--hotfix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.branch, tt.repoRoot, tt.worktreeBase)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRepoName(t *testing.T) {
	assert.Equal(t, "web", RepoName("/Users/dev/projects/web"))
	assert.Equal(t, "ocis", RepoName("/home/user/repos/ocis"))
	assert.Equal(t, "tak", RepoName("/Users/dev/tak"))
}
