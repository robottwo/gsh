package predict

import (
	"github.com/atinylittleshell/gsh/internal/utils"
)

type predictedCommand struct {
	Thought          string `json:"thought" jsonschema_description:"Your step by step thinking for what my intent might be" jsonschema_required:"true"`
	PredictedCommand string `json:"predicted_command" jsonschema_description:"The full bash command predicted by the model" jsonschema_required:"true"`
}

var PREDICTED_COMMAND_SCHEMA = utils.GenerateJsonSchema(predictedCommand{})

type explainedCommand struct {
	Explanation string `json:"explanation" jsonschema_description:"A concise explanation of what the command will do for me" jsonschema_required:"true"`
}

var EXPLAINED_COMMAND_SCHEMA = utils.GenerateJsonSchema(explainedCommand{})
