package completion

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
	Value   string        // function name or wordlist
	Options []string      // additional options like -o dirname
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

