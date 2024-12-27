package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type Agent struct {
	runner    *interp.Runner
	logger    *zap.Logger
	llmClient *openai.Client

	modelId     string
	temperature float64
	messages    []openai.ChatCompletionMessage
}

func NewAgent(runner *interp.Runner, logger *zap.Logger) *Agent {
	apiKey := runner.Vars["GSH_SLOW_MODEL_API_KEY"].String()
	if apiKey == "" {
		apiKey = "ollama"
	}
	baseURL := runner.Vars["GSH_SLOW_MODEL_BASE_URL"].String()
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1/"
	}
	modelId := runner.Vars["GSH_SLOW_MODEL_ID"].String()
	if modelId == "" {
		modelId = "qwen2.5:32b"
	}
	temperature, err := strconv.ParseFloat(runner.Vars["GSH_SLOW_MODEL_TEMPERATURE"].String(), 64)
	if err != nil {
		temperature = 0.1
	}

	llmClientConfig := openai.DefaultConfig(apiKey)
	llmClientConfig.BaseURL = baseURL

	llmClient := openai.NewClientWithConfig(llmClientConfig)

	return &Agent{
		runner:      runner,
		logger:      logger,
		llmClient:   llmClient,
		modelId:     modelId,
		temperature: temperature,
		messages: []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: `
You are gsh, an intelligent shell program. You answer users' questions or help them complete tasks.
* Whenever possible, prefer using the bash tool to complete tasks for users rather than telling them how to do it themselves.
* The user is able to see the output of any bash tool you run so there's no need to repeat that in your response. 
* If you believe the output from the bash commands is sufficient for fulfilling the user's request, end the conversation by calling the "done" tool.
        `,
			},
		},
	}
}

func (agent *Agent) Chat(prompt string) (<-chan string, error) {
	appendMessage := openai.ChatCompletionMessage{
		Role:    "user",
		Content: prompt,
	}
	agent.messages = append(agent.messages, appendMessage)

	responseChannel := make(chan string)

	go func() {
		defer close(responseChannel)

		continueSession := true

		for continueSession {
			// By default the session should stop after the first response, unless we handled a tool call,
			// in which case we'll set this to true and continue the session.
			continueSession = false

			stream, err := agent.llmClient.CreateChatCompletionStream(
				context.Background(),
				openai.ChatCompletionRequest{
					Model:    agent.modelId,
					Messages: agent.messages,
					Tools: []openai.Tool{
						doneToolDefinition,
						bashToolDefinition,
					},
					Stream: true,
				},
			)
			if err != nil {
				agent.logger.Error("Error creating LLM chat stream", zap.Error(err))
				return
			}
			defer stream.Close()

			currentMessageRole := ""
			currentMessageContent := ""
			currentToolCall := openai.ToolCall{}

			for {
				response, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						return
					}
					agent.logger.Error("Error receiving LLM chat response", zap.Error(err))
					return
				}

				msg := response.Choices[0]
				agent.logger.Debug("LLM chat response", zap.Any("message", msg))

				if len(msg.Delta.ToolCalls) > 0 {
					currentToolCall.ID += msg.Delta.ToolCalls[0].ID
					currentToolCall.Function.Name += msg.Delta.ToolCalls[0].Function.Name
					currentToolCall.Function.Arguments += msg.Delta.ToolCalls[0].Function.Arguments
				}
				if msg.Delta.Role != "" {
					currentMessageRole += msg.Delta.Role
				}
				if msg.Delta.Content != "" {
					currentMessageContent += msg.Delta.Content
					responseChannel <- msg.Delta.Content
				}

				if msg.FinishReason == "stop" || msg.FinishReason == "tool_calls" || msg.FinishReason == "function_call" {
					if currentMessageContent != "" {
						agent.messages = append(agent.messages, openai.ChatCompletionMessage{
							Role:    currentMessageRole,
							Content: currentMessageContent,
						})
						fmt.Println()
					}

					hasToolCall := currentToolCall.ID != "" && currentToolCall.Function.Name != ""
					if hasToolCall {
						continueSession = agent.handleToolCall(currentToolCall)
					}
					break
				} else if msg.FinishReason != "" {
					agent.logger.Warn("LLM chat response finished for unexpected reason", zap.String("reason", string(msg.FinishReason)))
					break
				}
			}
		}
	}()

	return responseChannel, nil
}

func (agent *Agent) handleToolCall(toolCall openai.ToolCall) bool {
	// the "done" tool is a special tool that signals the end of the conversation
	if toolCall.Function.Name == "done" {
		return false
	}

	var params map[string]string
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
		agent.logger.Error("Failed to parse function call arguments", zap.Error(err))
		return false
	}

	toolResponse := fmt.Sprintf("Unknown tool: %s", toolCall.Function.Name)

	if toolCall.Function.Name == "bash" {
		command := params["command"]
		stdout, stderr, exitCode, executed := bashTool(agent.runner, agent.logger, command)

		jsonBuffer, err := json.Marshal(map[string]string{
			"command_executed": strconv.FormatBool(executed),
			"stdout":           stdout,
			"stderr":           stderr,
			"exitCode":         strconv.Itoa(exitCode),
		})
		if err != nil {
			agent.logger.Error("Failed to marshal tool response", zap.Error(err))
			return false
		}

		toolResponse = string(jsonBuffer)
	}

	agent.messages = append(agent.messages, openai.ChatCompletionMessage{
		Role:    "tool",
		Content: toolResponse,
	})
	return true
}
