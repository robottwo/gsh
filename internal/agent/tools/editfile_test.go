package tools

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestEditFileToolDefinition(t *testing.T) {
	assert.Equal(t, openai.ToolType("function"), EditFileToolDefinition.Type)
	assert.Equal(t, "edit_file", EditFileToolDefinition.Function.Name)
	assert.Equal(
		t,
		"Edit the content of a file.",
		EditFileToolDefinition.Function.Description,
	)
	parameters, ok := EditFileToolDefinition.Function.Parameters.(*jsonschema.Definition)
	assert.True(t, ok, "Parameters should be of type *jsonschema.Definition")
	assert.Equal(t, jsonschema.DataType("object"), parameters.Type)
	assert.Equal(t, "Absolute path to the file", parameters.Properties["path"].Description)
	assert.Equal(t, jsonschema.DataType("string"), parameters.Properties["path"].Type)
	assert.Equal(t, "The old string in the file to be replaced. This must be a unique chunk in the file, ideally complete lines.", parameters.Properties["old_str"].Description)
	assert.Equal(t, jsonschema.DataType("string"), parameters.Properties["old_str"].Type)
	assert.Equal(t, "The new string that will replace the old one", parameters.Properties["new_str"].Description)
	assert.Equal(t, jsonschema.DataType("string"), parameters.Properties["new_str"].Type)
	assert.Equal(t, []string{"path", "old_str", "new_str"}, parameters.Required)
}