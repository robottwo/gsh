package tools

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

const LINES_TO_READ = 100

var ViewFileToolDefinition = openai.Tool{
	Type: "function",
	Function: &openai.FunctionDefinition{
		Name:        "view_file",
		Description: fmt.Sprintf(`View the content of a text file, at most %d lines at a time.`, LINES_TO_READ),
		Parameters: utils.GenerateJsonSchema(struct {
			Path      string `json:"path" jsonschema_description:"Absolute path to the file" jsonschema_required:"true"`
			StartLine int    `json:"start_line" jsonschema_description:"Optional. The line number to start viewing. This is zero indexed, inclusive. If not provided, we will read from the beginning of the file." jsonschema_required:"false"`
		}{}),
	},
}

func ViewFileTool(runner *interp.Runner, logger *zap.Logger, params map[string]any) string {
	path, ok := params["path"].(string)
	if !ok {
		logger.Error("The view_file tool failed to parse parameter 'path'")
		return failedToolResponse("The view_file tool failed to parse parameter 'path'")
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(environment.GetPwd(runner), path)
	}

	startLine := -1
	startLineVal, startLineExists := params["start_line"]
	if startLineExists {
		startLineFloat, ok := startLineVal.(float64)
		if !ok {
			logger.Error("The view_file tool failed to parse parameter 'start_line'")
			return failedToolResponse("The view_file tool failed to parse parameter 'start_line'")
		}
		startLine = int(startLineFloat)
	}

	endLine := startLine + LINES_TO_READ

	file, err := os.Open(path)
	if err != nil {
		logger.Error("view_file tool received invalid path", zap.Error(err))
		return failedToolResponse(fmt.Sprintf("Error opening file: %s", err))
	}
	defer file.Close()

	printToolMessage("gsh: I'm reading the following file:")
	fmt.Println(path)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		logger.Error("view_file tool received invalid path", zap.Error(err))
		return failedToolResponse(fmt.Sprintf("Error reading file: %s", err))
	}

	lines := strings.Split(buf.String(), "\n")
	if startLine < 0 {
		return failedToolResponse("start_line must be greater than or equal to 0")
	}
	if startLine > len(lines) {
		return failedToolResponse("start_line is greater than the number of lines in the file")
	}
	if endLine < 0 {
		return failedToolResponse("end_line must be greater than or equal to 0")
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if endLine <= startLine {
		return failedToolResponse("end_line must be greater than start_line")
	}

	result := strings.Join(lines[startLine:endLine], "\n")
	if len(result) > 32*1024 {
		return failedToolResponse("File content is too large. Please specify start_line and end_line to view a smaller subset of the file.")
	}

	return result
}
