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
	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/history"
	"go.uber.org/zap"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

// go:embed ../../.gshrc.default
var DEFAULT_VARS []byte

var command = flag.String("c", "", "run a command")
var listHistory = flag.Int("lh", 0, "list the most N history entries")
var resetHistory = flag.Bool("rh", false, "reset history")
var loginShell = flag.Bool("l", false, "run as a login shell")

var helpFlag = flag.Bool("h", false, "display help information")

func main() {
	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage of gsh:")
		flag.PrintDefaults()
		return
	}

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
	logLevel := environment.GetLogLevel(runner)

	if environment.ShouldCleanLogFile(runner) {
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
	shellPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	env := expand.ListEnviron(append(os.Environ(), fmt.Sprintf("SHELL=%s", shellPath))...)

	runner, err := interp.New(
		interp.Interactive(true),
		interp.Env(env),
		interp.StdIO(os.Stdin, os.Stdout, os.Stderr),
	)
	if err != nil {
		panic(err)
	}

	// load default vars
	if err := bash.RunBashScriptFromReader(
		runner,
		bytes.NewReader(DEFAULT_VARS),
		"DEFAULT_VARS",
	); err != nil {
		panic(err)
	}

	configFiles := []string{
		filepath.Join(core.HomeDir(), ".gshenv"),
		filepath.Join(core.HomeDir(), ".gshrc"),
	}

	// Check if this is a login shell
	if *loginShell {
		// Prepend .gsh_profile to the list of config files
		configFiles = append([]string{filepath.Join(core.HomeDir(), ".gsh_profile")}, configFiles...)
	}

	for _, configFile := range configFiles {
		if stat, err := os.Stat(configFile); err == nil && stat.Size() > 0 {
			if err := bash.RunBashScriptFromFile(runner, configFile); err != nil {
				fmt.Fprintf(os.Stderr, "failed to load %s: %v\n", configFile, err)
			}
		}
	}

	return runner, nil
}
