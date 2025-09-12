package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGenerateCommandPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected []string
	}{
		{
			name:     "simple command",
			command:  "ls",
			expected: []string{"ls"},
		},
		{
			name:     "command with single flag",
			command:  "ls -la",
			expected: []string{"ls", "ls -la"},
		},
		{
			name:     "command with multiple flags and arguments",
			command:  "ls --foo bar baz",
			expected: []string{"ls", "ls --foo", "ls --foo bar baz"},
		},
		{
			name:     "git command",
			command:  "git commit -m message",
			expected: []string{"git", "git commit", "git commit -m message"},
		},
		{
			name:     "empty command",
			command:  "",
			expected: []string{},
		},
		{
			name:     "command with extra spaces",
			command:  "  ls   -la   ",
			expected: []string{"ls", "ls -la"},
		},
		{
			name:     "command with quoted arguments",
			command:  "awk 'NR==1 {print \"=== ADVANCED FILE LISTING ===\"; print \"test\"}'",
			expected: []string{"awk", "awk 'NR==1 {print \"=== ADVANCED FILE LISTING ===\"; print \"test\"}'"},
		},
		{
			name:     "command with single quoted argument",
			command:  "sed 's/\\x1b\\[[0-9;]*m//g'",
			expected: []string{"sed", "sed 's/\\x1b\\[[0-9;]*m//g'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateCommandPrefixes(tt.command)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShowPermissionsMenuCompoundCommands(t *testing.T) {
	// Test compound command with pipe
	command := "ls -la | grep txt"

	// We can't easily test the interactive menu, but we can test that it doesn't crash
	// and that the compound command parsing works by checking the atoms are created correctly
	individualCommands, err := ExtractCommands(command)
	assert.NoError(t, err)
	assert.Equal(t, []string{"ls -la", "grep txt"}, individualCommands)

	// Generate atoms for each individual command
	var atoms []PermissionAtom
	for _, cmd := range individualCommands {
		prefixes := GenerateCommandPrefixes(cmd)
		for _, prefix := range prefixes {
			atoms = append(atoms, PermissionAtom{
				Command: prefix,
				Enabled: false,
				IsNew:   true,
			})
		}
	}

	// Should have atoms for both ls and grep commands
	assert.Len(t, atoms, 4) // ["ls", "ls -la", "grep", "grep txt"]
	assert.Equal(t, "ls", atoms[0].Command)
	assert.Equal(t, "ls -la", atoms[1].Command)
	assert.Equal(t, "grep", atoms[2].Command)
	assert.Equal(t, "grep txt", atoms[3].Command)
}

func TestPermissionAtom(t *testing.T) {
	atom := PermissionAtom{
		Command: "ls -la",
		Enabled: true,
		IsNew:   false,
	}

	assert.Equal(t, "ls -la", atom.Command)
	assert.True(t, atom.Enabled)
	assert.False(t, atom.IsNew)
}

func TestPermissionsMenuState(t *testing.T) {
	atoms := []PermissionAtom{
		{Command: "ls", Enabled: false, IsNew: true},
		{Command: "ls -la", Enabled: true, IsNew: true},
	}

	state := &PermissionsMenuState{
		atoms:           atoms,
		selectedIndex:   1,
		originalCommand: "ls -la",
		active:          true,
	}

	assert.Len(t, state.atoms, 2)
	assert.Equal(t, 1, state.selectedIndex)
	assert.Equal(t, "ls -la", state.originalCommand)
	assert.True(t, state.active)
}

func TestRenderPermissionsMenu(t *testing.T) {
	atoms := []PermissionAtom{
		{Command: "ls", Enabled: false, IsNew: true},
		{Command: "ls -la", Enabled: true, IsNew: true},
	}

	state := &PermissionsMenuState{
		atoms:           atoms,
		selectedIndex:   0,
		originalCommand: "ls -la",
		active:          true,
	}

	result := renderPermissionsMenu(state)

	// Check that the menu contains expected elements
	assert.Contains(t, result, "Permission Management")
	assert.Contains(t, result, "ls")
	assert.Contains(t, result, "ls -la")
	assert.Contains(t, result, ">")        // Selection indicator
	assert.Contains(t, result, "[ ]")      // Unchecked box
	assert.Contains(t, result, "[✓]")      // Checked box
	assert.Contains(t, result, "Navigate") // Instructions
}

func TestHandleMenuInput(t *testing.T) {
	atoms := []PermissionAtom{
		{Command: "ls", Enabled: false, IsNew: true},
		{Command: "ls -la", Enabled: false, IsNew: true},
	}

	state := &PermissionsMenuState{
		atoms:           atoms,
		selectedIndex:   0,
		originalCommand: "ls -la",
		active:          true,
	}

	// Test navigation down
	result := handleMenuInput(state, "j")
	assert.Equal(t, "", result) // Should continue
	assert.Equal(t, 1, state.selectedIndex)

	// Test navigation up
	result = handleMenuInput(state, "k")
	assert.Equal(t, "", result) // Should continue
	assert.Equal(t, 0, state.selectedIndex)

	// Test toggle
	assert.False(t, state.atoms[0].Enabled)
	result = handleMenuInput(state, " ")
	assert.Equal(t, "", result) // Should continue
	assert.True(t, state.atoms[0].Enabled)

	// Test direct yes
	result = handleMenuInput(state, "y")
	assert.Equal(t, "y", result)

	// Test direct no
	result = handleMenuInput(state, "n")
	assert.Equal(t, "n", result)

	// Test escape
	state.active = true
	result = handleMenuInput(state, "escape")
	assert.Equal(t, "n", result)
	assert.False(t, state.active)
}

func TestProcessMenuSelection(t *testing.T) {
	// Test with no enabled permissions
	atoms := []PermissionAtom{
		{Command: "ls", Enabled: false, IsNew: true},
		{Command: "ls -la", Enabled: false, IsNew: true},
	}

	state := &PermissionsMenuState{
		atoms:           atoms,
		selectedIndex:   0,
		originalCommand: "ls -la",
		active:          true,
	}

	result := processMenuSelection(state)
	assert.Equal(t, "y", result) // Should return "y" for one-time execution
	assert.False(t, state.active)

	// Test with enabled permissions
	atoms[0].Enabled = true
	state.active = true
	state.atoms = atoms

	result = processMenuSelection(state)
	assert.Equal(t, "manage", result) // Should return "manage" for permission management
	assert.False(t, state.active)
}

func TestGetEnabledPermissions(t *testing.T) {
	atoms := []PermissionAtom{
		{Command: "ls", Enabled: true, IsNew: true},
		{Command: "ls -la", Enabled: false, IsNew: true},
		{Command: "ls -la /tmp", Enabled: true, IsNew: true},
	}

	state := &PermissionsMenuState{
		atoms:           atoms,
		selectedIndex:   0,
		originalCommand: "ls -la /tmp",
		active:          true,
	}

	enabled := GetEnabledPermissions(state)
	assert.Len(t, enabled, 2)
	assert.Equal(t, "ls", enabled[0].Command)
	assert.Equal(t, "ls -la /tmp", enabled[1].Command)
}

// Mock test for ShowPermissionsMenu - this would require more complex mocking
// of the gline.Gline function, so we'll test the individual components instead
func TestShowPermissionsMenuComponents(t *testing.T) {
	logger := zap.NewNop()

	// Test that we can generate prefixes correctly
	prefixes := GenerateCommandPrefixes("ls --foo bar")
	assert.Equal(t, []string{"ls", "ls --foo", "ls --foo bar"}, prefixes)

	// Test that we can create atoms correctly
	atoms := make([]PermissionAtom, len(prefixes))
	for i, prefix := range prefixes {
		atoms[i] = PermissionAtom{
			Command: prefix,
			Enabled: false,
			IsNew:   true,
		}
	}

	assert.Len(t, atoms, 3)
	assert.Equal(t, "ls", atoms[0].Command)
	assert.Equal(t, "ls --foo", atoms[1].Command)
	assert.Equal(t, "ls --foo bar", atoms[2].Command)

	// Test state initialization
	state := &PermissionsMenuState{
		atoms:           atoms,
		selectedIndex:   0,
		originalCommand: "ls --foo bar",
		active:          true,
	}

	assert.True(t, state.active)
	assert.Equal(t, 0, state.selectedIndex)
	assert.Equal(t, "ls --foo bar", state.originalCommand)

	// Suppress unused variable warning
	_ = logger
}
