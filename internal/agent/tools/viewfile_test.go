package tools

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestViewFileToolDefinition(t *testing.T) {
	assert.Equal(t, openai.ToolType("function"), ViewFileToolDefinition.Type)
	assert.Equal(t, "view_file", ViewFileToolDefinition.Function.Name)
	assert.Equal(
		t,
		"View the content of a text file, at most 100 lines at a time. If the content is too large, tail will be truncated and replaced with <gsh:truncated />.",
		ViewFileToolDefinition.Function.Description,
	)
	parameters, ok := ViewFileToolDefinition.Function.Parameters.(*jsonschema.Definition)
	assert.True(t, ok, "Parameters should be of type *jsonschema.Definition")
	assert.Equal(t, jsonschema.DataType("object"), parameters.Type)
	assert.Equal(t, "Absolute path to the file", parameters.Properties["path"].Description)
	assert.Equal(t, jsonschema.DataType("string"), parameters.Properties["path"].Type)
	assert.Equal(
		t,
		"Optional. Line number to start viewing. The first line in the file has line number 1. If not provided, we will read from the beginning of the file.",
		parameters.Properties["start_line"].Description,
	)
	assert.Equal(t, jsonschema.DataType("integer"), parameters.Properties["start_line"].Type)
	assert.Equal(t, []string{"path"}, parameters.Required)
}
