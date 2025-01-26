package completion

import (
	"context"
	"strings"

	"github.com/atinylittleshell/gsh/internal/environment"
	"mvdan.cc/sh/v3/interp"
)

// ShellCompletionProvider implements shellinput.CompletionProvider using the shell's CompletionManager
type ShellCompletionProvider struct {
	CompletionManager CompletionManagerInterface
	Runner            *interp.Runner
}

// NewShellCompletionProvider creates a new ShellCompletionProvider
func NewShellCompletionProvider(manager CompletionManagerInterface, runner *interp.Runner) *ShellCompletionProvider {
	return &ShellCompletionProvider{
		CompletionManager: manager,
		Runner:            runner,
	}
}

// GetCompletions returns completion suggestions for the current input line
func (p *ShellCompletionProvider) GetCompletions(line string, pos int) []string {
	// Split the line into words, preserving quotes
	line = line[:pos]
	words := splitPreservingQuotes(line)
	if len(words) == 0 {
		return make([]string, 0)
	}

	// Get the command (first word)
	command := words[0]

	// Look up completion spec for this command
	spec, ok := p.CompletionManager.GetSpec(command)
	if !ok {
		// No specific completion spec, try file path completion
		var prefix string
		if len(words) > 1 {
			// Get the last word as the prefix for file completion
			prefix = words[len(words)-1]
		} else if strings.HasSuffix(line, " ") {
			// If line ends with space, use empty prefix to list all files
			prefix = ""
		} else {
			return make([]string, 0)
		}

		completions := getFileCompletions(prefix, environment.GetPwd(p.Runner))

		// Prepend command to maintain the full command line
		cmdPrefix := command + " "
		for i, completion := range completions {
			if strings.Contains(completion, " ") {
				// Quote completions that contain spaces
				completions[i] = cmdPrefix + "\"" + completion + "\""
			} else {
				completions[i] = cmdPrefix + completion
			}
		}
		return completions
	}

	// Execute the completion
	suggestions, err := p.CompletionManager.ExecuteCompletion(context.Background(), p.Runner, spec, words)
	if err != nil {
		return make([]string, 0)
	}

	if suggestions == nil {
		return make([]string, 0)
	}
	return suggestions
}

