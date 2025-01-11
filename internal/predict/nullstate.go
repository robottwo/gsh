package predict

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type LLMNullStatePredictor struct {
	runner      *interp.Runner
	llmClient   *openai.Client
	contextText string
	logger      *zap.Logger
	modelId     string
	temperature float32
}

func NewLLMNullStatePredictor(
	runner *interp.Runner,
	logger *zap.Logger,
) *LLMNullStatePredictor {
	llmClient, modelId, temperature := utils.GetLLMClient(runner, utils.FastModel)
	return &LLMNullStatePredictor{
		runner:      runner,
		llmClient:   llmClient,
		contextText: "",
		logger:      logger,
		modelId:     modelId,
		temperature: float32(temperature),
	}
}

func (p *LLMNullStatePredictor) UpdateContext(context *map[string]string) {
	contextTypes := environment.GetContextTypesForPredictionWithoutPrefix(p.runner, p.logger)
	p.contextText = utils.ComposeContextText(context, contextTypes, p.logger)
}

func (p *LLMNullStatePredictor) Predict(input string) (string, error) {
	if input != "" {
		// this predictor is only for null state
		return "", nil
	}

	schema, err := PREDICTED_COMMAND_SCHEMA.MarshalJSON()
	if err != nil {
		return "", err
	}

	systemMessage := fmt.Sprintf(`You are gsh, an intelligent shell program.
You are asked to predict the next command I'm likely to want to run.

# Instructions
* Based on the context, analyze the my potential intent
* Your prediction must be a valid, single-line, complete bash command

# Best Practices
%s

# Latest Context
%s

# Response JSON Schema
%s`,
		BEST_PRACTICES,
		p.contextText,
		string(schema),
	)

	userMessage := `Now predict what my next command should be.`

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
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
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
