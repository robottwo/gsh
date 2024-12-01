package core

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/internal/terminal"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

const (
	EXIT_COMMAND = "exit"
)

// REPL represents the an interactive shell session
type REPL struct {
	Prompt      string
	History     []string
	OriginalTTY *term.State
	Runner      *interp.Runner
}

// NewREPL creates and initializes a new repl instance
func NewREPL(runner *interp.Runner) (*REPL, error) {
	origTTY, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	return &REPL{
		Prompt:      "gsh> ", // Default prompt
		History:     []string{},
		OriginalTTY: origTTY,
		Runner:      runner,
	}, nil
}

// Run starts the main repl loop
func (repl *REPL) Run() error {
	defer repl.restoreTerminal()

	fmt.Print(terminal.CLEAR_SCREEN)

	for {
		// Display prompt
		fmt.Print(repl.Prompt)

		// Read user input
		input, err := repl.readCommand()
		if err != nil {
			fmt.Fprintln(os.Stderr, "gsh: error reading input - ", err)
			continue
		}

		// Process input
		input = strings.TrimSpace(input)

		// Exit condition
		if input == EXIT_COMMAND {
			break
		}

		// Execute command
		if err := repl.executeCommand(input); err != nil {
			fmt.Fprintf(os.Stderr, "gsh: %s\n", err)
		}
		fmt.Print(terminal.RESET_CURSOR_COLUMN)

		// Save to history
		repl.History = append(repl.History, input)
	}

	// Clear screen on exit
	fmt.Print(terminal.CLEAR_SCREEN)

	return nil
}

func (repl *REPL) readCommand() (string, error) {
	var input []byte = []byte{}
	buffer := make([]byte, 1)

	for {
		_, err := os.Stdin.Read(buffer)
		if err != nil {
			return "", err
		}

		char := buffer[0]

		// Backspace
		if char == 127 {
			if len(input) > 0 {
				input = input[:len(input)-1]
				fmt.Print(terminal.BACKSPACE) // Clear character from terminal
			}
			continue
		}

		// Enter
		if char == '\n' || char == '\r' {
			fmt.Println()
			fmt.Print(terminal.RESET_CURSOR_COLUMN)
			break
		}

		// Normal character
		input = append(input, char)
		fmt.Print(string(buffer))
	}

	return string(input), nil
}

// cleanup restores the original terminal state
func (repl *REPL) restoreTerminal() error {
	return term.Restore(int(os.Stdin.Fd()), repl.OriginalTTY) // Restore terminal to original state
}

// executeCommand parses and executes a shell command
func (repl *REPL) executeCommand(input string) error {
	// Restore terminal to canonical mode
	if err := repl.restoreTerminal(); err != nil {
		return err
	}
	defer term.MakeRaw(int(os.Stdin.Fd())) // Re-enter raw mode after the command

	prog, err := syntax.NewParser().Parse(strings.NewReader(input), "")
	if err != nil {
		return err
	}
	return repl.Runner.Run(context.Background(), prog)
}
