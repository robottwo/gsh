package core

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/atinylittleshell/gsh/internal/bash"
	"github.com/atinylittleshell/gsh/pkg/gline"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

const (
	EXIT_COMMAND = "exit"

	DEFAULT_PROMPT = "gsh> "
)

func RunApp(runner *interp.Runner) error {
	loadShellConfigs(runner)

	predictor := NewLLMPredictor()

	commandIndex := 0

	for {
		prompt := getPrompt(runner)

		// Read input
		line, err := gline.NextLine(prompt, predictor, gline.Options{
			ClearScreen: false,
		})
		commandIndex++

		if err != nil {
			return err
		}

		// Exit if the user types the exit command
		if line == EXIT_COMMAND {
			break
		}

		// Execute the command
		executeCommand(line, runner)
	}

	// Clear screen on exit
	fmt.Print(gline.CLEAR_SCREEN)
	return nil
}

func executeCommand(input string, runner *interp.Runner) error {
	prog, err := syntax.NewParser().Parse(strings.NewReader(input), "")
	if err != nil {
		return err
	}
	return runner.Run(context.Background(), prog)
}

func getPrompt(runner *interp.Runner) string {
	promptUpdater := runner.Funcs["GSH_UPDATE_PROMPT"]
	if promptUpdater != nil {
		runner.Run(context.Background(), promptUpdater)
	}

	prompt := runner.Vars["GSH_PROMPT"].String()
	if prompt != "" {
		return prompt
	}

	return DEFAULT_PROMPT
}

// loadShellConfigs loads and executes .gshrc
func loadShellConfigs(runner *interp.Runner) error {
	user, err := user.Current()
	if err != nil {
		return fmt.Errorf("unable to determine current user: %w", err)
	}

	homeDir := user.HomeDir
	configFiles := []string{
		filepath.Join(homeDir, ".gshrc"),
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			if err := bash.RunBashScript(runner, configFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", configFile, err)
			}
		}
	}

	return nil
}
