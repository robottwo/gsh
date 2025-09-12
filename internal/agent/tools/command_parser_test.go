package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple command",
			input:    "ls -la",
			expected: []string{"ls -la"},
		},
		{
			name:     "semicolon separated commands",
			input:    "ls; pwd; echo hello",
			expected: []string{"ls", "pwd", "echo hello"},
		},
		{
			name:     "AND operator commands",
			input:    "ls && pwd && echo success",
			expected: []string{"ls", "pwd", "echo success"},
		},
		{
			name:     "OR operator commands",
			input:    "test -f file || echo 'file not found'",
			expected: []string{"test -f file", "echo 'file not found'"},
		},
		{
			name:     "pipe commands",
			input:    "ls -la | grep txt | sort",
			expected: []string{"ls -la", "grep txt", "sort"},
		},
		{
			name:     "mixed operators",
			input:    "ls && pwd; echo done || echo failed",
			expected: []string{"ls", "pwd", "echo done", "echo failed"},
		},
		{
			name:     "subshell commands",
			input:    "(cd /tmp && ls)",
			expected: []string{"cd /tmp", "ls"},
		},
		{
			name:     "nested subshells",
			input:    "(cd /tmp && (ls -la && pwd))",
			expected: []string{"cd /tmp", "ls -la", "pwd"},
		},
		{
			name:     "command substitution",
			input:    "echo $(date)",
			expected: []string{"echo $(date)", "date"},
		},
		{
			name:     "complex command substitution",
			input:    "echo $(ls | head -1) && pwd",
			expected: []string{"echo $(ls)", "pwd", "ls", "head -1"},
		},
		{
			name:     "malicious command injection",
			input:    "ls; rm -rf /",
			expected: []string{"ls", "rm -rf /"},
		},
		{
			name:     "complex malicious injection",
			input:    "ls && (echo safe && rm -rf /tmp/*)",
			expected: []string{"ls", "echo safe", "rm -rf /tmp/*"},
		},
		{
			name:     "quoted arguments",
			input:    "echo 'hello world' && ls \"my file.txt\"",
			expected: []string{"echo 'hello world'", "ls \"my file.txt\""},
		},
		{
			name:     "empty command",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			input:    "   \t\n  ",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractCommands(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractCommandsErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid syntax - unmatched parenthesis",
			input: "ls (",
		},
		{
			name:  "invalid syntax - unmatched quote",
			input: "echo 'unclosed quote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractCommands(tt.input)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestValidateCompoundCommand(t *testing.T) {
	patterns := []string{
		"^ls.*",
		"^pwd.*",
		"^echo.*",
		"^git status.*",
	}

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "single approved command",
			command:  "ls -la",
			expected: true,
		},
		{
			name:     "multiple approved commands",
			command:  "ls && pwd && echo done",
			expected: true,
		},
		{
			name:     "single unapproved command",
			command:  "rm -rf /",
			expected: false,
		},
		{
			name:     "mixed approved and unapproved",
			command:  "ls && rm -rf /",
			expected: false,
		},
		{
			name:     "malicious injection attempt",
			command:  "ls; rm -rf /",
			expected: false,
		},
		{
			name:     "approved pipe commands",
			command:  "ls | grep txt",
			expected: false, // grep is not in approved patterns
		},
		{
			name:     "approved subshell commands",
			command:  "(ls && pwd)",
			expected: true,
		},
		{
			name:     "unapproved subshell commands",
			command:  "(ls && rm file)",
			expected: false,
		},
		{
			name:     "command substitution with approved commands",
			command:  "echo $(pwd)",
			expected: true,
		},
		{
			name:     "command substitution with unapproved commands",
			command:  "echo $(rm file)",
			expected: false,
		},
		{
			name:     "empty command",
			command:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateCompoundCommand(tt.command, patterns)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateCompoundCommandErrors(t *testing.T) {
	patterns := []string{"^ls.*"}

	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "invalid syntax",
			command: "ls (",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateCompoundCommand(tt.command, patterns)
			assert.Error(t, err)
			assert.False(t, result)
		})
	}
}

func TestGenerateCompoundCommandRegex(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected []string
	}{
		{
			name:     "single command",
			command:  "ls -la",
			expected: []string{"^ls.*"},
		},
		{
			name:     "multiple commands",
			command:  "ls && pwd && echo hello",
			expected: []string{"^ls.*", "^pwd.*", "^echo.*"},
		},
		{
			name:     "git commands with subcommands",
			command:  "git status && git commit -m 'test'",
			expected: []string{"^git status.*", "^git commit.*"},
		},
		{
			name:     "pipe commands",
			command:  "ls | grep txt | sort",
			expected: []string{"^ls.*", "^grep.*", "^sort.*"},
		},
		{
			name:     "subshell commands",
			command:  "(cd /tmp && ls)",
			expected: []string{"^cd.*", "^ls.*"},
		},
		{
			name:     "command substitution",
			command:  "echo $(date)",
			expected: []string{"^echo.*", "^date.*"},
		},
		{
			name:     "duplicate commands",
			command:  "ls && ls -la && ls",
			expected: []string{"^ls.*"}, // Should deduplicate
		},
		{
			name:     "empty command",
			command:  "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateCompoundCommandRegex(tt.command)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateCompoundCommandRegexErrors(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "invalid syntax",
			command: "ls (",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateCompoundCommandRegex(tt.command)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestSecurityScenarios(t *testing.T) {
	// Test various security attack scenarios
	patterns := []string{"^ls.*", "^echo.*"}

	securityTests := []struct {
		name     string
		command  string
		expected bool
		reason   string
	}{
		{
			name:     "basic injection with semicolon",
			command:  "ls; rm -rf /",
			expected: false,
			reason:   "rm command should not be approved",
		},
		{
			name:     "injection with AND operator",
			command:  "ls && rm -rf /",
			expected: false,
			reason:   "rm command should not be approved",
		},
		{
			name:     "injection with OR operator",
			command:  "ls || rm -rf /",
			expected: false,
			reason:   "rm command should not be approved",
		},
		{
			name:     "injection in subshell",
			command:  "(ls && rm -rf /)",
			expected: false,
			reason:   "rm command in subshell should not be approved",
		},
		{
			name:     "injection in command substitution",
			command:  "echo $(ls && rm -rf /)",
			expected: false,
			reason:   "rm command in substitution should not be approved",
		},
		{
			name:     "nested injection",
			command:  "ls && (echo safe && rm -rf /)",
			expected: false,
			reason:   "nested rm command should not be approved",
		},
		{
			name:     "pipe injection",
			command:  "ls | rm -rf /",
			expected: false,
			reason:   "rm command in pipe should not be approved",
		},
		{
			name:     "complex nested injection",
			command:  "ls && (echo $(pwd && rm file) && ls)",
			expected: false,
			reason:   "rm command in nested substitution should not be approved",
		},
		{
			name:     "legitimate complex command",
			command:  "ls && echo $(ls -la) && (echo done && ls)",
			expected: true,
			reason:   "all commands should be approved",
		},
	}

	for _, tt := range securityTests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateCompoundCommand(tt.command, patterns)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}
