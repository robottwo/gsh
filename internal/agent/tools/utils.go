package tools

import (
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/pkg/gline"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

var LIGHT_YELLOW_BOLD = color.New(color.Bold, color.FgHiYellow).SprintFunc()
var WHITE = color.New(color.FgWhite).SprintFunc()

func failedToolResponse(errorMessage string) string {
	return fmt.Sprintf("<gsh_tool_call_error>%s</gsh_tool_call_error>", errorMessage)
}

func printToolMessage(message string) {
	LIGHT_YELLOW_BOLD(message + "\n")
}

func userConfirmation(runner *interp.Runner, logger *zap.Logger, question string, preview string) bool {
	prompt :=
		LIGHT_YELLOW_BOLD(question) + gline.SAVE_CURSOR + "\n" +
			gline.RESET_CURSOR_COLUMN + WHITE(preview) +
			gline.RESTORE_CURSOR + LIGHT_YELLOW_BOLD("(y/N) ")

	options := gline.NewOptions()

	line, err := gline.NextLine(prompt, runner.Vars["PWD"].String(), nil, logger, *options)
	if err != nil {
		return false
	}

	fmt.Print(gline.CLEAR_AFTER_CURSOR)
	fmt.Println(preview)

	if strings.ToLower(line) != "y" {
		return false
	}

	return true
}
