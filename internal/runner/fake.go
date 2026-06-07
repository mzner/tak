package runner

import (
	"strings"
)

// Call represents a recorded command execution.
type Call struct {
	Dir  string
	Name string
	Args []string
}

// String returns a string representation of the Call suitable for matching.
// Format: "name arg1 arg2 ..." if dir is empty, otherwise "dir|name arg1 arg2 ..."
func (c *Call) String() string {
	cmdStr := c.Name
	if len(c.Args) > 0 {
		cmdStr = cmdStr + " " + strings.Join(c.Args, " ")
	}
	if c.Dir == "" {
		return cmdStr
	}
	return c.Dir + "|" + cmdStr
}

// Response represents a preset response for a command.
type Response struct {
	Error error
}

// FakeRunner is a CommandRunner implementation that records calls and returns preset responses.
type FakeRunner struct {
	Calls     []*Call
	Responses map[string]*Response
}

// NewFakeRunner creates a new FakeRunner with empty responses.
func NewFakeRunner() *FakeRunner {
	return &FakeRunner{
		Calls:     []*Call{},
		Responses: make(map[string]*Response),
	}
}

// SetResponse sets a preset response for a command key.
// The key can be a full command (with or without dir prefix) or a substring.
func (f *FakeRunner) SetResponse(key string, response *Response) {
	f.Responses[key] = response
}

// Run records a call and returns a preset response if available.
func (f *FakeRunner) Run(name string, args ...string) error {
	call := &Call{Dir: "", Name: name, Args: args}
	f.Calls = append(f.Calls, call)
	return f.lookupResponse(call)
}

// RunInDir records a call with directory and returns a preset response if available.
func (f *FakeRunner) RunInDir(dir, name string, args ...string) error {
	call := &Call{Dir: dir, Name: name, Args: args}
	f.Calls = append(f.Calls, call)
	return f.lookupResponse(call)
}

// lookupResponse returns a preset response by progressively trying longer key prefixes.
// It tries the exact call string first, then progressively shorter variations.
func (f *FakeRunner) lookupResponse(call *Call) error {
	callStr := call.String()

	// Try exact match first
	if resp, ok := f.Responses[callStr]; ok {
		return resp.Error
	}

	// Try just the command name
	if resp, ok := f.Responses[call.Name]; ok {
		return resp.Error
	}

	// Try with dir prefix if present
	if call.Dir != "" {
		if resp, ok := f.Responses[call.Dir]; ok {
			return resp.Error
		}
	}

	// Return nil if no response is configured
	return nil
}
