package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/atinylittleshell/gsh/internal/agent/tools"
	"github.com/atinylittleshell/gsh/internal/utils"
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

	var headers map[string]string
	json.Unmarshal([]byte(runner.Vars["GSH_SLOW_MODEL_HEADERS"].String()), &headers)

	llmClientConfig := openai.DefaultConfig(apiKey)
	llmClientConfig.BaseURL = baseURL
	llmClientConfig.HTTPClient = utils.NewLLMHttpClient(headers)

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
* You do not need to complete the task with a single command. You are able to run multiple commands in sequence.
* The user is able to see the output of any bash tool you run so there's no need to repeat that in your response. 
* If you believe the output from the bash commands is sufficient for fulfilling the user's request, end the conversation by calling the "done" tool.
* If you see a tool call response enclosed in <gsh_tool_call_error> tags, that means the tool call failed; otherwise, the tool call succeeded and whatever you see in the response is the actual result from the tool.
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
						tools.DoneToolDefinition,
						tools.BashToolDefinition,
						tools.ViewFileToolDefinition,
						tools.ViewDirectoryToolDefinition,
						tools.CreateFileToolDefinition,
						tools.EditFileToolDefinition,
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
				agent.logger.Debug("LLM chat response", zap.Any("messages", agent.messages), zap.Any("response", msg))

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

	var params map[string]any
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
		agent.logger.Error("Failed to parse function call arguments", zap.Error(err), zap.String("arguments", toolCall.Function.Arguments))
		return false
	}

	toolResponse := fmt.Sprintf("Unknown tool: %s", toolCall.Function.Name)

	switch toolCall.Function.Name {
	case tools.BashToolDefinition.Function.Name:
		// bash
		toolResponse = tools.BashTool(agent.runner, agent.logger, params)
	case tools.ViewFileToolDefinition.Function.Name:
		// view_file
		toolResponse = tools.ViewFileTool(agent.runner, agent.logger, params)
	case tools.ViewDirectoryToolDefinition.Function.Name:
		// view_directory
		toolResponse = tools.ViewDirectoryTool(agent.runner, agent.logger, params)
	case tools.CreateFileToolDefinition.Function.Name:
		// create_file
		toolResponse = tools.CreateFileTool(agent.runner, agent.logger, params)
	case tools.EditFileToolDefinition.Function.Name:
		// edit_file
		toolResponse = tools.EditFileTool(agent.runner, agent.logger, params)
	}

	agent.messages = append(agent.messages, openai.ChatCompletionMessage{
		Role:    "tool",
		Content: toolResponse,
	})
	return true
}
