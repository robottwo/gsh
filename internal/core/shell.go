package core

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/internal/agent"
	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/history"
	"github.com/atinylittleshell/gsh/internal/predict"
	"github.com/atinylittleshell/gsh/internal/rag"
	"github.com/atinylittleshell/gsh/internal/rag/retrievers"
	"github.com/atinylittleshell/gsh/internal/styles"
	"github.com/atinylittleshell/gsh/pkg/gline"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func RunInteractiveShell(runner *interp.Runner, historyManager *history.HistoryManager, logger *zap.Logger) error {
	contextProvider := &rag.ContextProvider{
		Logger: logger,
		Retrievers: []rag.ContextRetriever{
			retrievers.SystemInfoContextRetriever{Runner: runner},
			retrievers.WorkingDirectoryContextRetriever{Runner: runner},
			retrievers.GitContextRetriever{Runner: runner},
			retrievers.HistoryContextRetriever{Runner: runner, HistoryManager: historyManager},
		},
	}
	predictor := &predict.PredictRouter{
		PrefixPredictor:    predict.NewLLMPrefixPredictor(runner, contextProvider, logger),
		NullStatePredictor: predict.NewLLMNullStatePredictor(runner, contextProvider, logger),
	}
	explainer := predict.NewLLMExplainer(runner, contextProvider, logger)
	agent := agent.NewAgent(runner, historyManager, logger)

	for {
		prompt := environment.GetPrompt(runner, logger)
		logger.Debug("prompt updated", zap.String("prompt", prompt))

		// Read input
		options := gline.NewOptions()
		options.MinHeight = environment.GetMinimumLines(runner, logger)

		line, err := gline.Gline(prompt, "", predictor, explainer, logger, options)

		logger.Debug("received command", zap.String("line", line))

		if err != nil {
			logger.Error("error reading input through gline", zap.Error(err))
			return err
		}

		// Handle agent chat
		if strings.HasPrefix(line, "#") {
			chatMessage := fmt.Sprintf(
				"%s\n\nContext:\n%s",
				line[1:],
				contextProvider.GetContext(
					rag.ContextRetrievalOptions{
						Concise:      false,
						HistoryLimit: environment.GetHistoryContextLimit(runner, logger),
					},
				),
			)
			chatChannel, err := agent.Chat(chatMessage)
			if err != nil {
				logger.Error("error chatting with agent", zap.Error(err))
				continue
			}

			for message := range chatChannel {
				fmt.Println(styles.AGENT_MESSAGE("gsh: " + message))
			}

			continue
		}

		// Handle empty input
		if strings.TrimSpace(line) == "" {
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

	return nil
}

func executeCommand(input string, historyManager *history.HistoryManager, runner *interp.Runner, logger *zap.Logger) (bool, error) {
	var prog *syntax.Stmt
	err := syntax.NewParser().Stmts(strings.NewReader(input), func(stmt *syntax.Stmt) bool {
		prog = stmt
		return false
	})
	if prog == nil {
		logger.Error("invalid command", zap.String("command", input))
		return false, nil
	}
	if err != nil {
		logger.Error("error parsing command", zap.String("command", input), zap.Error(err))
		return false, err
	}

	historyEntry, _ := historyManager.StartCommand(input, environment.GetPwd(runner))

	err = runner.Run(context.Background(), prog)
	if err != nil {
		var exitCode int
		status, ok := interp.IsExitStatus(err)
		if !ok {
			exitCode = -1
		} else {
			exitCode = int(status)
		}
		historyManager.FinishCommand(historyEntry, exitCode)
	} else {
		historyManager.FinishCommand(historyEntry, 0)
	}

	return runner.Exited(), nil
}
