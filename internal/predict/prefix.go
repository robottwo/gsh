package predict

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type LLMPrefixPredictor struct {
	runner      *interp.Runner
	llmClient   *openai.Client
	contextText string
	logger      *zap.Logger
	modelId     string
	temperature float32
}

func NewLLMPrefixPredictor(
	runner *interp.Runner,
	logger *zap.Logger,
) *LLMPrefixPredictor {
	llmClient, modelId, temperature := utils.GetLLMClient(runner, utils.FastModel)
	return &LLMPrefixPredictor{
		runner:      runner,
		llmClient:   llmClient,
		contextText: "",
		logger:      logger,
		modelId:     modelId,
		temperature: float32(temperature),
	}
}

func (p *LLMPrefixPredictor) UpdateContext(context *map[string]string) {
	contextTypes := environment.GetContextTypesForPredictionWithPrefix(p.runner, p.logger)
	p.contextText = utils.ComposeContextText(context, contextTypes, p.logger)
}

func (p *LLMPrefixPredictor) Predict(input string) (string, error) {
	if strings.HasPrefix(input, "#") {
		// Don't do prediction for agent chat messages
		return "", nil
	}

	schema, err := PREDICTED_COMMAND_SCHEMA.MarshalJSON()
	if err != nil {
		return "", err
	}

	systemMessage := fmt.Sprintf(`You are gsh, an intelligent shell program.
You will be given a partial bash command prefix entered by me, enclosed in <prefix> tags.
You are asked to predict what the complete bash command is.

# Instructions
* Based on the prefix and other context, analyze the my potential intent
* Your prediction must start with the partial command as a prefix
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

	userMessage := fmt.Sprintf(
		`<prefix>%s</prefix>`,
		input,
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
