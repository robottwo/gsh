package tools

import (
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/pkg/gline"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

func failedToolResponse(errorMessage string) string {
	return fmt.Sprintf("<gsh_tool_call_error>%s</gsh_tool_call_error>", errorMessage)
}

func printToolMessage(message string) {
	fmt.Print(gline.LIGHT_YELLOW)
	fmt.Printf(message + "\n")
	fmt.Print(gline.RESET_COLOR)
}

func userConfirmation(runner *interp.Runner, logger *zap.Logger, question string, preview string) bool {
	prompt :=
		gline.LIGHT_YELLOW + question + gline.SAVE_CURSOR + "\n" +
			gline.RESET_COLOR + gline.RESET_CURSOR_COLUMN + preview +
			gline.RESTORE_CURSOR + gline.LIGHT_YELLOW + "(y/N) " + gline.RESET_COLOR

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
