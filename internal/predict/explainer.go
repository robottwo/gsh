package predict

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/rag"
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type LLMExplainer struct {
	runner          *interp.Runner
	llmClient       *openai.Client
	contextProvider *rag.ContextProvider
	logger          *zap.Logger
	modelId         string
	temperature     float32
}

func NewLLMExplainer(
	runner *interp.Runner,
	contextProvider *rag.ContextProvider,
	logger *zap.Logger,
) *LLMExplainer {
	llmClient, modelId, temperature := utils.GetLLMClient(runner, utils.FastModel)
	return &LLMExplainer{
		runner:          runner,
		llmClient:       llmClient,
		contextProvider: contextProvider,
		logger:          logger,
		modelId:         modelId,
		temperature:     float32(temperature),
	}
}

func (e *LLMExplainer) Explain(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	systemMessage := `You are gsh, an intelligent shell program.
You will be given a bash command entered by me, enclosed in <command> tags.

Instructions:
* Give a concise explanation of what the command will do for me
* If any uncommon arguments are present in the command, 
  format your explanation in markdown and explain arguments in a bullet point list
`

	userMessage := fmt.Sprintf(
		`<command>%s</command>

Context:
%s`,
		input,
		e.contextProvider.GetContext(
			rag.ContextRetrievalOptions{
				Concise:      false,
				HistoryLimit: environment.GetHistoryContextLimit(e.runner, e.logger),
			},
		),
	)

	e.logger.Debug(
		"explaining prediction using LLM",
		zap.String("system", systemMessage),
		zap.String("user", userMessage),
	)

	chatCompletion, err := e.llmClient.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:       e.modelId,
		Temperature: e.temperature,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: systemMessage,
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		ResponseFormat: &EXPLAINED_COMMAND_SCHEMA_PARAM,
	})

	if err != nil {
		return "", err
	}

	explanation := explainedCommand{}
	_ = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &explanation)

	e.logger.Debug(
		"LLM explanation response",
		zap.Any("response", explanation),
	)

	return explanation.Explanation, nil
}
