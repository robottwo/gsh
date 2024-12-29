package retrievers

import (
	"fmt"

	"mvdan.cc/sh/v3/interp"
)

type WorkingDirectoryContextRetriever struct {
	Runner *interp.Runner
}

func (r WorkingDirectoryContextRetriever) GetContext() (string, error) {
	return fmt.Sprintf("<working_dir>%s</working_dir>", r.Runner.Vars["PWD"].String()), nil
}
