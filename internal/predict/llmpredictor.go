package predict

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/atinylittleshell/gsh/internal/rag"
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type predictedCommand struct {
	UserIntent       string `json:"user_intent" jsonschema_description:"Your concise analysis of the user's intent" jsonschema_required:"true"`
	PredictedCommand string `json:"predicted_command" jsonschema_description:"The full bash command predicted by the model" jsonschema_required:"true"`
}

var PREDICTED_COMMAND_SCHEMA = utils.GenerateJsonSchema(&predictedCommand{})

var PREDICTED_COMMAND_SCHEMA_PARAM = openai.ChatCompletionResponseFormat{
	Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
	JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "prediction",
		Schema: PREDICTED_COMMAND_SCHEMA,
		Strict: true,
	},
}

type LLMPredictor struct {
	llmClient       *openai.Client
	contextProvider *rag.ContextProvider
	logger          *zap.Logger
	modelId         string
	temperature     float32
}

func NewLLMPredictor(
	runner *interp.Runner,
	contextProvider *rag.ContextProvider,
	logger *zap.Logger,
) *LLMPredictor {
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

	var headers map[string]string
	json.Unmarshal([]byte(runner.Vars["GSH_SLOW_MODEL_HEADERS"].String()), &headers)

	llmClientConfig := openai.DefaultConfig(apiKey)
	llmClientConfig.BaseURL = baseURL
	llmClientConfig.HTTPClient = utils.NewLLMHttpClient(headers)

	llmClient := openai.NewClientWithConfig(llmClientConfig)
	return &LLMPredictor{
		llmClient:       llmClient,
		contextProvider: contextProvider,
		logger:          logger,
		modelId:         modelId,
		temperature:     float32(temperature),
	}
}

func (p *LLMPredictor) Predict(input string, directory string) (string, error) {
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
* Your prediction must be a valid, single-line, complete bash command`

	userMessage := fmt.Sprintf(
		`<prefix>%s</prefix>

Additional context to be aware of:
%s`,
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
