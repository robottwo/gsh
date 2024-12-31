package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/atinylittleshell/gsh/internal/agent/tools"
	"github.com/atinylittleshell/gsh/internal/utils"
	"github.com/fatih/color"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type Agent struct {
	runner    *interp.Runner
	logger    *zap.Logger
	llmClient *openai.Client

	modelId     string
	temperature float32
	messages    []openai.ChatCompletionMessage
}

var RED = color.New(color.FgRed).PrintlnFunc()

func NewAgent(runner *interp.Runner, logger *zap.Logger) *Agent {
	llmClient, modelId, temperature := utils.GetLLMClient(runner, utils.SlowModel)

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
You are gsh, an intelligent shell program. You answer my questions or help me complete tasks.

# Instructions

* Whenever possible, prefer using the bash tool to complete tasks for me rather than telling them how to do it themselves.
* You do not need to complete the task with a single command. You are able to run multiple commands in sequence.
* I'm able to see the output of any bash tool you run so there's no need to repeat that in your response. 
* If you believe the output from the bash commands is sufficient for fulfilling my request, end the conversation by calling the "done" tool.
* If you see a tool call response enclosed in <gsh_tool_call_error> tags, that means the tool call failed; otherwise, the tool call succeeded and whatever you see in the response is the actual result from the tool.

# Best practices

Whenever you are trying to make code changes:
* You can use the "view_directory" tool to understand the structure of the repository
* You can use "git grep" command through the bash tool to help locate relevant files

Whenever you are trying to create a git commit:
* Follow conventional commit message format
* Use "git diff" or "git diff --staged" through the bash tool to understand the changes you are committing
* Make sure commit messages are concise and descriptive of the changes made
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
					Model:       agent.modelId,
					Messages:    agent.messages,
					Temperature: agent.temperature,
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
				RED(fmt.Sprintf("Error sending request to LLM: %s", err))
				agent.logger.Error("Error creating LLM chat stream", zap.Error(err))
				return
			}
			defer stream.Close()

			currentMessageRole := ""
			currentMessageContent := ""
			currentToolCall := openai.ToolCall{
				Type: "function",
			}

			for {
				response, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						return
					}
					RED(fmt.Sprintf("Error receiving response from LLM: %s", err))
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
					hasToolCall := currentToolCall.ID != "" && currentToolCall.Function.Name != ""
					toolCalls := []openai.ToolCall{}
					if hasToolCall {
						toolCalls = append(toolCalls, currentToolCall)
					}
					agent.messages = append(agent.messages, openai.ChatCompletionMessage{
						Role:      currentMessageRole,
						Content:   currentMessageContent,
						ToolCalls: toolCalls,
					})

					if currentMessageContent != "" {
						fmt.Println()
					}

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
	var params map[string]any
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
		agent.logger.Error(fmt.Sprintf("Failed to parse function call arguments: %v", err), zap.String("arguments", toolCall.Function.Arguments))
		RED("LLM responded with something invalid. This is typically an indication that the model being used is not intelligent enough for the current task. Please try again.")
		return false
	}

	toolResponse := fmt.Sprintf("Unknown tool: %s", toolCall.Function.Name)

	switch toolCall.Function.Name {
	case tools.DoneToolDefinition.Function.Name:
		// done
		toolResponse = "ok"
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
		Role:       "tool",
		ToolCallID: toolCall.ID,
		Content:    toolResponse,
	})
	return true
}
