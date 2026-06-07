// Package runner provides command execution abstractions for the tak CLI.
// It defines the CommandRunner interface which allows executing system commands
// with support for working directory changes. This abstraction is critical for
// testability - FakeRunner allows tests to mock command execution without
// actually running external processes, while ExecRunner provides the real
// implementation using os/exec.
package runner
