package agent

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

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

	if len(agent.messages) == 0 {
		t.Fatalf("Expected some messages to be retained, but got none")
	}

	if agent.messages[0].Role != "system" {
		t.Errorf("Expected the first message to be 'system', got %s", agent.messages[0].Role)
	}

	if len(agent.messages) != 2 {
		t.Errorf("Expected pruned messages to be 2, got %d", len(agent.messages))
	}

	if agent.messages[1].Content != "Assistant message 2" {
		t.Errorf("Expected the second message to be 'Assistant message 2', got %s", agent.messages[1].Content)
	}

}
