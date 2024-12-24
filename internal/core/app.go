package core

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/internal/history"
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
	historyManager, err := history.NewHistoryManager(HistoryFile(), logger)
	if err != nil {
		return err
	}

	predictor := NewLLMPredictor(historyManager, logger)

	commandIndex := 0

	for {
		prompt := getPrompt(runner)
		logger.Debug("prompt updated", zap.String("prompt", prompt))

		// Read input
		line, err := gline.NextLine(prompt, runner.Vars["PWD"].String(), predictor, logger, gline.Options{
			ClearScreen: commandIndex == 0,
		})
		commandIndex++

		logger.Debug("received command", zap.String("line", line))

		if err != nil {
			logger.Error("error reading input through gline", zap.Error(err))
			return err
		}

		// Execute the command
		err = executeCommand(line, historyManager, runner, logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		}

		if runner.Exited() {
			logger.Debug("exiting...")
			break
		}
	}

	// Clear screen on exit
	fmt.Print(gline.CLEAR_SCREEN)
	return nil
}

func executeCommand(input string, historyManager *history.HistoryManager, runner *interp.Runner, logger *zap.Logger) error {
	var prog *syntax.Stmt
	err := syntax.NewParser().Stmts(strings.NewReader(input), func(stmt *syntax.Stmt) bool {
		prog = stmt
		return false
	})
	if err != nil {
		logger.Error("error parsing command", zap.Error(err))
		return err
	}

	historyEntry, _ := historyManager.StartCommand(input, runner.Vars["PWD"].String())

	err = runner.Run(context.Background(), prog)
	if err != nil {
		status, _ := interp.IsExitStatus(err)
		historyManager.FinishCommand(historyEntry, "", "", int(status))
	} else {
		historyManager.FinishCommand(historyEntry, "", "", 0)
	}

	return nil
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
