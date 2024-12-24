package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atinylittleshell/gsh/internal/history"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.uber.org/zap"
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
	llmClient      *openai.Client
	historyManager *history.HistoryManager
	logger         *zap.Logger
}

func NewLLMPredictor(historyManager *history.HistoryManager, logger *zap.Logger) *LLMPredictor {
	llmClient := openai.NewClient(
		option.WithAPIKey("ollama"),
		option.WithBaseURL("http://localhost:11434/v1/"),
	)
	return &LLMPredictor{
		llmClient:      llmClient,
		historyManager: historyManager,
		logger:         logger,
	}
}

func (p *LLMPredictor) Predict(input string, directory string) (string, error) {
	historyEntries, err := p.historyManager.GetRecentEntries(10)
	if err != nil {
		return "", err
	}

	var commandHistory string
	for _, entry := range historyEntries {
		commandHistory += fmt.Sprintf("%s: %s\n", entry.Directory, entry.Command)
	}

	p.logger.Info("predicting with history", zap.String("commandHistory", commandHistory))

	llmPrompt := fmt.Sprintf(
		`
You are gsh, an intelligent shell program.
You are asked to predict a complete bash command based on a partial one from the user.

<command_history>
%s
</command_history>

<current_directory>
%s
</current_directory>

<partial_command>
%s
</partial_command>
`,
		commandHistory,
		directory,
		input,
	)

	chatCompletion, err := p.llmClient.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(llmPrompt),
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

	p.logger.Info("predicting using LLM", zap.String("prompt", llmPrompt))

	predictedCommand := predictedCommand{}
	_ = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &predictedCommand)

	return predictedCommand.FullCommand, nil
}
