package analytics

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/interp"
)

func NewAnalyticsCommandHandler(analyticsManager *AnalyticsManager) func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	return func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
		return func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return next(ctx, args)
			}

			if args[0] != "gsh_analytics" {
				return next(ctx, args)
			}

			// Parse flags and arguments
			if len(args) > 1 {
				switch args[1] {
				case "-c", "--clear":
					// Clear the analytics
					return analyticsManager.ResetAnalytics()

				case "-d", "--delete":
					// Delete a specific entry
					if len(args) < 3 {
						return fmt.Errorf("analytics -d requires an entry number")
					}
					offset, err := strconv.Atoi(args[2])
					if err != nil {
						return fmt.Errorf("invalid analytics entry number: %s", args[2])
					}
					id := uint(offset)
					if err := analyticsManager.DeleteEntry(id); err != nil {
						return fmt.Errorf("failed to delete analytics entry %d: %v", id, err)
					}
					return nil

				case "-h", "--help":
					printAnalyticsHelp()
					return nil

				case "-n", "--count":
					// Show total count of entries
					count, err := analyticsManager.GetTotalCount()
					if err != nil {
						return fmt.Errorf("failed to get analytics count: %v", err)
					}
					fmt.Printf("Total analytics entries: %d\n", count)
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
			entries, err := analyticsManager.GetRecentEntries(limit)
			if err != nil {
				return err
			}

			// Print entries
			for _, entry := range entries {
				fmt.Printf("%d [%s] Input: %s, Prediction: %s, Actual: %s\n",
					entry.ID,
					entry.CreatedAt.Format("2006-01-02 15:04:05"),
					entry.Input,
					entry.Prediction,
					entry.Actual)
			}

			return nil
		}
	}
}

func printAnalyticsHelp() {
	help := []string{
		"Usage: gsh_analytics [option] [n]",
		"Display or manipulate the analytics data.",
		"",
		"Options:",
		"  -c, --clear    clear all analytics data",
		"  -d, --delete   delete analytics entry at offset",
		"  -h, --help     display this help message",
		"  -n, --count    display total number of entries",
		"",
		"If n is given, display only the last n entries.",
		"If no options are given, display the analytics list with line numbers.",
	}
	fmt.Println(strings.Join(help, "\n"))
}
