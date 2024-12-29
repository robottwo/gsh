package predict

import (
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
)

type predictedCommand struct {
	UserIntent       string `json:"user_intent" jsonschema_description:"Your concise analysis of the user's intent" jsonschema_required:"true"`
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
