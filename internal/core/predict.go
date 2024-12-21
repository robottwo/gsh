package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type predictedCommand struct {
	FullCommand string `json:"full_command" jsonschema_description:"The full bash command predicted by the model" jsonschema_required:"true"`
}

var PREDICTED_COMMAND_SCHEMA = GenerateJsonSchema[predictedCommand]()

var PREDICTED_COMMAND_SCHEMA_PARAM = openai.ResponseFormatJSONSchemaJSONSchemaParam{
	Name:        openai.F("predicted_command"),
	Description: openai.F("The predicted bash command"),
	Schema:      openai.F(PREDICTED_COMMAND_SCHEMA),
	Strict:      openai.Bool(true),
}

type LLMPredictor struct {
	llmClient *openai.Client
}

func NewLLMPredictor() *LLMPredictor {
	llmClient := openai.NewClient(
		option.WithAPIKey("ollama"),
		option.WithBaseURL("http://localhost:11434/v1/"),
	)
	return &LLMPredictor{llmClient: llmClient}
}

func (p *LLMPredictor) Predict(input string) (string, error) {
	chatCompletion, err := p.llmClient.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf(`
You are gsh, an intelligent shell program.
You are asked to predict a complete bash command based on a partial one from the user.

<partial_command>
%s
</partial_command>`, input)),
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

	predictedCommand := predictedCommand{}
	_ = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &predictedCommand)

	return predictedCommand.FullCommand, nil
}
