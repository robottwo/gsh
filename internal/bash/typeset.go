package bash

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// For testing purposes
var typesetPrintf = fmt.Printf

// Global runner reference that can be set after initialization
var globalRunner *interp.Runner

// SetTypesetRunner sets the Runner used by the package for handling the `typeset`/`declare` command handler.
// Passing nil clears the stored runner.
func SetTypesetRunner(runner *interp.Runner) {
	globalRunner = runner
}

// NewTypesetCommandHandler returns an exec handler middleware that intercepts the `typeset`, `declare` and `gsh_typeset` commands.
// If the first argument is one of those command names the returned handler implements typeset behavior using the package runner; otherwise it forwards the call to the next handler. If the package runner has not been initialized the handler returns an error.
func NewTypesetCommandHandler() func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	return func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
		return func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return next(ctx, args)
			}

			// Handle both typeset and declare commands
			if args[0] != "typeset" && args[0] != "declare" && args[0] != "gsh_typeset" {
				return next(ctx, args)
			}

			// Use the global runner reference
			if globalRunner == nil {
				return fmt.Errorf("typeset: runner not initialized")
			}

			// Now we have access to the runner, so we can implement the real functionality
			return handleTypesetCommand(globalRunner, args)
		}
	}
}

// handleTypesetCommand parses typeset/declare command arguments and invokes
// the corresponding listing routine.
//
// The function recognizes the flags `-f` (list full function definitions),
// `-F` (list function names only), and `-p` (list variables with attributes).
// Flags may be combined (for example `-fp`). Option parsing stops at the first
// non-option argument. If no flag is provided, it defaults to listing variables.
// Returns an error for any unrecognized flag or any error returned by the
// selected printing routine.
func handleTypesetCommand(runner *interp.Runner, args []string) error {
	// Parse options - skip the command name (args[0])
	var (
		listFunctions     bool // -f: list function definitions
		listFunctionNames bool // -F: list function names only
		listVariables     bool // -p: list variables with attributes
	)

	// If no options provided, default to listing variables
	if len(args) <= 1 {
		listVariables = true
	}

	// Parse command-line options - start from args[1] to skip command name
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			// Non-option argument, stop parsing options
			break
		}

		// Handle combined options like -fp
		for _, ch := range arg[1:] {
			switch ch {
			case 'f':
				listFunctions = true
			case 'F':
				listFunctionNames = true
			case 'p':
				listVariables = true
			default:
				return fmt.Errorf("typeset: -%c: invalid option", ch)
			}
		}
	}

	// If no specific option was set, default to listing variables
	if !listFunctions && !listFunctionNames && !listVariables {
		listVariables = true
	}

	// Handle function listing
	if listFunctions {
		return printFunctionDefinitions(runner)
	}

	if listFunctionNames {
		return printFunctionNames(runner)
	}

	// Handle variable listing
	if listVariables {
		return printVariables(runner)
	}

	return nil
}

// printFunctionDefinitions prints all function definitions in a bash-compatible format.
// If runner.Funcs is nil the function does nothing and returns nil.
// Function names are printed in sorted order; nil function entries are skipped.
func printFunctionDefinitions(runner *interp.Runner) error {
	if runner.Funcs == nil {
		return nil
	}

	// Get all function names and sort them
	names := make([]string, 0, len(runner.Funcs))
	for name := range runner.Funcs {
		names = append(names, name)
	}
	sort.Strings(names)

	// Print each function definition
	for _, name := range names {
		fn := runner.Funcs[name]
		if fn == nil {
			continue
		}

		// Format: function_name () { body }
		typesetPrintf("%s () \n{ \n", name)

		// Print the function body
		printFunctionBody(fn)

		typesetPrintf("}\n")
	}

	return nil
}

// printFunctionBody prints the statements of a function body, formatted using the shell syntax printer and indented for typeset output.
// If fn is nil, it does nothing.
func printFunctionBody(fn *syntax.Stmt) {
	if fn == nil {
		return
	}

	// Use the syntax printer to format the function body
	// This will give us a bash-compatible representation
	var buf strings.Builder
	syntax.NewPrinter().Print(&buf, fn)

	// Indent each line of the body
	lines := strings.Split(buf.String(), "\n")
	for _, line := range lines {
		if line != "" {
			typesetPrintf("    %s\n", line)
		}
	}
}

// printFunctionNames prints just the function names (one per line)
func printFunctionNames(runner *interp.Runner) error {
	if runner.Funcs == nil {
		return nil
	}

	// Get all function names and sort them
	names := make([]string, 0, len(runner.Funcs))
	for name := range runner.Funcs {
		names = append(names, name)
	}
	sort.Strings(names)

	// Print each function name
	for _, name := range names {
		typesetPrintf("declare -f %s\n", name)
	}

	return nil
}

// printVariables prints all variables from the provided runner in Bash-compatible
// `declare` form, using `-x` for exported variables and `--` for non-exported ones.
// Values are written quoted. If the runner has no variables, the function does nothing.
func printVariables(runner *interp.Runner) error {
	if runner.Vars == nil {
		return nil
	}

	// Collect all variable names
	var names []string
	for name := range runner.Vars {
		names = append(names, name)
	}
	sort.Strings(names)

	// Print each variable
	for _, name := range names {
		vr, ok := runner.Vars[name]
		if !ok {
			continue
		}

		value := vr.String()

		// Determine if the variable is exported
		exported := vr.Exported

		// Format the output
		if exported {
			typesetPrintf("declare -x %s=%q\n", name, value)
		} else {
			typesetPrintf("declare -- %s=%q\n", name, value)
		}
	}

	return nil
}