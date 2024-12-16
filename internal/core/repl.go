package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/internal/terminal"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

const (
	EXIT_COMMAND = "exit"
)

// REPL represents the an interactive shell session
type REPL struct {
	prompt         string
	history        []string
	originalTTY    *term.State
	runner         *interp.Runner
	userInput      string
	predictedInput string
	llmClient      *openai.Client
}

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

type PredictedCommand struct {
	FullCommand string `json:"full_command" jsonschema_description:"The full bash command predicted by the model" jsonschema_required:"true"`
}

var PREDICTED_COMMAND_SCHEMA = GenerateSchema[PredictedCommand]()

var PREDICTED_COMMAND_SCHEMA_PARAM = openai.ResponseFormatJSONSchemaJSONSchemaParam{
	Name:        openai.F("predicted_command"),
	Description: openai.F("The predicted bash command"),
	Schema:      openai.F(PREDICTED_COMMAND_SCHEMA),
	Strict:      openai.Bool(true),
}

// NewREPL creates and initializes a new repl instance
func NewREPL(runner *interp.Runner) (*REPL, error) {
	origTTY, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	return &REPL{
		prompt:         "gsh> ", // Default prompt
		history:        []string{},
		originalTTY:    origTTY,
		runner:         runner,
		userInput:      "",
		predictedInput: "",
		llmClient: openai.NewClient(
			option.WithAPIKey("ollama"),
			option.WithBaseURL("http://localhost:11434/v1/"),
		),
	}, nil
}

// Run starts the main repl loop
func (repl *REPL) Run() error {
	defer repl.restoreTerminal()

	fmt.Print(terminal.CLEAR_SCREEN)

	for {
		repl.redrawLine()

		// Read user input
		input, err := repl.readCommand()
		if err != nil {
			fmt.Fprintln(os.Stderr, "gsh: error reading input - ", err)
			continue
		}

		// Exit condition
		if input == EXIT_COMMAND {
			break
		}

		// Execute command
		if err := repl.executeCommand(input); err != nil {
			fmt.Fprintf(os.Stderr, "gsh: %s\n", err)
		}

		// Save to history
		repl.history = append(repl.history, input)
	}

	// Clear screen on exit
	fmt.Print(terminal.CLEAR_SCREEN)

	return nil
}

func (repl *REPL) redrawLine() error {
	fmt.Print(terminal.CLEAR_LINE)

	// Prompt
	fmt.Print(terminal.WHITE)
	fmt.Print(repl.prompt)
	fmt.Print(terminal.RESET)

	// User input
	fmt.Print(terminal.WHITE)
	fmt.Print(repl.userInput)
	fmt.Print(terminal.RESET)

	// Predicted input
	if len(repl.predictedInput) > 0 && strings.HasPrefix(repl.predictedInput, repl.userInput) {
		suffix := repl.predictedInput[len(repl.userInput):]
		suffix_length := len(suffix)

		fmt.Print(terminal.GRAY)
		fmt.Print(suffix)
		fmt.Print(terminal.RESET)

		fmt.Print(terminal.ESC + fmt.Sprintf("[%dD", suffix_length))
	}

	return nil
}

func (repl *REPL) readCommand() (string, error) {
	repl.userInput = ""
	repl.predictedInput = ""

	buffer := make([]byte, 1)

	for {
		_, err := os.Stdin.Read(buffer)
		if err != nil {
			return "", err
		}

		char := buffer[0]

		// Enter
		if char == '\n' || char == '\r' {
			fmt.Println()
			fmt.Print(terminal.RESET_CURSOR_COLUMN)

			result := strings.TrimSpace(repl.userInput)

			repl.userInput = ""
			repl.predictedInput = ""

			return result, nil
		}

		if char == 127 {
			// Backspace
			if len(repl.userInput) > 0 {
				repl.userInput = repl.userInput[:len(repl.userInput)-1]
			}
		} else {
			// Normal character
			repl.userInput += string(char)
		}

		// Predict input
		repl.predictInput()

		repl.redrawLine()

	}
}

func (repl *REPL) predictInput() {
	if len(repl.userInput) == 0 {
		repl.predictedInput = ""
		return
	}

	if strings.HasPrefix(repl.predictedInput, repl.userInput) {
		return
	}

	go repl.generatePredictedInput(repl.userInput)
}

func (repl *REPL) generatePredictedInput(userInput string) {
	chatCompletion, err := repl.llmClient.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf(`
You are gsh, an intelligent shell program.
You are asked to predict a complete bash command based on a partial one from the user.

<partial_command>
%s
</partial_command>`, userInput)),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(PREDICTED_COMMAND_SCHEMA_PARAM),
			},
		),
		Model: openai.F("qwen2.5"),
	})

	if err != nil {
		panic(err)
	}

	predictedCommand := PredictedCommand{}
	_ = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &predictedCommand)

	if repl.userInput != userInput {
		return
	}

	repl.predictedInput = predictedCommand.FullCommand
	repl.redrawLine()
}

// cleanup restores the original terminal state
func (repl *REPL) restoreTerminal() error {
	return term.Restore(int(os.Stdin.Fd()), repl.originalTTY) // Restore terminal to original state
}

// executeCommand parses and executes a shell command
func (repl *REPL) executeCommand(input string) error {
	// Restore terminal to canonical mode
	if err := repl.restoreTerminal(); err != nil {
		return err
	}
	defer term.MakeRaw(int(os.Stdin.Fd())) // Re-enter raw mode after the command
	defer fmt.Print(terminal.RESET_CURSOR_COLUMN)

	prog, err := syntax.NewParser().Parse(strings.NewReader(input), "")
	if err != nil {
		return err
	}
	return repl.runner.Run(context.Background(), prog)
}
