package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/internal/core"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

var command = flag.String("c", "", "command to run")

func main() {
	flag.Parse()

	err := run()

	if code, ok := interp.IsExitStatus(err); ok {
		os.Exit(int(code))
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "gsh: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	runner, err := interp.New(interp.Interactive(true), interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		return err
	}

	if *command != "" {
		return runCommand(runner, strings.NewReader(*command), "")
	}

	if flag.NArg() == 0 {
		if term.IsTerminal(int(os.Stdin.Fd())) {
			repl, err := core.NewREPL(runner)
			if err != nil {
				return err
			}
			return repl.Run()
		}

		return runCommand(runner, os.Stdin, "")
	}

	for _, filePath := range flag.Args() {
		if err := runFile(runner, filePath); err != nil {
			return err
		}
	}

	return nil
}

func runCommand(runner *interp.Runner, reader io.Reader, name string) error {
	prog, err := syntax.NewParser().Parse(reader, name)
	if err != nil {
		return err
	}
	runner.Reset()
	ctx := context.Background()
	return runner.Run(ctx, prog)
}

func runFile(runner *interp.Runner, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return runCommand(runner, f, filePath)
}
