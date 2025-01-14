package tools

import (
	"testing"

	"github.com/atinylittleshell/gsh/internal/filesystem"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

func TestValidateAndExtractParams(t *testing.T) {
	logger := zap.NewNop()
	runner, _ := interp.New()

	tests := []struct {
		name           string
		params         map[string]any
		expectedParams *editFileParams
		expectedError  string
	}{
		{
			name: "valid parameters",
			params: map[string]any{
				"path":    "/test/path",
				"old_str": "old content",
				"new_str": "new content",
			},
			expectedParams: &editFileParams{
				path:   "/test/path",
				oldStr: "old content",
				newStr: "new content",
			},
			expectedError: "",
		},
		{
			name: "missing path",
			params: map[string]any{
				"old_str": "old content",
				"new_str": "new content",
			},
			expectedParams: nil,
			expectedError: "The create_file tool failed to parse parameter 'path'",
		},
		{
			name: "missing old_str",
			params: map[string]any{
				"path":    "/test/path",
				"new_str": "new content",
			},
			expectedParams: nil,
			expectedError: "The create_file tool failed to parse parameter 'old_str'",
		},
		{
			name: "missing new_str",
			params: map[string]any{
				"path":    "/test/path",
				"old_str": "old content",
			},
			expectedParams: nil,
			expectedError: "The create_file tool failed to parse parameter 'new_str'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, errMsg := validateAndExtractParams(runner, logger, tt.params)
			assert.Equal(t, tt.expectedParams, params)
			assert.Equal(t, tt.expectedError, errMsg)
		})
	}
}

func TestValidateAndReplaceContent(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		oldStr        string
		newStr        string
		expectedOut   string
		expectedError string
	}{
		{
			name:          "successful replacement",
			content:       "Hello world!",
			oldStr:        "world",
			newStr:        "there",
			expectedOut:   "Hello there!",
			expectedError: "",
		},
		{
			name:          "no matches",
			content:       "Hello world!",
			oldStr:        "foo",
			newStr:        "bar",
			expectedOut:   "",
			expectedError: "The old string must be unique in the file",
		},
		{
			name:          "multiple matches",
			content:       "Hello world world!",
			oldStr:        "world",
			newStr:        "there",
			expectedOut:   "",
			expectedError: "The old string must be unique in the file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newContent, errMsg := validateAndReplaceContent(tt.content, tt.oldStr, tt.newStr)
			assert.Equal(t, tt.expectedOut, newContent)
			assert.Equal(t, tt.expectedError, errMsg)
		})
	}
}

type mockFileSystem struct {
	filesystem.FileSystem
	readFileError  error
	writeFileError error
	fileContent    string
}

func (m *mockFileSystem) ReadFile(path string) (string, error) {
	if m.readFileError != nil {
		return "", m.readFileError
	}
	return m.fileContent, nil
}

func (m *mockFileSystem) WriteFile(path string, content string) error {
	return m.writeFileError
}

func TestReadFileContents(t *testing.T) {
	logger := zap.NewNop()
	tests := []struct {
		name          string
		fs            *mockFileSystem
		expectedOut   string
		expectedError string
	}{
		{
			name: "successful read",
			fs: &mockFileSystem{
				fileContent: "test content",
			},
			expectedOut:   "test content",
			expectedError: "",
		},
		{
			name: "read error",
			fs: &mockFileSystem{
				readFileError: assert.AnError,
			},
			expectedOut:   "",
			expectedError: "Error reading file: assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, errMsg := readFileContents(logger, tt.fs, "/test/path")
			assert.Equal(t, tt.expectedOut, content)
			assert.Equal(t, tt.expectedError, errMsg)
		})
	}
}

func TestWriteFile(t *testing.T) {
	logger := zap.NewNop()
	tests := []struct {
		name          string
		fs            *mockFileSystem
		expectedError string
	}{
		{
			name: "successful write",
			fs: &mockFileSystem{
				writeFileError: nil,
			},
			expectedError: "",
		},
		{
			name: "write error",
			fs: &mockFileSystem{
				writeFileError: assert.AnError,
			},
			expectedError: "Error writing to file: assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := writeFile(logger, tt.fs, "/test/path", "test content")
			assert.Equal(t, tt.expectedError, errMsg)
		})
	}
}