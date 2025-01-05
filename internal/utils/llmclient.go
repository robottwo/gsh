package utils

import (
	"encoding/json"
	"strconv"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"mvdan.cc/sh/v3/interp"
)

type LLMModelType string

const (
	FastModel LLMModelType = "FAST"
	SlowModel LLMModelType = "SLOW"
)

func GetLLMClient(runner *interp.Runner, modelType LLMModelType) (*openai.Client, string, float32) {
	varPrefix := "GSH_" + string(modelType) + "_MODEL_"

	apiKey := runner.Vars[varPrefix+"API_KEY"].String()
	if apiKey == "" {
		apiKey = "ollama"
	}
	baseURL := runner.Vars[varPrefix+"BASE_URL"].String()
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1/"
	}
	modelId := runner.Vars[varPrefix+"ID"].String()
	if modelId == "" {
		modelId = "qwen2.5"
	}
	temperature, err := strconv.ParseFloat(varPrefix+runner.Vars["TEMPERATURE"].String(), 32)
	if err != nil {
		temperature = 0.1
	}

	var headers map[string]string
	json.Unmarshal([]byte(runner.Vars[varPrefix+"HEADERS"].String()), &headers)

	// Special headers for the openrouter.ai API
	if strings.HasPrefix(strings.ToLower(baseURL), "https://openrouter.ai/") {
		headers["HTTP-Referer"] = "https://github.com/atinylittleshell/gsh"
		headers["X-Title"] = "gsh - The Generative Shell"
	}

	llmClientConfig := openai.DefaultConfig(apiKey)
	llmClientConfig.BaseURL = baseURL
	llmClientConfig.HTTPClient = NewLLMHttpClient(headers)

	return openai.NewClientWithConfig(llmClientConfig), modelId, float32(temperature)
}
