package tools

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

var ViewFileToolDefinition = openai.Tool{
	Type: "function",
	Function: &openai.FunctionDefinition{
		Name:        "view_file",
		Description: `View the content of a text file.`,
		Parameters: utils.GenerateJsonSchema(struct {
			Path      string `json:"path" jsonschema_description:"Absolute path to the file" jsonschema_required:"true"`
			StartLine int    `json:"start_line" jsonschema_description:"Optional. The line number to start viewing. This is zero indexed, inclusive. If not provided, we will read from the beginning of the file." jsonschema_required:"false"`
			EndLine   int    `json:"end_line" jsonschema_description:"Optional. The line number to stop viewing. This is zero indexed, exclusive. If not provided, we will read till the end of the file." jsonschema_required:"false"`
		}{}),
	},
}

func ViewFileTool(runner *interp.Runner, logger *zap.Logger, params map[string]any) string {
	path, ok := params["path"].(string)
	if !ok {
		logger.Error("The view_file tool failed to parse parameter 'path'")
		return failedToolResponse("The view_file tool failed to parse parameter 'path'")
	}

	startLine := -1
	startLineVal, startLineExists := params["startLine"]
	if startLineExists {
		startLine, ok = startLineVal.(int)
		if !ok {
			logger.Error("The view_file tool failed to parse parameter 'startLine'")
			return failedToolResponse("The view_file tool failed to parse parameter 'startLine'")
		}
	}

	endLine := -1
	endLineVal, endLineExists := params["endLine"]
	if endLineExists {
		endLine, ok = endLineVal.(int)
		if !ok {
			logger.Error("The view_file tool failed to parse parameter 'endLine'")
			return failedToolResponse("The view_file tool failed to parse parameter 'endLine'")
		}
	}

	file, err := os.Open(path)
	if err != nil {
		logger.Error("view_file tool received invalid path", zap.Error(err))
		return failedToolResponse(fmt.Sprintf("Error opening file: %s", err))
	}
	defer file.Close()

	printToolMessage(fmt.Sprintf("I'm reading a file: %s", path))

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		logger.Error("view_file tool received invalid path", zap.Error(err))
		return failedToolResponse(fmt.Sprintf("Error reading file: %s", err))
	}

	lines := strings.Split(buf.String(), "\n")
	if startLine < 0 {
		startLine = 0
	}
	if startLine > len(lines) {
		startLine = len(lines)
	}
	if endLine < 0 {
		endLine = len(lines)
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if endLine <= startLine {
		return failedToolResponse("EndLine must be greater than StartLine")
	}

	result := strings.Join(lines[startLine:endLine], "\n")
	if len(result) > 32*1024 {
		return failedToolResponse("File content is too large. Please specify start_line and end_line to view a smaller subset of the file.")
	}

	return result
}
