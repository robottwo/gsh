package bash

import (
	"context"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// PreprocessTypesetCommands replaces occurrences of `typeset` and `declare` with `gsh_typeset` for the flags -f, -F, and -p.
// It also handles simple extra-space variants (for example, "typeset  -f"). The input string is returned with those substitutions applied.
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

// RunBashScriptFromReader reads all data from reader, preprocesses typeset/declare
// variants into gsh_typeset, parses the result as a shell program named by name,
// and executes it using runner.
//
// It returns any read or parse error encountered. Errors produced by executing
// the parsed program via runner are suppressed (nil is returned) so the caller's
// script processing can continue even if individual commands fail.
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

// RunBashScriptFromFile opens the file at filePath and executes its contents as a Bash-like script using the provided runner.
// It returns an error if opening the file fails or if executing the script (via RunBashScriptFromReader) fails.
func RunBashScriptFromFile(ctx context.Context, runner *interp.Runner, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return RunBashScriptFromReader(ctx, runner, f, filePath)
}

// RunBashCommandInSubShell runs a single shell command in a subshell and captures its stdout and stderr.
// 
// It returns the captured stdout, the captured stderr, and any error encountered while parsing or executing the command.
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

// RunBashCommand runs the given shell command using the provided runner and returns its captured stdout and stderr.
// It preprocesses common `typeset`/`declare` variants before parsing, executes the first parsed statement, and captures its standard output and standard error.
// The returned values are the captured stdout, the captured stderr, and an error if parsing or execution fails.
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