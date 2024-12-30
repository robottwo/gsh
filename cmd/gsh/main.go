package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atinylittleshell/gsh/internal/bash"
	"github.com/atinylittleshell/gsh/internal/core"
	"github.com/atinylittleshell/gsh/internal/history"
	"go.uber.org/zap"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/interp"
)

// go:embed ../../.gshrc.default
var DEFAULT_VARS []byte

var command = flag.String("c", "", "command to run")
var listHistory = flag.Int("lh", 0, "list the most N history entries")
var resetHistory = flag.Bool("rh", false, "reset the history")

func main() {
	// Initialize the shell interpreter
	runner, err := initializeRunner()
	if err != nil {
		panic(err)
	}

	// Initialize the logger
	logger, err := initializeLogger(runner)
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // Flush any buffered log entries

	// Initialize the history manager
	historyManager, err := initializeHistoryManager(logger)
	if err != nil {
		logger.Error("failed to initialize history manager", zap.Error(err))
		os.Exit(1)
	}

	// Start running
	err = run(runner, historyManager, logger)

	// Handle exit status
	if code, ok := interp.IsExitStatus(err); ok {
		os.Exit(int(code))
	}

	if err != nil {
		logger.Error("unhandled error", zap.Error(err))
		os.Exit(1)
	}
}

func run(runner *interp.Runner, historyManager *history.HistoryManager, logger *zap.Logger) error {
	logger.Info("-------- new gsh session --------")

	flag.Parse()

	// gsh -c "echo hello"
	if *command != "" {
		return bash.RunBashScriptFromReader(runner, strings.NewReader(*command), "")
	}

	// gsh -lh 5
	if *listHistory > 0 {
		entries, err := historyManager.GetRecentEntries("", *listHistory)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			fmt.Println(entry.Command)
		}

		return nil
	}

	// gsh -rh
	if *resetHistory {
		if err := historyManager.ResetHistory(); err != nil {
			return err
		}

		return nil
	}

	// gsh
	if flag.NArg() == 0 {
		if term.IsTerminal(int(os.Stdin.Fd())) {
			return core.RunInteractiveShell(runner, historyManager, logger)
		}

		return bash.RunBashScriptFromReader(runner, os.Stdin, "")
	}

	// gsh script.sh
	for _, filePath := range flag.Args() {
		if err := bash.RunBashScriptFromFile(runner, filePath); err != nil {
			return err
		}
	}

	return nil
}

func initializeLogger(runner *interp.Runner) (*zap.Logger, error) {
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
		return nil, err
	}

	return logger, nil
}

func initializeHistoryManager(logger *zap.Logger) (*history.HistoryManager, error) {
	historyManager, err := history.NewHistoryManager(core.HistoryFile(), logger)
	if err != nil {
		return nil, err
	}

	return historyManager, nil
}

// initializeRunner loads the shell configuration files and sets up the interpreter.
func initializeRunner() (*interp.Runner, error) {
	runner, err := interp.New(interp.Interactive(true), interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		panic(err)
	}

	// load default vars
	if err := bash.RunBashScriptFromReader(runner, bytes.NewReader(DEFAULT_VARS), "DEFAULT_VARS"); err != nil {
		panic(err)
	}

	configFiles := []string{
		filepath.Join(core.HomeDir(), ".gshenv"),
		filepath.Join(core.HomeDir(), ".gshrc"),
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			if err := bash.RunBashScriptFromFile(runner, configFile); err != nil {
				fmt.Fprintf(os.Stderr, "failed to load %s: %v\n", configFile, err)
			}
		}
	}

	return runner, nil
}
