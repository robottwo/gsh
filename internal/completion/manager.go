package completion

import (
	"context"
	"fmt"
	"strings"

	"mvdan.cc/sh/v3/interp"
)

// CompletionType represents the type of completion
type CompletionType string

const (
	// WordListCompletion represents word list based completion (-W option)
	WordListCompletion CompletionType = "W"
	// FunctionCompletion represents function based completion (-F option)
	FunctionCompletion CompletionType = "F"
)

// CompletionSpec represents a completion specification for a command
type CompletionSpec struct {
	Command string
	Type    CompletionType
	Value   string   // function name or wordlist
	Options []string // additional options like -o dirname
}

// CompletionManager manages command completion specifications
type CompletionManager struct {
	specs map[string]CompletionSpec
}

// NewCompletionManager creates a new CompletionManager
func NewCompletionManager() *CompletionManager {
	return &CompletionManager{
		specs: make(map[string]CompletionSpec),
	}
}

// AddSpec adds or updates a completion specification
func (m *CompletionManager) AddSpec(spec CompletionSpec) {
	m.specs[spec.Command] = spec
}

// RemoveSpec removes a completion specification
func (m *CompletionManager) RemoveSpec(command string) {
	delete(m.specs, command)
}

// GetSpec retrieves a completion specification
func (m *CompletionManager) GetSpec(command string) (CompletionSpec, bool) {
	spec, ok := m.specs[command]
	return spec, ok
}

// ListSpecs returns all completion specifications
func (m *CompletionManager) ListSpecs() []CompletionSpec {
	specs := make([]CompletionSpec, 0, len(m.specs))
	for _, spec := range m.specs {
		specs = append(specs, spec)
	}
	return specs
}

// ExecuteCompletion executes a completion specification for a given command line
// and returns the list of possible completions
func (m *CompletionManager) ExecuteCompletion(ctx context.Context, runner *interp.Runner, spec CompletionSpec, args []string) ([]string, error) {
	switch spec.Type {
	case WordListCompletion:
		words := strings.Fields(spec.Value)
		completions := make([]string, 0)
		word := ""
		if len(args) > 0 {
			word = args[len(args)-1]
		}
		for _, w := range words {
			if word == "" || strings.HasPrefix(w, word) {
				completions = append(completions, w)
			}
		}
		return completions, nil

	case FunctionCompletion:
		fn := NewCompletionFunction(spec.Value, runner)
		return fn.Execute(ctx, args)

	default:
		return nil, fmt.Errorf("unsupported completion type: %s", spec.Type)
	}
}
