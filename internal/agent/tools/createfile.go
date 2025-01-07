package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/utils"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

var CreateFileToolDefinition = openai.Tool{
	Type: "function",
	Function: &openai.FunctionDefinition{
		Name:        "create_file",
		Description: `Create a file with the specified content.`,
		Parameters: utils.GenerateJsonSchema(struct {
			Path    string `json:"path" jsonschema_description:"Absolute path to the file" jsonschema_required:"true"`
			Content string `json:"content" jsonschema_description:"The content to write to the file" jsonschema_required:"true"`
		}{}),
	},
}

func CreateFileTool(runner *interp.Runner, logger *zap.Logger, params map[string]any) string {
	path, ok := params["path"].(string)
	if !ok {
		logger.Error("The create_file tool failed to parse parameter 'path'")
		return failedToolResponse("The create_file tool failed to parse parameter 'path'")
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(environment.GetPwd(runner), path)
	}

	content, ok := params["content"].(string)
	if !ok {
		logger.Error("The create_file tool failed to parse parameter 'content'")
		return failedToolResponse("The create_file tool failed to parse parameter 'content'")
	}

	confirmResponse := userConfirmation(
		logger,
		"gsh: Do I have your permission to create the following file?",
		fmt.Sprintf("%s\n\n%s", path, content),
	)
	if confirmResponse == "n" {
		return failedToolResponse("User declined this request")
	} else if confirmResponse != "y" {
		return failedToolResponse(fmt.Sprintf("User declined this request: %s", confirmResponse))
	}

	fmt.Println(path)

	file, err := os.Create(path)
	defer file.Close()

	if err != nil {
		logger.Error("create_file tool failed to create file", zap.Error(err))
		return failedToolResponse(fmt.Sprintf("Error creating file: %s", err))
	}

	_, err = file.WriteString(content)
	if err != nil {
		logger.Error("create_file tool received invalid content", zap.Error(err))
		return failedToolResponse(fmt.Sprintf("Error writing to file: %s", err))
	}

	return fmt.Sprintf("File successfully created at %s", path)
}
