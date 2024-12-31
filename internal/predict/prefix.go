package predict

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/internal/rag"
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type LLMPrefixPredictor struct {
	llmClient       *openai.Client
	contextProvider *rag.ContextProvider
	logger          *zap.Logger
	modelId         string
	temperature     float32
}

func NewLLMPrefixPredictor(
	runner *interp.Runner,
	contextProvider *rag.ContextProvider,
	logger *zap.Logger,
) *LLMPrefixPredictor {
	llmClient, modelId, temperature := utils.GetLLMClient(runner, utils.FastModel)
	return &LLMPrefixPredictor{
		llmClient:       llmClient,
		contextProvider: contextProvider,
		logger:          logger,
		modelId:         modelId,
		temperature:     float32(temperature),
	}
}

func (p *LLMPrefixPredictor) Predict(input string) (string, error) {
	if strings.HasPrefix(input, "#") {
		// Don't do prediction for agent chat messages
		return "", nil
	}

	systemMessage := `You are gsh, an intelligent shell program.
You will be given a partial bash command prefix entered by the user, enclosed in <prefix> tags.
You are asked to predict what the complete bash command is.

Instructions:
* Based on the prefix and other context, analyze the user's potential intent
* Your prediction must start with the partial command as a prefix
* Your prediction must be a valid, single-line, complete bash command
` + BEST_PRACTICES

	userMessage := fmt.Sprintf(
		`<prefix>%s</prefix>

Additional context to be aware of:
%s`,
		input,
		p.contextProvider.GetContext(rag.ContextRetrievalOptions{Concise: true}),
	)

	p.logger.Debug(
		"predicting using LLM",
		zap.String("system", systemMessage),
		zap.String("user", userMessage),
	)

	chatCompletion, err := p.llmClient.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:       p.modelId,
		Temperature: p.temperature,
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
		ResponseFormat: &PREDICTED_COMMAND_SCHEMA_PARAM,
	})

	if err != nil {
		return "", err
	}

	prediction := predictedCommand{}
	_ = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &prediction)

	p.logger.Debug(
		"LLM prediction response",
		zap.Any("response", prediction),
	)

	return prediction.PredictedCommand, nil
}
