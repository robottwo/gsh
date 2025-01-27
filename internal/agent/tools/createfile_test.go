package tools

import (
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
	"os"
	"testing"
)

func TestCreateFileToolDefinition(t *testing.T) {
	assert.Equal(t, openai.ToolType("function"), CreateFileToolDefinition.Type)
	assert.Equal(t, "create_file", CreateFileToolDefinition.Function.Name)
	assert.Equal(
		t,
		"Create a file with the specified content.",
		CreateFileToolDefinition.Function.Description,
	)
	parameters, ok := CreateFileToolDefinition.Function.Parameters.(*jsonschema.Definition)
	assert.True(t, ok, "Parameters should be of type *jsonschema.Definition")
	assert.Equal(t, jsonschema.DataType("object"), parameters.Type)
	assert.Equal(t, "Absolute path to the file", parameters.Properties["path"].Description)
	assert.Equal(t, jsonschema.DataType("string"), parameters.Properties["path"].Type)
	assert.Equal(t, "The content to write to the file", parameters.Properties["content"].Description)
	assert.Equal(t, jsonschema.DataType("string"), parameters.Properties["content"].Type)
	assert.Equal(t, []string{"path", "content"}, parameters.Required)
}

func TestCreateFileToolParams(t *testing.T) {
	logger := zap.NewNop()
	runner, _ := interp.New()

	origUserConfirmation := userConfirmation
	userConfirmation = func(logger *zap.Logger, question string, explanation string) string {
		return "y"
	}
	defer func() { userConfirmation = origUserConfirmation }()

	tests := []struct {
		name          string
		params        map[string]any
		expectedError bool
	}{
		{
			name: "valid parameters",
			params: map[string]any{
				"path":    "/test/path",
				"content": "test content",
			},
			expectedError: false,
		},
		{
			name: "missing path",
			params: map[string]any{
				"content": "test content",
			},
			expectedError: true,
		},
		{
			name: "missing content",
			params: map[string]any{
				"path": "/test/path",
			},
			expectedError: true,
		},
		{
			name: "invalid path type",
			params: map[string]any{
				"path":    123,
				"content": "test content",
			},
			expectedError: true,
		},
		{
			name: "invalid content type",
			params: map[string]any{
				"path":    "/test/path",
				"content": 123,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateFileTool(runner, logger, tt.params)
			if tt.expectedError {
				assert.Contains(t, result, "failed")
			} else {
				// Since we can't actually create files in this test, we expect it to fail at file creation
				assert.Contains(t, result, "Error creating")
			}
		})
	}
}

func TestCreateFile(t *testing.T) {
	logger := zap.NewNop()
	runner, _ := interp.New()

	origUserConfirmation := userConfirmation
	userConfirmation = func(logger *zap.Logger, question string, explanation string) string {
		return "y"
	}
	defer func() { userConfirmation = origUserConfirmation }()

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "gsh_create_file_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		name          string
		path          string
		content       string
		expectedError bool
	}{
		{
			name:          "successful create",
			path:          tmpFile.Name(),
			content:       "test content",
			expectedError: false,
		},
		{
			name:          "invalid path",
			path:          "/nonexistent/directory/file.txt",
			content:       "test content",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]any{
				"path":    tt.path,
				"content": tt.content,
			}
			result := CreateFileTool(runner, logger, params)
			if tt.expectedError {
				assert.Contains(t, result, "Error")
			} else {
				assert.Contains(t, result, "successfully")
			}
		})
	}
}

