package core

import (
	"fmt"
	"github.com/atinylittleshell/gsh/internal/terminal"
	"golang.org/x/term"
	"os"
	"os/exec"
	"strings"
)

const (
	EXIT_COMMAND = "exit"
)

// Shell represents the main shell structure
type Shell struct {
	Prompt      string
	History     []string
	OriginalTTY *term.State
}

// NewShell creates and initializes a new shell instance
func NewShell() (*Shell, error) {
	origTTY, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	return &Shell{
		Prompt:      "gsh> ", // Default prompt
		History:     []string{},
		OriginalTTY: origTTY,
	}, nil
}

// Run starts the main shell loop
func (s *Shell) Run() {
	defer s.restoreTerminal()

	fmt.Print(terminal.CLEAR_SCREEN)

	for {
		// Display prompt
		fmt.Print(s.Prompt)

		// Read user input
		input, err := s.readCommand()
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
		if err := s.executeCommand(input); err != nil {
			fmt.Fprintf(os.Stderr, "gsh: %s\n", err)
		}
		fmt.Print(terminal.RESET_CURSOR_COLUMN)

		// Save to history
		s.History = append(s.History, input)
	}

	// Clear screen on exit
	fmt.Print(terminal.CLEAR_SCREEN)
}

func (s *Shell) readCommand() (string, error) {
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
func (s *Shell) restoreTerminal() error {
	return term.Restore(int(os.Stdin.Fd()), s.OriginalTTY) // Restore terminal to original state
}

// executeCommand parses and executes a shell command
func (s *Shell) executeCommand(input string) error {
	// Split input into command and arguments
	args := strings.Split(input, " ")
	cmd := exec.Command(args[0], args[1:]...)

	// Restore terminal to canonical mode
	if err := s.restoreTerminal(); err != nil {
		return fmt.Errorf("failed to restore terminal mode - %w", err)
	}
	defer term.MakeRaw(int(os.Stdin.Fd())) // Re-enter raw mode after the command

	// Set command output to standard output/error
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	return cmd.Run()
}
