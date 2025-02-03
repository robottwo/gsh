package predict

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/history"
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type LLMPrefixPredictor struct {
	runner            *interp.Runner
	historyManager    *history.HistoryManager
	llmClient         *openai.Client
	contextText       string
	logger            *zap.Logger
	modelId           string
	temperature       float32
	numHistoryContext int
}

func NewLLMPrefixPredictor(
	runner *interp.Runner,
	historyManager *history.HistoryManager,
	logger *zap.Logger,
) *LLMPrefixPredictor {
	llmClient, modelId, temperature := utils.GetLLMClient(runner, utils.FastModel)
	return &LLMPrefixPredictor{
		runner:         runner,
		historyManager: historyManager,
		llmClient:      llmClient,
		contextText:    "",
		logger:         logger,
		modelId:        modelId,
		temperature:    float32(temperature),
	}
}

func (p *LLMPrefixPredictor) UpdateContext(context *map[string]string) {
	contextTypes := environment.GetContextTypesForPredictionWithPrefix(p.runner, p.logger)
	p.contextText = utils.ComposeContextText(context, contextTypes, p.logger)
	p.numHistoryContext = environment.GetContextNumHistoryConcise(p.runner, p.logger)
}

func (p *LLMPrefixPredictor) Predict(input string) (string, string, error) {
	if strings.HasPrefix(input, "#") {
		// Don't do prediction for agent chat messages
		return "", "", nil
	}

	schema, err := PREDICTED_COMMAND_SCHEMA.MarshalJSON()
	if err != nil {
		return "", "", err
	}

	matchingHistoryEntries, err := p.historyManager.GetRecentEntriesByPrefix(
		input,
		p.numHistoryContext,
	)
	matchingHistoryContext := strings.Builder{}
	if err == nil {
		for _, entry := range matchingHistoryEntries {
			matchingHistoryContext.WriteString(fmt.Sprintf(
				"%s\n",
				entry.Command,
			))
		}
	}

	userMessage := fmt.Sprintf(`You are gsh, an intelligent shell program.
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

# Previous Commands with Similar Prefix
%s

# Response JSON Schema
%s

<prefix>%s</prefix>`,
		BEST_PRACTICES,
		p.contextText,
		matchingHistoryContext.String(),
		string(schema),
		input,
	)

	p.logger.Debug(
		"predicting using LLM",
		zap.String("user", userMessage),
	)

	chatCompletion, err := p.llmClient.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:       p.modelId,
		Temperature: p.temperature,
		Messages: []openai.ChatCompletionMessage{
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
		return "", "", err
	}

	prediction := PredictedCommand{}
	_ = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &prediction)

	p.logger.Debug(
		"LLM prediction response",
		zap.Any("response", prediction),
	)

	return prediction.PredictedCommand, userMessage, nil
}
