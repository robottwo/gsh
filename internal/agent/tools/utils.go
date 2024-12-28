package tools

import (
	"fmt"

	"github.com/atinylittleshell/gsh/pkg/gline"
)

func failedToolResponse(errorMessage string) string {
	return fmt.Sprintf("<gsh_tool_call_error>%s</gsh_tool_call_error>", errorMessage)
}

func printToolMessage(message string) {
	fmt.Print(gline.LIGHT_YELLOW)
	fmt.Printf(message + "\n")
	fmt.Print(gline.RESET_COLOR)
}
