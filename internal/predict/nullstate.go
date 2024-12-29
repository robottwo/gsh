package predict

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atinylittleshell/gsh/internal/rag"
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type LLMNullStatePredictor struct {
	llmClient       *openai.Client
	contextProvider *rag.ContextProvider
	logger          *zap.Logger
	modelId         string
	temperature     float32
}

func NewLLMNullStatePredictor(
	runner *interp.Runner,
	contextProvider *rag.ContextProvider,
	logger *zap.Logger,
) *LLMNullStatePredictor {
	llmClient, modelId, temperature := utils.GetLLMClient(runner, utils.FastModel)
	return &LLMNullStatePredictor{
		llmClient:       llmClient,
		contextProvider: contextProvider,
		logger:          logger,
		modelId:         modelId,
		temperature:     float32(temperature),
	}
}

func (p *LLMNullStatePredictor) Predict(input string, directory string) (string, error) {
	if input != "" {
		// this predictor is only for null state
		return "", nil
	}

	systemMessage := `You are gsh, an intelligent shell program.
You are asked to predict the next command the user is likely to want to run.

Instructions:
* Based on the context, analyze the user's potential intent
* Your prediction must be a valid, single-line, complete bash command`

	userMessage := fmt.Sprintf(
		`Context:
%s

Now predict what the next command should be.`,
		input,
		p.contextProvider.GetContext(),
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
