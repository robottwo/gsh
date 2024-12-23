package bash

import (
	"context"
	"io"
	"os"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func RunBashCommandFromReader(runner *interp.Runner, reader io.Reader, name string) error {
	prog, err := syntax.NewParser().Parse(reader, name)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return runner.Run(ctx, prog)
}

func RunBashScript(runner *interp.Runner, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return RunBashCommandFromReader(runner, f, filePath)
}
