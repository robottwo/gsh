package predict

import (
	"github.com/atinylittleshell/gsh/internal/utils"
)

type PredictedCommand struct {
	PredictedCommand string `json:"predicted_command" description:"The full bash command predicted by the model" required:"true"`
}

var PREDICTED_COMMAND_SCHEMA = utils.GenerateJsonSchema(PredictedCommand{})

type explainedCommand struct {
	Explanation string `json:"explanation" description:"A concise explanation of what the command will do for me" required:"true"`
}

var EXPLAINED_COMMAND_SCHEMA = utils.GenerateJsonSchema(explainedCommand{})
