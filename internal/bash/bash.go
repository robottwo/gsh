package bash

import (
	"context"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func PreprocessTypesetCommands(input string) string {
	// Simple string replacement approach - more reliable than complex regex
	result := input

	// Replace typeset -f with gsh_typeset -f
	result = strings.ReplaceAll(result, "typeset -f", "gsh_typeset -f")
	result = strings.ReplaceAll(result, "declare -f", "gsh_typeset -f")
	result = strings.ReplaceAll(result, "typeset -F", "gsh_typeset -F")
	result = strings.ReplaceAll(result, "declare -F", "gsh_typeset -F")
	result = strings.ReplaceAll(result, "typeset -p", "gsh_typeset -p")
	result = strings.ReplaceAll(result, "declare -p", "gsh_typeset -p")

	// Handle cases where there might be extra spaces
	result = strings.ReplaceAll(result, "typeset  -f", "gsh_typeset -f")
	result = strings.ReplaceAll(result, "declare  -f", "gsh_typeset -f")
	result = strings.ReplaceAll(result, "typeset  -F", "gsh_typeset -F")
	result = strings.ReplaceAll(result, "declare  -F", "gsh_typeset -F")
	result = strings.ReplaceAll(result, "typeset  -p", "gsh_typeset -p")
	result = strings.ReplaceAll(result, "declare  -p", "gsh_typeset -p")

	return result
}

func RunBashScriptFromReader(ctx context.Context, runner *interp.Runner, reader io.Reader, name string) error {
	// Read the entire input first
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	// Pre-process the content to transform typeset/declare commands
	processedContent := PreprocessTypesetCommands(string(content))

	prog, err := syntax.NewParser().Parse(strings.NewReader(processedContent), name)
	if err != nil {
		return err
	}

	// Run the script, but continue even if individual commands fail
	// This mimics bash behavior where the script continues unless set -e is used
	err = runner.Run(ctx, prog)
	if err != nil {
		// Log the error but don't fail the entire script
		// This allows .gshrc to continue loading even if individual commands fail
		return nil
	}
	return nil
}

func RunBashScriptFromFile(ctx context.Context, runner *interp.Runner, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return RunBashScriptFromReader(ctx, runner, f, filePath)
}

func RunBashCommandInSubShell(ctx context.Context, runner *interp.Runner, command string) (string, string, error) {
	// Pre-process the command to transform typeset/declare commands
	processedCommand := PreprocessTypesetCommands(command)

	subShell := runner.Subshell()

	outBuf := &threadSafeBuffer{}
	outWriter := io.Writer(outBuf)
	errBuf := &threadSafeBuffer{}
	errWriter := io.Writer(errBuf)
	interp.StdIO(nil, outWriter, errWriter)(subShell)

	var prog *syntax.Stmt
	err := syntax.NewParser().Stmts(strings.NewReader(processedCommand), func(stmt *syntax.Stmt) bool {
		prog = stmt
		return false
	})
	if err != nil {
		return "", "", err
	}

	err = subShell.Run(ctx, prog)
	if err != nil {
		return "", "", err
	}

	return outBuf.String(), errBuf.String(), nil
}

func RunBashCommand(ctx context.Context, runner *interp.Runner, command string) (string, string, error) {
	// Pre-process the command to transform typeset/declare commands
	processedCommand := PreprocessTypesetCommands(command)

	outBuf := &threadSafeBuffer{}
	outWriter := io.Writer(outBuf)
	errBuf := &threadSafeBuffer{}
	errWriter := io.Writer(errBuf)
	interp.StdIO(nil, outWriter, errWriter)(runner)
	defer interp.StdIO(os.Stdin, os.Stdout, os.Stderr)(runner)

	var prog *syntax.Stmt
	err := syntax.NewParser().Stmts(strings.NewReader(processedCommand), func(stmt *syntax.Stmt) bool {
		prog = stmt
		return false
	})
	if err != nil {
		return "", "", err
	}

	err = runner.Run(ctx, prog)
	if err != nil {
		return "", "", err
	}

	return outBuf.String(), errBuf.String(), nil
}
