package bash

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func RunBashScriptFromReader(runner *interp.Runner, reader io.Reader, name string) error {
	prog, err := syntax.NewParser().Parse(reader, name)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return runner.Run(ctx, prog)
}

func RunBashScriptFromFile(runner *interp.Runner, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return RunBashScriptFromReader(runner, f, filePath)
}

func RunBashCommandInSubShell(runner *interp.Runner, command string) (string, string, error) {
	subShell := runner.Subshell()

	outBuf := &bytes.Buffer{}
	outWriter := io.Writer(outBuf)
	errBuf := &bytes.Buffer{}
	errWriter := io.Writer(errBuf)
	interp.StdIO(nil, outWriter, errWriter)(subShell)

	var prog *syntax.Stmt
	err := syntax.NewParser().Stmts(strings.NewReader(command), func(stmt *syntax.Stmt) bool {
		prog = stmt
		return false
	})
	if err != nil {
		return "", "", err
	}

	err = subShell.Run(context.Background(), prog)
	if err != nil {
		return "", "", err
	}

	return outBuf.String(), errBuf.String(), nil
}

func RunBashCommand(runner *interp.Runner, command string) (string, string, error) {
	outBuf := &bytes.Buffer{}
	outWriter := io.Writer(outBuf)
	errBuf := &bytes.Buffer{}
	errWriter := io.Writer(errBuf)
	interp.StdIO(nil, outWriter, errWriter)(runner)
	defer interp.StdIO(os.Stdin, os.Stdout, os.Stderr)(runner)

	var prog *syntax.Stmt
	err := syntax.NewParser().Stmts(strings.NewReader(command), func(stmt *syntax.Stmt) bool {
		prog = stmt
		return false
	})
	if err != nil {
		return "", "", err
	}

	err = runner.Run(context.Background(), prog)
	if err != nil {
		return "", "", err
	}

	return outBuf.String(), errBuf.String(), nil
}
