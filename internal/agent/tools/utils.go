package tools

import (
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

var userConfirmation = func(logger *zap.Logger, question string, explanation string) string {
	prompt :=
		styles.AGENT_QUESTION(question + " (y/N/freeform reply) ")

	line, err := gline.Gline(prompt, []string{}, explanation, nil, nil, nil, logger, gline.NewOptions())
	if err != nil {
		return "no"
	}

	// Handle Ctrl+C case: gline returns empty string when Ctrl+C is pressed
	if line == "" {
		logger.Debug("User pressed Ctrl+C (gline returned empty string), treating as 'n' response")
		return "n"
	}

	lowerLine := strings.ToLower(line)

	if lowerLine == "y" || lowerLine == "yes" {
		return "y"
	}

	if lowerLine == "n" || lowerLine == "no" {
		return "n"
	}

	return line
}
