package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/atinylittleshell/gsh/internal/history"
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type predictedCommand struct {
	FullCommand string `json:"full_command" jsonschema_description:"The full bash command predicted by the model" jsonschema_required:"true"`
}

var PREDICTED_COMMAND_SCHEMA = utils.GenerateJsonSchema(&predictedCommand{})

var PREDICTED_COMMAND_SCHEMA_PARAM = openai.ChatCompletionResponseFormat{
	Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
	JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
		Name:        "predicted_command",
		Description: "The predicted bash command",
		Schema:      PREDICTED_COMMAND_SCHEMA,
		Strict:      true,
	},
}

type LLMPredictor struct {
	llmClient      *openai.Client
	historyManager *history.HistoryManager
	logger         *zap.Logger
	modelId        string
	temperature    float32
}

func NewLLMPredictor(runner *interp.Runner, historyManager *history.HistoryManager, logger *zap.Logger) *LLMPredictor {
	apiKey := runner.Vars["GSH_FAST_MODEL_API_KEY"].String()
	if apiKey == "" {
		apiKey = "ollama"
	}
	baseURL := runner.Vars["GSH_FAST_MODEL_BASE_URL"].String()
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1/"
	}
	modelId := runner.Vars["GSH_FAST_MODEL_ID"].String()
	if modelId == "" {
		modelId = "qwen2.5"
	}
	temperature, err := strconv.ParseFloat(runner.Vars["GSH_FAST_MODEL_TEMPERATURE"].String(), 32)
	if err != nil {
		temperature = 0.1
	}

	llmClientConfig := openai.DefaultConfig(apiKey)
	llmClientConfig.BaseURL = baseURL

	llmClient := openai.NewClientWithConfig(llmClientConfig)
	return &LLMPredictor{
		llmClient:      llmClient,
		historyManager: historyManager,
		logger:         logger,
		modelId:        modelId,
		temperature:    float32(temperature),
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

	chatCompletion, err := p.llmClient.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:       p.modelId,
		Temperature: p.temperature,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: llmPrompt,
			},
		},
		ResponseFormat: &PREDICTED_COMMAND_SCHEMA_PARAM,
	})

	if err != nil {
		return "", err
	}

	p.logger.Info("predicting using LLM", zap.String("prompt", llmPrompt))

	predictedCommand := predictedCommand{}
	_ = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &predictedCommand)

	return predictedCommand.FullCommand, nil
}
