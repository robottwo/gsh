package retrievers

import (
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/history"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type ConciseHistoryContextRetriever struct {
	Runner         *interp.Runner
	Logger         *zap.Logger
	HistoryManager *history.HistoryManager
}

type VerboseHistoryContextRetriever struct {
	Runner         *interp.Runner
	Logger         *zap.Logger
	HistoryManager *history.HistoryManager
}

func (r ConciseHistoryContextRetriever) Name() string {
	return "history_concise"
}

func (r VerboseHistoryContextRetriever) Name() string {
	return "history_verbose"
}

func (r ConciseHistoryContextRetriever) GetContext() (string, error) {
	historyEntries, err := r.HistoryManager.GetRecentEntries("", environment.GetContextNumHistoryConcise(r.Runner, r.Logger))
	if err != nil {
		return "", err
	}

	var commandHistory string
	for _, entry := range historyEntries {
		commandHistory += entry.Command + "\n"
	}

	return fmt.Sprintf(`<recent_commands>
%s
</recent_commands>`, strings.TrimSpace(commandHistory)), nil
}

func (r VerboseHistoryContextRetriever) GetContext() (string, error) {
	historyEntries, err := r.HistoryManager.GetRecentEntries("", environment.GetContextNumHistoryVerbose(r.Runner, r.Logger))
	if err != nil {
		return "", err
	}

	var commandHistory string = "#sequence,directory,exit_code,command\n"
	for _, entry := range historyEntries {
		commandHistory += fmt.Sprintf("%d,%s,%d,%s\n",
			entry.ID,
			entry.Directory,
			entry.ExitCode.Int32,
			entry.Command,
		)
	}

	return fmt.Sprintf(`<recent_commands>
%s
</recent_commands>`, strings.TrimSpace(commandHistory)), nil
}
