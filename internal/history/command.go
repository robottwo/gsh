package history

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/interp"
)

func NewHistoryCommandHandler(historyManager *HistoryManager) func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	return func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
		return func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return next(ctx, args)
			}

			if args[0] != "history" {
				return next(ctx, args)
			}

			// Parse flags and arguments
			if len(args) > 1 {
				switch args[1] {
				case "-c", "--clear":
					// Clear the history
					return historyManager.ResetHistory()

				case "-d", "--delete":
					// Delete a specific entry
					if len(args) < 3 {
						return fmt.Errorf("history -d requires an entry number")
					}
					offset, err := strconv.Atoi(args[2])
					if err != nil {
						return fmt.Errorf("invalid history entry number: %s", args[2])
					}
					id := uint(offset)
					if err := historyManager.DeleteEntry(id); err != nil {
						return fmt.Errorf("failed to delete history entry %d: %v", id, err)
					}
					return nil

				case "-h", "--help":
					printHistoryHelp()
					return nil
				}
			}

			// Default limit is 20 entries, or use provided number
			limit := 20
			if len(args) > 1 {
				providedLimit, err := strconv.Atoi(args[1])
				if err == nil && providedLimit > 0 {
					limit = providedLimit
				}
			}

			// Get recent entries
			entries, err := historyManager.GetRecentEntries("", limit)
			if err != nil {
				return err
			}

			// Print entries
			for _, entry := range entries {
				fmt.Printf("%d %s\n", entry.ID, entry.Command)
			}

			return nil
		}
	}
}

func printHistoryHelp() {
	help := []string{
		"Usage: history [option] [n]",
		"Display or manipulate the history list.",
		"",
		"Options:",
		"  -c, --clear    clear the history list",
		"  -d, --delete   delete history entry at offset",
		"  -h, --help     display this help message",
		"",
		"If n is given, display only the last n entries.",
		"If no options are given, display the history list with line numbers.",
	}
	fmt.Println(strings.Join(help, "\n"))
}
