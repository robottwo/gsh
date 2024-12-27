package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/internal/agent"
	"github.com/atinylittleshell/gsh/internal/history"
	"github.com/atinylittleshell/gsh/pkg/gline"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

const (
	DEFAULT_PROMPT = "gsh> "
)

func RunInteractiveShell(runner *interp.Runner, historyManager *history.HistoryManager, logger *zap.Logger) error {
	predictor := NewLLMPredictor(runner, historyManager, logger)
	agent := agent.NewAgent(runner, logger)

	commandIndex := 0

	for {
		prompt := getPrompt(runner)
		logger.Debug("prompt updated", zap.String("prompt", prompt))

		options := gline.NewOptions()
		options.ClearScreen = commandIndex == 0

		// Read input
		line, err := gline.NextLine(prompt, runner.Vars["PWD"].String(), predictor, logger, *options)
		commandIndex++

		logger.Debug("received command", zap.String("line", line))

		if err != nil {
			logger.Error("error reading input through gline", zap.Error(err))
			return err
		}

		// Handle agent chat
		if strings.HasPrefix(line, "#") {
			chatMessage := line[1:]
			chatChannel, err := agent.Chat(chatMessage)
			if err != nil {
				logger.Error("error chatting with agent", zap.Error(err))
				continue
			}

			fmt.Print(gline.LIGHT_BLUE)
			for message := range chatChannel {
				fmt.Print(message)
			}
			fmt.Print(gline.RESET_COLOR)

			continue
		}

		// Execute the command
		shouldExit, err := executeCommand(line, historyManager, runner, logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		}

		if shouldExit {
			logger.Debug("exiting...")
			break
		}
	}

	// Clear screen on exit
	fmt.Print(gline.CLEAR_SCREEN)
	return nil
}

func executeCommand(input string, historyManager *history.HistoryManager, runner *interp.Runner, logger *zap.Logger) (bool, error) {
	var prog *syntax.Stmt
	err := syntax.NewParser().Stmts(strings.NewReader(input), func(stmt *syntax.Stmt) bool {
		prog = stmt
		return false
	})
	if err != nil {
		logger.Error("error parsing command", zap.Error(err))
		return false, err
	}

	historyEntry, _ := historyManager.StartCommand(input, runner.Vars["PWD"].String())

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	multiOut := io.MultiWriter(os.Stdout, outBuf)
	multiErr := io.MultiWriter(os.Stderr, errBuf)

	childShell := runner.Subshell()
	interp.StdIO(os.Stdin, multiOut, multiErr)(childShell)

	err = childShell.Run(context.Background(), prog)
	if err != nil {
		status, _ := interp.IsExitStatus(err)
		historyManager.FinishCommand(historyEntry, outBuf.String(), errBuf.String(), int(status))
	} else {
		historyManager.FinishCommand(historyEntry, outBuf.String(), errBuf.String(), 0)
	}

	return childShell.Exited(), nil
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
