package tools

import (
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/pkg/gline"
	"github.com/fatih/color"
	"go.uber.org/zap"
)

var LIGHT_YELLOW_BOLD = color.New(color.Bold, color.FgHiYellow).SprintFunc()
var WHITE = color.New(color.FgWhite).SprintFunc()

func failedToolResponse(errorMessage string) string {
	return fmt.Sprintf("<gsh_tool_call_error>%s</gsh_tool_call_error>", errorMessage)
}

func printToolMessage(message string) {
	fmt.Println(LIGHT_YELLOW_BOLD(message))
}

func userConfirmation(logger *zap.Logger, question string, preview string) string {
	prompt :=
		LIGHT_YELLOW_BOLD(question + "(y/N/[freeform feedback]) ")

	options := gline.NewOptions()

	line, err := gline.NextLine(prompt, preview, nil, logger, *options)
	if err != nil {
		return "no"
	}

	fmt.Print(gline.CLEAR_AFTER_CURSOR)
	fmt.Println(preview)

	lowerLine := strings.ToLower(line)

	if lowerLine == "y" || lowerLine == "yes" {
		return "y"
	}

	if lowerLine == "n" || lowerLine == "no" {
		return "n"
	}

	return line
}
