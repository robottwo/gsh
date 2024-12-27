package tools

import openai "github.com/sashabaranov/go-openai"

var DoneToolDefinition = openai.Tool{
	Type: "function",
	Function: &openai.FunctionDefinition{
		Name:        "done",
		Description: `Confirm that the current user request is done.`,
	},
}
