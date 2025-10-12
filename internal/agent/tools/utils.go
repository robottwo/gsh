package tools

import (
	"flag"
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/internal/styles"
	"github.com/atinylittleshell/gsh/pkg/gline"
	"go.uber.org/zap"
)

func failedToolResponse(errorMessage string) string {
	return fmt.Sprintf("<gsh_tool_call_error>%s</gsh_tool_call_error>", errorMessage)
}

func printToolMessage(message string) {
	fmt.Print(gline.RESET_CURSOR_COLUMN + styles.AGENT_QUESTION(message) + "\n")
}

// defaultUserConfirmation is the default implementation that calls gline.Gline
var defaultUserConfirmation = func(logger *zap.Logger, question string, explanation string) string {
	prompt :=
		styles.AGENT_QUESTION(question + " (y/N/freeform/m) ")

	// Retry logic for transient errors
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		line, err := gline.Gline(prompt, []string{}, explanation, nil, nil, nil, logger, gline.NewOptions())
		if err != nil {
			// Check if the error is specifically from Ctrl+C interruption
			if err == gline.ErrInterrupted {
				logger.Debug("User pressed Ctrl+C, treating as 'n' response")
				return "n"
			}

			// Log the error for debugging
			logger.Warn("gline.Gline returned error during user confirmation",
				zap.Error(err),
				zap.String("question", question),
				zap.Int("attempt", attempt+1),
				zap.Int("maxRetries", maxRetries))

			// If this is not the last attempt, retry
			if attempt < maxRetries-1 {
				logger.Debug("Retrying gline.Gline due to error", zap.Int("nextAttempt", attempt+2))
				continue
			}

			// If all retries failed, log and return default "n" response
			logger.Error("All gline.Gline attempts failed, defaulting to 'n' response",
				zap.Error(err),
				zap.String("question", question))
			return "n"
		}

		// Success - process the input
		// Handle empty input as default "no" response
		if strings.TrimSpace(line) == "" {
			return "n"
		}

		lowerLine := strings.ToLower(line)

		if lowerLine == "y" || lowerLine == "yes" {
			return "y"
		}

		if lowerLine == "n" || lowerLine == "no" {
			return "n"
		}

		if lowerLine == "m" || lowerLine == "manage" {
			return "manage"
		}

		// Legacy support for "always"
		if lowerLine == "a" || lowerLine == "always" {
			return "always"
		}

		return line
	}

	// This should never be reached due to the loop structure, but included for safety
	logger.Error("Unexpected code path in userConfirmation")
	return "n"
}

// userConfirmation is a wrapper that checks for test mode before calling the real implementation
var userConfirmation = func(logger *zap.Logger, question string, explanation string) string {
	// Check if we're in test mode and this function hasn't been mocked
	// We detect if it's been mocked by checking if the function pointer has changed
	if flag.Lookup("test.v") != nil {
		// In test mode, return "n" to avoid blocking on gline.Gline
		// Tests that need different behavior should mock this function
		if logger != nil {
			logger.Debug("userConfirmation called in test mode without mock, returning 'n'")
		}
		return "n"
	}

	return defaultUserConfirmation(logger, question, explanation)
}
