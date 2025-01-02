package retrievers

import (
	"fmt"

	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/rag"
	"mvdan.cc/sh/v3/interp"
)

type WorkingDirectoryContextRetriever struct {
	Runner *interp.Runner
}

func (r WorkingDirectoryContextRetriever) GetContext(options rag.ContextRetrievalOptions) (string, error) {
	return fmt.Sprintf("<working_dir>%s</working_dir>", environment.GetPwd(r.Runner)), nil
}
