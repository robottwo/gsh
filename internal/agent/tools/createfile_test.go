package tools

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/stretchr/testify/assert"
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