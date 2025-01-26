package agent

import (
	"github.com/stretchr/testify/assert"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

func TestResetChat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	runner, _ := interp.New(
		interp.StdIO(nil, nil, nil),
	)

	agent := &Agent{
		runner: runner,
		logger: logger,
		messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: "Old system message"},
			{Role: "user", Content: "User message 1"},
			{Role: "assistant", Content: "Assistant message 1"},
			{Role: "user", Content: "User message 2"},
			{Role: "assistant", Content: "Assistant message 2"},
		},
	}

	// Reset the chat
	agent.ResetChat()

	// Verify that only one system message remains
	assert.Len(t, agent.messages, 1, "Expected only one message after reset")
	assert.Equal(t, "system", agent.messages[0].Role, "Expected the remaining message to be 'system'")

	// Verify that the system message contains the latest context
	assert.Contains(t, agent.messages[0].Content, "You are gsh", "Expected system message to contain the latest context")
}

func TestPruneMessages(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	runner, _ := interp.New(
		interp.StdIO(nil, nil, nil),
	)

	runner.Vars = map[string]expand.Variable{
		"GSH_AGENT_CONTEXT_WINDOW_TOKENS": {Kind: expand.String, Str: "20"},
	}

	agent := &Agent{
		runner: runner,
		logger: logger,
		messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: "System message"},
			{Role: "user", Content: "User message 1"},
			{Role: "assistant", Content: "Assistant message 1"},
			{Role: "user", Content: "User message 2"},
			{Role: "assistant", Content: "Assistant message 2"},
		},
	}

	agent.pruneMessages()

	assert.NotEmpty(t, agent.messages, "Expected some messages to be retained, but got none")
	assert.Equal(t, "system", agent.messages[0].Role, "Expected the first message to be 'system'")
	assert.Len(t, agent.messages, 2, "Expected pruned messages to be 2")
	assert.Equal(t, "Assistant message 2", agent.messages[1].Content, "Expected the second message to be 'Assistant message 2'")

}
