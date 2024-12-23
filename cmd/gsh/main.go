package main

import (
	"flag"
	"os"
	"strings"

	"github.com/atinylittleshell/gsh/internal/bash"
	"github.com/atinylittleshell/gsh/internal/core"
	"go.uber.org/zap"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/interp"
)

var command = flag.String("c", "", "command to run")

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	// Initialize the logger
	loggerConfig := zap.NewProductionConfig()
	if debug {
		os.Remove(core.LogFile())
		loggerConfig = zap.NewDevelopmentConfig()
	}
	loggerConfig.OutputPaths = []string{
		core.LogFile(),
	}

	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // Flush any buffered log entries

	err = run(logger)

	if code, ok := interp.IsExitStatus(err); ok {
		os.Exit(int(code))
	}

	if err != nil {
		logger.Error("unhandled error", zap.Error(err))
		os.Exit(1)
	}
}

func run(logger *zap.Logger) error {
	runner, err := interp.New(interp.Interactive(true), interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		return err
	}

	if *command != "" {
		return bash.RunBashCommandFromReader(runner, strings.NewReader(*command), "")
	}

	if flag.NArg() == 0 {
		if term.IsTerminal(int(os.Stdin.Fd())) {
			return core.RunApp(runner, logger)
		}

		return bash.RunBashCommandFromReader(runner, os.Stdin, "")
	}

	for _, filePath := range flag.Args() {
		if err := bash.RunBashScript(runner, filePath); err != nil {
			return err
		}
	}

	return nil
}
