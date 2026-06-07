package cmd

import "runtime"

func tmuxInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install with: brew install tmux"
	case "linux":
		return "Install with: apt install tmux (Debian/Ubuntu) or dnf install tmux (Fedora)"
	default:
		return "Install tmux to use this feature"
	}
}
