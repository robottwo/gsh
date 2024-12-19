package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
)

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

type PredictedCommand struct {
	FullCommand string `json:"full_command" jsonschema_description:"The full bash command predicted by the model" jsonschema_required:"true"`
}

var PREDICTED_COMMAND_SCHEMA = GenerateSchema[PredictedCommand]()

var PREDICTED_COMMAND_SCHEMA_PARAM = openai.ResponseFormatJSONSchemaJSONSchemaParam{
	Name:        openai.F("predicted_command"),
	Description: openai.F("The predicted bash command"),
	Schema:      openai.F(PREDICTED_COMMAND_SCHEMA),
	Strict:      openai.Bool(true),
}

func predictInput(llmClient *openai.Client, userInput string) (string, error) {
	chatCompletion, err := llmClient.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf(`
You are gsh, an intelligent shell program.
You are asked to predict a complete bash command based on a partial one from the user.

<partial_command>
%s
</partial_command>`, userInput)),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(PREDICTED_COMMAND_SCHEMA_PARAM),
			},
		),
		Model: openai.F("qwen2.5"),
	})

	if err != nil {
		return "", err
	}

	predictedCommand := PredictedCommand{}
	_ = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &predictedCommand)

	return predictedCommand.FullCommand, nil
}
