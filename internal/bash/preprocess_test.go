package bash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreprocessTypesetCommands_EmptyInput(t *testing.T) {
	result := PreprocessTypesetCommands("")
	assert.Equal(t, "", result, "empty input should return empty string")
}

func TestPreprocessTypesetCommands_NilLikeInput(t *testing.T) {
	// Test various edge cases that might be problematic
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"whitespace only", "   ", "   "},
		{"single space", " ", " "},
		{"newline only", "\n", "\n"},
		{"mixed whitespace", " \t\n ", " \t\n "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := PreprocessTypesetCommands(tc.input)
			assert.Equal(t, tc.expected, result, "input %q should return %q", tc.input, tc.expected)
		})
	}
}

func TestPreprocessTypesetCommands_NormalTransformation(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"typeset -f", "typeset -f", "gsh_typeset -f"},
		{"declare -f", "declare -f", "gsh_typeset -f"},
		{"typeset -F", "typeset -F", "gsh_typeset -F"},
		{"declare -F", "declare -F", "gsh_typeset -F"},
		{"typeset -p", "typeset -p", "gsh_typeset -p"},
		{"declare -p", "declare -p", "gsh_typeset -p"},
		{"extra spaces typeset -f", "typeset  -f", "gsh_typeset -f"},
		{"extra spaces declare -f", "declare  -f", "gsh_typeset -f"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := PreprocessTypesetCommands(tc.input)
			assert.Equal(t, tc.expected, result, "input %q should be transformed to %q", tc.input, tc.expected)
		})
	}
}

func TestPreprocessTypesetCommands_MixedContent(t *testing.T) {
	input := `#!/bin/bash
typeset -f myfunc
declare -F
echo "hello world"
typeset -p VAR1
declare -p VAR2`

	expected := `#!/bin/bash
gsh_typeset -f myfunc
gsh_typeset -F
echo "hello world"
gsh_typeset -p VAR1
gsh_typeset -p VAR2`

	result := PreprocessTypesetCommands(input)
	assert.Equal(t, expected, result, "mixed content should be transformed correctly")
}

func TestPreprocessTypesetCommands_NoTransformation(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"other command", "echo hello", "echo hello"},
		{"typeset without flag", "typeset VAR=value", "typeset VAR=value"},
		{"declare without flag", "declare VAR=value", "declare VAR=value"},
		{"different flag", "typeset -i x=1", "typeset -i x=1"},
		{"comment", "# typeset -f", "# typeset -f"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := PreprocessTypesetCommands(tc.input)
			assert.Equal(t, tc.expected, result, "input %q should remain unchanged", tc.input)
		})
	}
}

func TestPreprocessTypesetCommands_LargeInput(t *testing.T) {
	// Create a large input (but under the 10MB limit)
	largeInput := make([]byte, 1024*1024) // 1MB
	for i := range largeInput {
		largeInput[i] = 'a'
	}
	largeInput[0] = 't'
	largeInput[1] = 'y'
	largeInput[2] = 'p'
	largeInput[3] = 'e'
	largeInput[4] = 's'
	largeInput[5] = 'e'
	largeInput[6] = 't'
	largeInput[7] = ' '
	largeInput[8] = '-'
	largeInput[9] = 'f'

	result := PreprocessTypesetCommands(string(largeInput))
	assert.Contains(t, result, "gsh_typeset -f", "large input should still be processed correctly")
}

func TestPreprocessTypesetCommands_InputSizeLimit(t *testing.T) {
	// Test that inputs over 10MB are truncated
	oversizedInput := make([]byte, 11*1024*1024) // 11MB
	for i := range oversizedInput {
		oversizedInput[i] = 'x'
	}

	// Add a transformation target at the beginning
	oversizedInput[0] = 't'
	oversizedInput[1] = 'y'
	oversizedInput[2] = 'p'
	oversizedInput[3] = 'e'
	oversizedInput[4] = 's'
	oversizedInput[5] = 'e'
	oversizedInput[6] = 't'
	oversizedInput[7] = ' '
	oversizedInput[8] = '-'
	oversizedInput[9] = 'f'

	result := PreprocessTypesetCommands(string(oversizedInput))
	assert.Contains(t, result, "gsh_typeset -f", "oversized input should still process the first portion")
	assert.Less(t, len(result), 11*1024*1024, "result should be truncated to under 11MB")
}
