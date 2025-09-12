package tools

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestBashToolDefinition(t *testing.T) {
	assert.Equal(t, openai.ToolType("function"), BashToolDefinition.Type)
	assert.Equal(t, "bash", BashToolDefinition.Function.Name)
	assert.Equal(
		t,
		`Run a single-line command in a bash shell.
* When invoking this tool, the contents of the "command" parameter does NOT need to be XML-escaped.
* Avoid combining multiple bash commands into one using "&&", ";" or multiple lines. Instead, run each command separately.
* State is persistent across command calls and discussions with the user.`,
		BashToolDefinition.Function.Description,
	)
	parameters, ok := BashToolDefinition.Function.Parameters.(*jsonschema.Definition)
	assert.True(t, ok, "Parameters should be of type *jsonschema.Definition")
	assert.Equal(t, jsonschema.DataType("object"), parameters.Type)
	assert.Equal(t, "A concise reason for why you need to run this command", parameters.Properties["reason"].Description)
	assert.Equal(t, jsonschema.DataType("string"), parameters.Properties["reason"].Type)
	assert.Equal(t, "The bash command to run", parameters.Properties["command"].Description)
	assert.Equal(t, jsonschema.DataType("string"), parameters.Properties["command"].Type)
	assert.Equal(t, []string{"reason", "command"}, parameters.Required)
}
func TestGenerateCommandRegex(t *testing.T) {
	// Test regular commands
	assert.Equal(t, "^ls.*", GenerateCommandRegex("ls -la /tmp"))
	assert.Equal(t, "^pwd.*", GenerateCommandRegex("pwd"))
	assert.Equal(t, "^cat.*", GenerateCommandRegex("cat file.txt"))

	// Test special commands with subcommands
	assert.Equal(t, "^git status.*", GenerateCommandRegex("git status"))
	assert.Equal(t, "^git commit.*", GenerateCommandRegex("git commit -m \"message\""))
	assert.Equal(t, "^npm install.*", GenerateCommandRegex("npm install package"))
	assert.Equal(t, "^npm run.*", GenerateCommandRegex("npm run test"))
	assert.Equal(t, "^yarn add.*", GenerateCommandRegex("yarn add package"))
	assert.Equal(t, "^docker run.*", GenerateCommandRegex("docker run image"))
	assert.Equal(t, "^kubectl get.*", GenerateCommandRegex("kubectl get pods"))

	// Test edge cases
	assert.Equal(t, "^$", GenerateCommandRegex(""))
}
