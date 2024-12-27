package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/internal/utils"
	"github.com/atinylittleshell/gsh/pkg/gline"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

var doneToolDefinition = openai.Tool{
	Type: "function",
	Function: &openai.FunctionDefinition{
		Name:        "done",
		Description: `Confirm that the current user request is done.`,
	},
}

var bashToolDefinition = openai.Tool{
	Type: "function",
	Function: &openai.FunctionDefinition{
		Name: "bash",
		Description: `Run commands in a bash shell.
* When invoking this tool, the contents of the \"command\" parameter does NOT need to be XML-escaped.
* You don't have access to the internet via this tool.
* State is persistent across command calls and discussions with the user.
* To inspect a particular line range of a file, e.g. lines 10-25, try 'sed -n 10,25p /path/to/the/file'.`,
		Parameters: utils.GenerateJsonSchema(struct {
			Command string `json:"command" jsonschema_description:"The bash command to run" jsonschema_required:"true"`
		}{}),
	},
}

func bashTool(runner *interp.Runner, logger *zap.Logger, command string) (string, string, int, bool) {
	var prog *syntax.Stmt
	err := syntax.NewParser().Stmts(strings.NewReader(command), func(stmt *syntax.Stmt) bool {
		prog = stmt
		return false
	})
	if err != nil {
		logger.Error("LLM bash tool received invalid command", zap.Error(err))
		return "", fmt.Sprintf("`%s` is not a valid bash command: %s", command, err), -1, false
	}

	prompt :=
		gline.LIGHT_YELLOW + "gsh is requesting to run the following command: " + gline.SAVE_CURSOR + "\n" +
			gline.RESET_COLOR + gline.RESET_CURSOR_COLUMN + command +
			gline.RESTORE_CURSOR + gline.LIGHT_YELLOW + "(y/N) " + gline.RESET_COLOR

	options := gline.NewOptions()

	line, err := gline.NextLine(prompt, runner.Vars["PWD"].String(), nil, logger, *options)
	if err != nil {
		logger.Error("error reading user confirmation for bash tool", zap.Error(err))
		return "", fmt.Sprintf("Error reading user confirmation: %s", err), -1, false
	}

	fmt.Print(gline.CLEAR_AFTER_CURSOR)
	fmt.Println(command)

	if strings.ToLower(line) != "y" {
		return "", "User refused to run the command", -1, false
	}

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	multiOut := io.MultiWriter(os.Stdout, outBuf)
	multiErr := io.MultiWriter(os.Stderr, errBuf)

	childShell := runner.Subshell()
	interp.StdIO(os.Stdin, multiOut, multiErr)(childShell)

	err = childShell.Run(context.Background(), prog)
	if err != nil {
		exitCode, _ := interp.IsExitStatus(err)
		return outBuf.String(), errBuf.String(), int(exitCode), true
	} else {
		return outBuf.String(), errBuf.String(), 0, true
	}
}
