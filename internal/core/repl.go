package core

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atinylittleshell/gsh/internal/terminal"
	"github.com/atinylittleshell/gsh/pkg/debounce"
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
	prompt                          string
	history                         []string
	originalTTY                     *term.State
	runner                          *interp.Runner
	userInput                       string
	predictedInput                  string
	llmClient                       *openai.Client
	generatePredictedInputDebounced func()
}

// NewREPL creates and initializes a new repl instance
func NewREPL(runner *interp.Runner) (*REPL, error) {
	origTTY, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	repl := &REPL{
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
	}

	repl.generatePredictedInputDebounced = debounce.Debounce(200*time.Millisecond, func() {
		repl.generatePredictedInput()
	})

	return repl, nil
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
	fmt.Print(terminal.RESET_COLOR)

	// User input
	fmt.Print(terminal.WHITE)
	fmt.Print(repl.userInput)
	fmt.Print(terminal.RESET_COLOR)

	// Predicted input
	if len(repl.predictedInput) > 0 && strings.HasPrefix(repl.predictedInput, repl.userInput) {
		fmt.Print(terminal.SAVE_CURSOR)

		fmt.Print(terminal.GRAY)
		fmt.Print(repl.predictedInput[len(repl.userInput):])
		fmt.Print(terminal.RESET_COLOR)

		fmt.Print(terminal.RESTORE_CURSOR)
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

		if char == '\n' || char == '\r' {
			// Enter
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

	go repl.generatePredictedInputDebounced()
}

func (repl *REPL) generatePredictedInput() {
	userInput := repl.userInput

	predicted, err := predictInput(repl.llmClient, userInput)

	if repl.userInput != userInput {
		return
	}

	if err != nil {
		repl.predictedInput = ""
	} else {
		repl.predictedInput = predicted
	}

	repl.redrawLine()
}

// cleanup restores the original terminal state
func (repl *REPL) restoreTerminal() error {
	return term.Restore(int(os.Stdin.Fd()), repl.originalTTY)
}

// executeCommand parses and executes a shell command
func (repl *REPL) executeCommand(input string) error {
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
