package tools

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

const (
	MAX_DEPTH = 2
)

var ViewDirectoryToolDefinition = openai.Tool{
	Type: "function",
	Function: &openai.FunctionDefinition{
		Name:        "view_directory",
		Description: `View the content in a directory up to 2 levels deep.`,
		Parameters: utils.GenerateJsonSchema(struct {
			Path string `json:"path" jsonschema_description:"Absolute path to the directory" jsonschema_required:"true"`
		}{}),
	},
}

func ViewDirectoryTool(runner *interp.Runner, logger *zap.Logger, params map[string]any) string {
	path, ok := params["path"].(string)
	if !ok {
		logger.Error("The view_directory tool failed to parse parameter 'path'")
		return failedToolResponse("The view_directory tool failed to parse parameter 'path'")
	}

	var buf bytes.Buffer
	writer := io.StringWriter(&buf)

	printToolMessage(fmt.Sprintf("I'm viewing a directory: %s", path))

	walkDir(logger, writer, path, 1)

	return buf.String()
}

func walkDir(logger *zap.Logger, writer io.StringWriter, dir string, depth int) {
	if depth > MAX_DEPTH {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.Error("Error reading directory", zap.String("dir", dir), zap.Error(err))
		return
	}

	// Print each entry, and if it's a directory, recurse into it (unless at depth 2).
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		// Print the entry. You could also format the depth with indentation if you like.
		writer.WriteString(fullPath + "\n")

		// If it's a directory, recurse one level deeper.
		if entry.IsDir() && depth < MAX_DEPTH {
			walkDir(logger, writer, fullPath, depth+1)
		}
	}
}
