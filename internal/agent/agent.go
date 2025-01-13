package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/internal/agent/tools"
	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/history"
	"github.com/atinylittleshell/gsh/internal/styles"
	"github.com/atinylittleshell/gsh/internal/utils"
	"github.com/atinylittleshell/gsh/pkg/gline"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type Agent struct {
	runner         *interp.Runner
	historyManager *history.HistoryManager
	contextText    string
	logger         *zap.Logger
	llmClient      *openai.Client

	modelId     string
	temperature float32
	messages    []openai.ChatCompletionMessage
}

func NewAgent(
	runner *interp.Runner,
	historyManager *history.HistoryManager,
	logger *zap.Logger,
) *Agent {
	llmClient, modelId, temperature := utils.GetLLMClient(runner, utils.SlowModel)

	return &Agent{
		runner:         runner,
		historyManager: historyManager,
		contextText:    "",
		logger:         logger,
		llmClient:      llmClient,
		modelId:        modelId,
		temperature:    temperature,
		messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "",
			},
		},
	}
}

func (agent *Agent) UpdateContext(context *map[string]string) {
	contextTypes := environment.GetContextTypesForAgent(agent.runner, agent.logger)
	agent.contextText = utils.ComposeContextText(context, contextTypes, agent.logger)
}

// updateSystemMessage resets the system message with latest context
func (agent *Agent) updateSystemMessage() {
	agent.messages[0].Content = `
You are gsh, an intelligent shell program. You answer my questions or help me complete tasks.

# Instructions

* Whenever possible, prefer using the bash tool to complete tasks for me rather than telling them how to do it themselves.
* You do not need to complete the task with a single command. You are able to run multiple commands in sequence.
* I'm able to see the output of any bash tool you run so there's no need to repeat that in your response. 
* If you see a tool call response enclosed in <gsh_tool_call_error> tags, that means the tool call failed; otherwise, the tool call succeeded and whatever you see in the response is the actual result from the tool.
* Never call multiple tools in parallel. Always call at most one tool at a time.

# Best practices

Whenever you are working in a git repository:
* You can use the "view_directory" tool to understand the structure of the repository
* You can use "git grep" command through the bash tool to help locate relevant code snippets
# You can use "git ls-files | grep <filename>" to find files by name

Whenever you are writing test cases:
* Always read the function or code snippet you are trying to test before writing the test case
* After writing the test case, try to run it and ensure it passes

Whenever you are trying to create a git commit:
* Unless explicitly instructed otherwise, follow conventional commit message format
* Always use "git diff" or "git diff --staged" through the bash tool to 
  understand the changes you are committing before coming up with the commit message
* Make sure commit messages are concise and descriptive of the changes made

# Latest Context
` + agent.contextText
}

func (agent *Agent) Chat(prompt string) (<-chan string, error) {
	agent.updateSystemMessage()
	agent.pruneMessages()

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

			response, err := agent.llmClient.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model:       agent.modelId,
					Messages:    agent.messages,
					Temperature: agent.temperature,
					Tools: []openai.Tool{
						tools.BashToolDefinition,
						tools.ViewFileToolDefinition,
						tools.ViewDirectoryToolDefinition,
						tools.CreateFileToolDefinition,
						tools.EditFileToolDefinition,
					},
					ParallelToolCalls: false,
				},
			)
			if err != nil {
				fmt.Print(gline.RESET_CURSOR_COLUMN + styles.ERROR(fmt.Sprintf("Error sending request to LLM: %s", err)) + "\n")
				agent.logger.Error("Error sending request to LLM", zap.Error(err))
				return
			}

			if len(response.Choices) == 0 {
				fmt.Print(gline.RESET_CURSOR_COLUMN + styles.ERROR("LLM responded with an empty response. This is typically a problem with the model being used. Please try again.") + "\n")
				agent.logger.Error("Error parsing LLM response", zap.String("response", fmt.Sprintf("%+v", response)))
				return
			}

			msg := response.Choices[0]
			agent.messages = append(agent.messages, msg.Message)
			agent.logger.Debug("LLM chat response", zap.Any("messages", agent.messages), zap.Any("response", msg))

			if msg.FinishReason == "stop" || msg.FinishReason == "end_turn" || msg.FinishReason == "tool_calls" || msg.FinishReason == "function_call" {
				if msg.Message.Content != "" {
					responseChannel <- strings.TrimSpace(msg.Message.Content)
				}

				if len(msg.Message.ToolCalls) > 0 {
					allToolCallsSucceeded := true
					for _, toolCall := range msg.Message.ToolCalls {
						if !agent.handleToolCall(toolCall) {
							allToolCallsSucceeded = false
						}
					}

					if allToolCallsSucceeded {
						continueSession = true
					}
				}
			} else if msg.FinishReason != "" {
				agent.logger.Warn("LLM chat response finished for unexpected reason", zap.String("reason", string(msg.FinishReason)))
			}
		}
	}()

	return responseChannel, nil
}

func (agent *Agent) handleToolCall(toolCall openai.ToolCall) bool {
	var params map[string]any
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
		agent.logger.Error(fmt.Sprintf("Failed to parse function call arguments: %v", err), zap.String("arguments", toolCall.Function.Arguments))
		fmt.Print(
			gline.RESET_CURSOR_COLUMN +
				styles.ERROR("LLM responded with something invalid. This is typically an indication that the model being used is not intelligent enough for the current task. Please try again.") +
				"\n",
		)
		return false
	}

	agent.logger.Debug("Handling tool call", zap.String("tool", toolCall.Function.Name), zap.Any("params", params))

	toolResponse := fmt.Sprintf("Unknown tool: %s", toolCall.Function.Name)

	switch toolCall.Function.Name {
	case tools.DoneToolDefinition.Function.Name:
		// done
		toolResponse = "ok"
	case tools.BashToolDefinition.Function.Name:
		// bash
		toolResponse = tools.BashTool(agent.runner, agent.historyManager, agent.logger, params)
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

func (agent *Agent) pruneMessages() {
	keptMessages := []openai.ChatCompletionMessage{}

	// This is a naive algorithm that assumes each llm token takes 4 bytes on average
	maxBytes := 4 * environment.GetAgentContextWindowTokens(agent.runner, agent.logger)

	usedBytes := 0
	for i := len(agent.messages) - 1; i > 0; i-- {
		bytes, err := agent.messages[i].MarshalJSON()
		if err != nil {
			agent.logger.Error("Failed to marshal message for pruning", zap.Error(err))
			break
		}

		length := len(bytes)
		if usedBytes+length > maxBytes {
			break
		}

		usedBytes += length
		keptMessages = append(
			[]openai.ChatCompletionMessage{agent.messages[i]},
			keptMessages...,
		)
	}

	agent.messages = append(
		[]openai.ChatCompletionMessage{agent.messages[0]},
		keptMessages...,
	)
}
