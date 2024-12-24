package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atinylittleshell/gsh/internal/bash"
	"github.com/atinylittleshell/gsh/internal/core"
	"go.uber.org/zap"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/interp"
)

var command = flag.String("c", "", "command to run")

func main() {
	runner, err := interp.New(interp.Interactive(true), interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		panic(err)
	}

	// Load gshrc
	loadShellConfigs(runner)

	// Determine the log level
	logLevel, err := zap.ParseAtomicLevel(runner.Vars["GSH_LOG_LEVEL"].String())
	if err != nil {
		logLevel = zap.NewAtomicLevel()
	}

	// Start with a clean log file if requested
	cleanLogFile := runner.Vars["GSH_CLEAN_LOG_FILE"].String()
	if cleanLogFile == "1" || cleanLogFile == "true" {
		os.Remove(core.LogFile())
	}

	// Initialize the logger
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = logLevel
	loggerConfig.OutputPaths = []string{
		core.LogFile(),
	}
	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // Flush any buffered log entries

	err = run(runner, logger)

	if code, ok := interp.IsExitStatus(err); ok {
		os.Exit(int(code))
	}

	if err != nil {
		logger.Error("unhandled error", zap.Error(err))
		os.Exit(1)
	}
}

func run(runner *interp.Runner, logger *zap.Logger) error {
	logger.Info("-------- new gsh session --------")

	flag.Parse()

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

// loadShellConfigs loads and executes .gshrc
func loadShellConfigs(runner *interp.Runner) error {
	configFiles := []string{
		filepath.Join(core.HomeDir(), ".gshrc"),
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			if err := bash.RunBashScript(runner, configFile); err != nil {
				fmt.Fprintf(os.Stderr, "failed to load %s: %v\n", configFile, err)
			}
		}
	}

	return nil
}
