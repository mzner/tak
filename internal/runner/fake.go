package runner

import "fmt"

// Call records a single command invocation made through FakeRunner.
type Call struct {
	Dir  string
	Name string
	Args []string
}

// Response holds a preset output and optional error for a command.
type Response struct {
	Output []byte
	Err    error
}

// FakeRunner implements CommandRunner for testing.
// It records all calls and returns preset responses.
type FakeRunner struct {
	// Calls records every command invocation in order.
	Calls []Call

	// Responses maps a command key to the output it should return.
	// The key format is "name arg1 arg2..." (space-separated).
	// If a command is not found in Responses, it returns empty output.
	Responses map[string]Response
}

// NewFakeRunner creates a FakeRunner with the given preset responses.
func NewFakeRunner(responses map[string]Response) *FakeRunner {
	if responses == nil {
		responses = make(map[string]Response)
	}
	return &FakeRunner{
		Responses: responses,
	}
}

// Run records the call and returns the preset response.
func (f *FakeRunner) Run(name string, args ...string) ([]byte, error) {
	f.Calls = append(f.Calls, Call{Name: name, Args: args})
	return f.lookup(name, args...)
}

// RunInDir records the call (including dir) and returns the preset response.
func (f *FakeRunner) RunInDir(dir string, name string, args ...string) ([]byte, error) {
	f.Calls = append(f.Calls, Call{Dir: dir, Name: name, Args: args})
	return f.lookup(name, args...)
}

// lookup finds a matching response. It tries progressively shorter key prefixes.
func (f *FakeRunner) lookup(name string, args ...string) ([]byte, error) {
	// Build full command key and try progressively longer prefixes
	key := name
	if resp, ok := f.Responses[key]; ok {
		return resp.Output, resp.Err
	}
	for _, arg := range args {
		key += " " + arg
		if resp, ok := f.Responses[key]; ok {
			return resp.Output, resp.Err
		}
	}

	return nil, nil
}

// CallCount returns how many times any command was invoked.
func (f *FakeRunner) CallCount() int {
	return len(f.Calls)
}

// LastCall returns the most recent call, or panics if none.
func (f *FakeRunner) LastCall() Call {
	if len(f.Calls) == 0 {
		panic("FakeRunner: no calls recorded")
	}
	return f.Calls[len(f.Calls)-1]
}

// String returns a readable summary of a Call (for debugging test failures).
func (c Call) String() string {
	s := c.Name
	for _, a := range c.Args {
		s += " " + a
	}
	if c.Dir != "" {
		s = fmt.Sprintf("[%s] %s", c.Dir, s)
	}
	return s
}
