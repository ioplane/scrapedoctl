package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec" //nolint:depguard // exec is the core mechanism of this plugin provider.
	"time"
)

// defaultExecTimeout is the default timeout for exec plugin commands.
const defaultExecTimeout = 30 * time.Second

// ErrExecEmptyOutput is returned when the plugin command produces no stdout.
var ErrExecEmptyOutput = errors.New("exec provider: empty output from command")

// execRequest is the JSON payload sent to the plugin on stdin.
type execRequest struct {
	Query   string      `json:"query"`
	Engine  string      `json:"engine"`
	Options execOptions `json:"options"`
}

// execOptions mirrors Options for JSON serialization to the plugin.
type execOptions struct {
	Lang    string `json:"lang"`
	Country string `json:"country"`
	Limit   int    `json:"limit"`
	Page    int    `json:"page"`
	Raw     bool   `json:"raw"`
}

// ExecProvider implements Provider by executing an external binary.
// The binary receives a JSON request on stdin and writes a JSON Response to stdout.
type ExecProvider struct {
	name    string
	command string
	args    []string
	engines []string
	timeout time.Duration
}

// NewExecProvider creates an ExecProvider for the given command and engine list.
// The default timeout is 30 seconds; use WithTimeout to override.
func NewExecProvider(name, command string, engines []string) *ExecProvider {
	return &ExecProvider{
		name:    name,
		command: command,
		engines: engines,
		timeout: defaultExecTimeout,
	}
}

// WithTimeout returns the provider with an overridden timeout.
func (p *ExecProvider) WithTimeout(d time.Duration) *ExecProvider {
	p.timeout = d
	return p
}

// WithArgs returns the provider with extra command-line arguments.
func (p *ExecProvider) WithArgs(args ...string) *ExecProvider {
	p.args = args
	return p
}

// Name returns the provider's unique identifier.
func (p *ExecProvider) Name() string {
	return p.name
}

// Engines returns the list of search engines this provider supports.
func (p *ExecProvider) Engines() []string {
	return p.engines
}

// Search executes the external command, passing the query as JSON on stdin
// and parsing the JSON Response from stdout.
func (p *ExecProvider) Search(ctx context.Context, query string, opts Options) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	req := execRequest{
		Query:  query,
		Engine: opts.Engine,
		Options: execOptions{
			Lang:    opts.Lang,
			Country: opts.Country,
			Limit:   opts.Limit,
			Page:    opts.Page,
			Raw:     opts.Raw,
		},
	}

	input, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("exec provider %q: marshal request: %w", p.name, err)
	}

	//nolint:gosec // G204: command is configured by the admin, not user input.
	cmd := exec.CommandContext(ctx, p.command, p.args...)
	cmd.Cancel = func() error {
		return cmd.Process.Kill()
	}
	cmd.WaitDelay = 500 * time.Millisecond
	cmd.Stdin = bytes.NewReader(input)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	// stderr is intentionally not captured; plugins may use it for logging.

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("exec provider %q: run command: %w", p.name, err)
	}

	out := stdout.Bytes()
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrExecEmptyOutput, p.name)
	}

	var resp Response
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("exec provider %q: parse response: %w", p.name, err)
	}

	resp.Provider = p.name

	return &resp, nil
}
