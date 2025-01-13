package predict

import (
	"github.com/atinylittleshell/gsh/internal/utils"
)

type predictedCommand struct {
	Thought          string `json:"thought" description:"Your step by step thinking for what my intent might be" required:"true"`
	PredictedCommand string `json:"predicted_command" description:"The full bash command predicted by the model" required:"true"`
}

var PREDICTED_COMMAND_SCHEMA = utils.GenerateJsonSchema(predictedCommand{})

type explainedCommand struct {
	Explanation string `json:"explanation" description:"A concise explanation of what the command will do for me" required:"true"`
}

var EXPLAINED_COMMAND_SCHEMA = utils.GenerateJsonSchema(explainedCommand{})
