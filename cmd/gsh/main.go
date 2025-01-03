package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/atinylittleshell/gsh/internal/bash"
	"github.com/atinylittleshell/gsh/internal/core"
	"github.com/atinylittleshell/gsh/internal/environment"
	"github.com/atinylittleshell/gsh/internal/history"
	"github.com/atinylittleshell/gsh/internal/styles"
	"github.com/atinylittleshell/gsh/pkg/gline"
	"github.com/creativeprojects/go-selfupdate"
	"go.uber.org/zap"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

var BUILD_VERSION = "dev"

// go:embed ../../.gshrc.default
var DEFAULT_VARS []byte

var command = flag.String("c", "", "run a command")
var listHistory = flag.Int("lh", 0, "list the most N history entries")
var resetHistory = flag.Bool("rh", false, "reset history")
var loginShell = flag.Bool("l", false, "run as a login shell")

var helpFlag = flag.Bool("h", false, "display help information")
var versionFlag = flag.Bool("ver", false, "display build version")

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Println(BUILD_VERSION)
		return
	}

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

	handleSelfUpdate(logger)

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
		return bash.RunBashScriptFromReader(runner, strings.NewReader(*command), "gsh")
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

		return bash.RunBashScriptFromReader(runner, os.Stdin, "gsh")
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
	env := expand.ListEnviron(append(
		os.Environ(),
		fmt.Sprintf("SHELL=%s", shellPath),
	)...)

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
		"gsh",
	); err != nil {
		panic(err)
	}

	configFiles := []string{
		filepath.Join(core.HomeDir(), ".gshrc"),
		filepath.Join(core.HomeDir(), ".gshenv"),
	}

	// Check if this is a login shell
	if *loginShell || strings.HasPrefix(os.Args[0], "-") {
		// Prepend .gsh_profile to the list of config files
		configFiles = append(
			[]string{
				"/etc/profile",
				filepath.Join(core.HomeDir(), ".gsh_profile"),
			},
			configFiles...,
		)
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

func handleSelfUpdate(logger *zap.Logger) {
	// No need to do anything if we are running a dev build
	if BUILD_VERSION == "dev" {
		logger.Debug("running a dev build, skipping self-update check")
		return
	}

	// Check if we have previously detected a newer version
	checkPreviouslyDetectedVersion(logger)

	// Check for newer versions from remote repository
	go detectUpdate(logger)
}

func checkPreviouslyDetectedVersion(logger *zap.Logger) {
	file, err := os.Open(core.DetectedVersionFile())
	if err != nil {
		logger.Debug("detected version file not found", zap.Error(err))
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		logger.Error("failed to read detected version", zap.Error(err))
		return
	}

	detectedVersion := strings.TrimSpace(buf.String())
	if detectedVersion == "" || detectedVersion == BUILD_VERSION {
		return
	}

	confirm, err := gline.Gline(
		styles.AGENT_QUESTION("New version of gsh available. Update now? (Y/n)"),
		detectedVersion,
		nil,
		nil,
		logger,
		gline.NewOptions(),
	)

	if strings.ToLower(confirm) == "n" {
		return
	}

	latest, found, err := selfupdate.DetectLatest(
		context.Background(),
		selfupdate.ParseSlug("atinylittleshell/gsh"),
	)
	if err != nil {
		logger.Warn("error occurred while detecting latest version", zap.Error(err))
		return
	}
	if !found {
		logger.Warn("latest version could not be detected")
		return
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		logger.Error("failed to get executable path to update", zap.Error(err))
		return
	}
	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		logger.Error("failed to update to latest version", zap.Error(err))
		return
	}

	logger.Info("successfully updated to latest version", zap.String("version", latest.Version()))
}

func detectUpdate(logger *zap.Logger) {
	latest, found, err := selfupdate.DetectLatest(
		context.Background(),
		selfupdate.ParseSlug("atinylittleshell/gsh"),
	)
	if err != nil {
		logger.Warn("error occurred while detecting latest version", zap.Error(err))
		return
	}
	if !found {
		logger.Warn("latest version could not be detected")
		return
	}

	currentVersion := BUILD_VERSION
	if currentVersion == "dev" {
		currentVersion = "0.0.0"
	}

	if latest.GreaterThan(currentVersion) {
		logger.Info("new version of gsh available", zap.String("version", latest.Version()))

		recordFilePath := core.DetectedVersionFile()
		file, err := os.Create(recordFilePath)
		defer file.Close()

		if err != nil {
			logger.Error("failed to save detected version", zap.Error(err))
			return
		}

		_, err = file.WriteString(latest.Version())
		if err != nil {
			logger.Error("failed to save detected version", zap.Error(err))
			return
		}
	}
}
