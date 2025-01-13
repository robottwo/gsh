package tools

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestViewDirectoryToolDefinition(t *testing.T) {
	assert.Equal(t, openai.ToolType("function"), ViewDirectoryToolDefinition.Type)
	assert.Equal(t, "view_directory", ViewDirectoryToolDefinition.Function.Name)
	assert.Equal(
		t,
		"View the content in a directory up to 2 levels deep.",
		ViewDirectoryToolDefinition.Function.Description,
	)
	parameters, ok := ViewDirectoryToolDefinition.Function.Parameters.(*jsonschema.Definition)
	assert.True(t, ok, "Parameters should be of type *jsonschema.Definition")
	assert.Equal(t, jsonschema.DataType("object"), parameters.Type)
	assert.Equal(t, "Absolute path to the directory", parameters.Properties["path"].Description)
	assert.Equal(t, jsonschema.DataType("string"), parameters.Properties["path"].Type)
	assert.Equal(t, []string{"path"}, parameters.Required)
}