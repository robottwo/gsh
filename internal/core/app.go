package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atinylittleshell/gsh/internal/bash"
	"github.com/atinylittleshell/gsh/pkg/gline"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

const (
	EXIT_COMMAND = "exit"

	DEFAULT_PROMPT = "gsh> "
)

func RunApp(runner *interp.Runner, logger *zap.Logger) error {
	err := loadShellConfigs(runner, logger)
	if err != nil {
		logger.Error("failed to load gshrc", zap.Error(err))
		return err
	}

	predictor := NewLLMPredictor()

	commandIndex := 0

	for {
		prompt := getPrompt(runner)
		logger.Debug("prompt updated", zap.String("prompt", prompt))

		// Read input
		line, err := gline.NextLine(prompt, predictor, logger, gline.Options{
			ClearScreen: commandIndex == 0,
		})
		commandIndex++

		logger.Debug("received command", zap.String("line", line))

		if err != nil {
			logger.Error("error reading input through gline", zap.Error(err))
			return err
		}

		// Exit if the user types the exit command
		if line == EXIT_COMMAND {
			logger.Info("exiting")
			break
		}

		// Execute the command
		executeCommand(line, runner, logger)
	}

	// Clear screen on exit
	fmt.Print(gline.CLEAR_SCREEN)
	return nil
}

func executeCommand(input string, runner *interp.Runner, logger *zap.Logger) error {
	prog, err := syntax.NewParser().Parse(strings.NewReader(input), "")
	if err != nil {
		logger.Error("error parsing command", zap.Error(err))
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
func loadShellConfigs(runner *interp.Runner, logger *zap.Logger) error {
	configFiles := []string{
		filepath.Join(HomeDir(), ".gshrc"),
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			if err := bash.RunBashScript(runner, configFile); err != nil {
				logger.Error("error loading config file", zap.String("configFile", configFile), zap.Error(err))
			}
		}
	}

	return nil
}
