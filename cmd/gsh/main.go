package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/internal/bash"
	"github.com/atinylittleshell/gsh/internal/core"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/interp"
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
		return bash.RunBashCommand(runner, strings.NewReader(*command), "")
	}

	if flag.NArg() == 0 {
		if term.IsTerminal(int(os.Stdin.Fd())) {
			return core.RunApp(runner)
		}

		return bash.RunBashCommand(runner, os.Stdin, "")
	}

	for _, filePath := range flag.Args() {
		if err := bash.RunBashScript(runner, filePath); err != nil {
			return err
		}
	}

	return nil
}
