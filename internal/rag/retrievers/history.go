package retrievers

import (
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/internal/history"
	"github.com/atinylittleshell/gsh/internal/rag"
	"mvdan.cc/sh/v3/interp"
)

type HistoryContextRetriever struct {
	Runner         *interp.Runner
	HistoryManager *history.HistoryManager
}

func (r HistoryContextRetriever) GetContext(options rag.ContextRetrievalOptions) (string, error) {
	historyEntries, err := r.HistoryManager.GetRecentEntries("", options.HistoryLimit)
	if err != nil {
		return "", err
	}

	var commandHistory string
	for _, entry := range historyEntries {
		if options.Concise {
			commandHistory += entry.Command + "\n"
		} else {
			commandHistory += fmt.Sprintf(`<entry>
  <time>%s</time>
  <dir>%s</dir>
  <cmd>%s</cmd>
  <exit_code>%d</exit_code>
</entry>
`,
				entry.CreatedAt.Format("2006-01-02T15:04:05"),
				entry.Directory,
				entry.Command,
				entry.ExitCode.Int32,
			)
		}
	}

	return fmt.Sprintf(`<recent_commands>
%s
</recent_commands>`, strings.TrimSpace(commandHistory)), nil
}
