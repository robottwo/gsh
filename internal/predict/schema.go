package predict

import (
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
)

type predictedCommand struct {
	Thought          string `json:"thought" jsonschema_description:"Your step by step thinking for what my intent might be" jsonschema_required:"true"`
	PredictedCommand string `json:"predicted_command" jsonschema_description:"The full bash command predicted by the model" jsonschema_required:"true"`
}

var PREDICTED_COMMAND_SCHEMA = utils.GenerateJsonSchema(&predictedCommand{})

var PREDICTED_COMMAND_SCHEMA_PARAM = openai.ChatCompletionResponseFormat{
	Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
	JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "prediction",
		Schema: PREDICTED_COMMAND_SCHEMA,
		Strict: true,
	},
}

type explainedCommand struct {
	Explanation string `json:"explanation" jsonschema_description:"A concise explanation of what the command will do for me" jsonschema_required:"true"`
}

var EXPLAINED_COMMAND_SCHEMA = utils.GenerateJsonSchema(&explainedCommand{})

var EXPLAINED_COMMAND_SCHEMA_PARAM = openai.ChatCompletionResponseFormat{
	Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
	JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "explanation",
		Schema: EXPLAINED_COMMAND_SCHEMA,
		Strict: true,
	},
}
