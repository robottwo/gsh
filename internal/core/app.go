package core

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/pkg/gline"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

const (
	EXIT_COMMAND = "exit"
)

func RunApp() error {
	runner, err := interp.New(interp.Interactive(true), interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		return err
	}
	predictor := NewLLMPredictor()

	commandIndex := 0

	for {
		// Read input
		line, err := gline.NextLine("gsh> ", predictor, gline.Options{
			ClearScreen: commandIndex == 0,
		})
		commandIndex++

		if err != nil {
			return err
		}

		// Exit if the user types the exit command
		if line == EXIT_COMMAND {
			break
		}

		// Execute the command
		executeCommand(line, runner)
	}

	// Clear screen on exit
	fmt.Print(gline.CLEAR_SCREEN)
	return nil
}

func executeCommand(input string, runner *interp.Runner) error {
	prog, err := syntax.NewParser().Parse(strings.NewReader(input), "")
	if err != nil {
		return err
	}
	return runner.Run(context.Background(), prog)
}
