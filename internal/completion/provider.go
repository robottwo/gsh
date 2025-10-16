package completion

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/atinylittleshell/gsh/internal/environment"
	"mvdan.cc/sh/v3/interp"
)

// Function variables for mocking in tests
var osReadDir = os.ReadDir

// SubagentInfo represents minimal information about a subagent for completion purposes
type SubagentInfo struct {
	ID           string
	Name         string
	Description  string
	AllowedTools []string
	FileRegex    string
	Model        string
}

// SubagentProvider defines the interface for providing subagent information to the completion system
type SubagentProvider interface {
	GetAllSubagents() map[string]*SubagentInfo
	GetSubagent(id string) (*SubagentInfo, bool)
}

// ShellCompletionProvider implements shellinput.CompletionProvider using the shell's CompletionManager
type ShellCompletionProvider struct {
	CompletionManager CompletionManagerInterface
	Runner            *interp.Runner
	SubagentProvider  SubagentProvider // Optional, for @ completions
}

// NewShellCompletionProvider creates a new ShellCompletionProvider
func NewShellCompletionProvider(manager CompletionManagerInterface, runner *interp.Runner) *ShellCompletionProvider {
	return &ShellCompletionProvider{
		CompletionManager: manager,
		Runner:            runner,
		SubagentProvider:  nil, // Set later via SetSubagentProvider if needed
	}
}

// SetSubagentProvider sets the subagent provider for @ completions
func (p *ShellCompletionProvider) SetSubagentProvider(provider SubagentProvider) {
	p.SubagentProvider = provider
}

// GetCompletions returns completion suggestions for the current input line
func (p *ShellCompletionProvider) GetCompletions(line string, pos int) []string {
	// First check for special prefixes (#/ and #!)
	if completion := p.checkSpecialPrefixes(line, pos); completion != nil {
		return completion
	}

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
		// No specific completion spec, check if we should complete command names
		if len(words) == 1 && !strings.HasSuffix(line, " ") {
			// Single word that doesn't end with space
			// Check if this looks like a path-based command
			if p.isPathBasedCommand(command) {
				// For path-based commands, complete with executable files in that path
				executableCompletions := p.getExecutableCompletions(command)
				if len(executableCompletions) > 0 {
					return executableCompletions
				}
			} else {
				// Regular command name completion
				commandCompletions := p.getAvailableCommands(command)
				if len(commandCompletions) > 0 {
					return commandCompletions
				}
			}
		}

		// No command matches or multiple words, try file path completion
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

		// Quote completions that contain spaces, but don't add command prefix
		// The completion handler will replace only the current word (file path)
		for i, completion := range completions {
			if strings.Contains(completion, " ") {
				// Quote completions that contain spaces
				completions[i] = "\"" + completion + "\""
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

// checkSpecialPrefixes checks for #/, #!, and #@ prefixes and returns appropriate completions
func (p *ShellCompletionProvider) checkSpecialPrefixes(line string, pos int) []string {
	// Get the current word being completed
	start, end := p.getCurrentWordBoundary(line, pos)
	if start < 0 || end < 0 {
		return nil
	}

	currentWord := line[start:end]

	// Check if the current word starts with @/, @!, or @
	if strings.HasPrefix(currentWord, "@/") {
		completions := p.getMacroCompletions(currentWord)
		if len(completions) == 0 {
			// No macro matches found, fall back to path completion
			pathPrefix := strings.TrimPrefix(currentWord, "@/")
			completions := getFileCompletions(pathPrefix, environment.GetPwd(p.Runner))

			// Build the proper prefix for the current line context
			var linePrefix string
			if start > 0 {
				linePrefix = line[:start]
			}

			// Add completions with proper prefix
			for i, completion := range completions {
				completions[i] = linePrefix + completion
			}
			return completions
		}
		return completions
	} else if strings.HasPrefix(currentWord, "@!") {
		completions := p.getBuiltinCommandCompletions(currentWord)
		if len(completions) == 0 {
			// No builtin command matches found, fall back to path completion
			pathPrefix := strings.TrimPrefix(currentWord, "@!")
			completions := getFileCompletions(pathPrefix, environment.GetPwd(p.Runner))

			// Build the proper prefix for the current line context
			var linePrefix string
			if start > 0 {
				linePrefix = line[:start]
			}

			// Add completions with proper prefix
			for i, completion := range completions {
				completions[i] = linePrefix + completion
			}
			return completions
		}
		return completions
	} else if strings.HasPrefix(currentWord, "@") && !strings.HasPrefix(currentWord, "@/") && !strings.HasPrefix(currentWord, "@!") {
		// Subagent completions - allow anywhere in the line, not just at the start
		completions := p.getSubagentCompletions(currentWord)

		// Build the proper prefix and suffix for the current line context
		var linePrefix string
		if start > 0 {
			linePrefix = line[:start]
		}
		var lineSuffix string
		if end < len(line) {
			lineSuffix = line[end:]
		}

		if len(completions) == 0 {
			// No subagent matches found, fall back to path completion
			pathPrefix := strings.TrimPrefix(currentWord, "@")
			completions := getFileCompletions(pathPrefix, environment.GetPwd(p.Runner))

			// Add completions with proper prefix and suffix
			for i, completion := range completions {
				completions[i] = linePrefix + completion + lineSuffix
			}
			return completions
		}

		// Add completions with proper prefix and suffix
		for i, completion := range completions {
			completions[i] = linePrefix + completion + lineSuffix
		}
		return completions
	}

	// Also check if we're at the beginning of a potential prefix
	// Look backwards to see if there's a @/, @!, or @ that we should complete
	if start > 0 {
		// Find the start of the word that might contain our prefix
		wordStart := start
		for wordStart > 0 && !unicode.IsSpace(rune(line[wordStart-1])) {
			wordStart--
		}

		potentialWord := line[wordStart:end]
		if strings.HasPrefix(potentialWord, "@/") {
			completions := p.getMacroCompletions(potentialWord)
			if len(completions) == 0 {
				// No macro matches found, fall back to path completion
				pathPrefix := strings.TrimPrefix(potentialWord, "@/")
				completions := getFileCompletions(pathPrefix, environment.GetPwd(p.Runner))

				// Build the proper prefix for the current line context
				var linePrefix string
				if wordStart > 0 {
					linePrefix = line[:wordStart]
				}

				// Add completions with proper prefix
				for i, completion := range completions {
					completions[i] = linePrefix + completion
				}
				return completions
			}
			return completions
		} else if strings.HasPrefix(potentialWord, "@!") {
			completions := p.getBuiltinCommandCompletions(potentialWord)
			if len(completions) == 0 {
				// No builtin command matches found, fall back to path completion
				pathPrefix := strings.TrimPrefix(potentialWord, "@!")
				completions := getFileCompletions(pathPrefix, environment.GetPwd(p.Runner))

				// Build the proper prefix for the current line context
				var linePrefix string
				if wordStart > 0 {
					linePrefix = line[:wordStart]
				}

				// Add completions with proper prefix
				for i, completion := range completions {
					completions[i] = linePrefix + completion
				}
				return completions
			}
			return completions
		} else if strings.HasPrefix(potentialWord, "@") && !strings.HasPrefix(potentialWord, "@/") && !strings.HasPrefix(potentialWord, "@!") {
			// Subagent completions - only if this is the first non-whitespace on the line
			if !p.isAtLineStart(line, wordStart) {
				return nil
			}
			completions := p.getSubagentCompletions(potentialWord)

			// Build the proper prefix and suffix for the current line context
			var linePrefix string
			if wordStart > 0 {
				linePrefix = line[:wordStart]
			}
			var lineSuffix string
			if end < len(line) {
				lineSuffix = line[end:]
			}

			if len(completions) == 0 {
				// No subagent matches found, fall back to path completion
				pathPrefix := strings.TrimPrefix(potentialWord, "@")
				completions := getFileCompletions(pathPrefix, environment.GetPwd(p.Runner))

				// Add completions with proper prefix and suffix
				for i, completion := range completions {
					completions[i] = linePrefix + completion + lineSuffix
				}
				return completions
			}

			// Add completions with proper prefix and suffix
			for i, completion := range completions {
				completions[i] = linePrefix + completion + lineSuffix
			}
			return completions
		}
	}

	return nil
}

// isAtLineStart checks if the given position is at the start of the line (after whitespace)
func (p *ShellCompletionProvider) isAtLineStart(line string, pos int) bool {
	if pos <= 0 {
		return true
	}
	// Check if all characters before this position are whitespace
	for i := 0; i < pos; i++ {
		if !unicode.IsSpace(rune(line[i])) {
			return false
		}
	}
	return true
}

// getCurrentWordBoundary finds the start and end of the current word at cursor position
func (p *ShellCompletionProvider) getCurrentWordBoundary(line string, pos int) (int, int) {
	if len(line) == 0 || pos > len(line) {
		return -1, -1
	}

	// Find start of word
	start := pos
	for start > 0 && !unicode.IsSpace(rune(line[start-1])) {
		start--
	}

	// Find end of word
	end := pos
	for end < len(line) && !unicode.IsSpace(rune(line[end])) {
		end++
	}

	return start, end
}

// getMacroCompletions returns completions for macros starting with @/
func (p *ShellCompletionProvider) getMacroCompletions(prefix string) []string {
	var macrosStr string
	if p.Runner != nil {
		macrosStr = p.Runner.Vars["GSH_AGENT_MACROS"].String()
	} else {
		// Fallback to environment variable for testing
		macrosStr = os.Getenv("GSH_AGENT_MACROS")
	}

	if macrosStr == "" {
		return []string{}
	}

	var macros map[string]interface{}
	if err := json.Unmarshal([]byte(macrosStr), &macros); err != nil {
		return []string{}
	}

	var completions []string
	prefixAfterSlash := strings.TrimPrefix(prefix, "@/")

	for macroName := range macros {
		if strings.HasPrefix(macroName, prefixAfterSlash) {
			completions = append(completions, "@/"+macroName)
		}
	}

	// Sort alphabetically for consistent ordering
	sort.Strings(completions)
	return completions
}

// isPathBasedCommand determines if a command looks like a path rather than a simple command name
func (p *ShellCompletionProvider) isPathBasedCommand(command string) bool {
	// Check for common path patterns
	return strings.HasPrefix(command, "/") || // Absolute path: /bin/ls
		strings.HasPrefix(command, "./") || // Relative path: ./script
		strings.HasPrefix(command, "../") || // Parent directory: ../script
		strings.HasPrefix(command, "~/") || // Home directory: ~/bin/script
		strings.Contains(command, "/") // Any path with directory separator
}

// getExecutableCompletions returns executable files that match the given path prefix
func (p *ShellCompletionProvider) getExecutableCompletions(pathPrefix string) []string {
	// Determine the directory to search and the filename prefix
	var searchDir, filePrefix string

	if strings.HasSuffix(pathPrefix, "/") {
		// Path ends with /, so we want all executables in that directory
		searchDir = pathPrefix
		filePrefix = ""
	} else {
		// Extract directory and filename parts
		searchDir = filepath.Dir(pathPrefix)
		filePrefix = filepath.Base(pathPrefix)

		// Handle special case where pathPrefix doesn't contain a directory separator
		if searchDir == "." && !strings.Contains(pathPrefix, "/") {
			return []string{} // This shouldn't be a path-based command
		}
	}

	// Resolve the search directory
	var resolvedDir string
	if strings.HasPrefix(searchDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return []string{}
		}
		resolvedDir = filepath.Join(homeDir, searchDir[2:])
	} else if filepath.IsAbs(searchDir) {
		resolvedDir = searchDir
	} else {
		// Relative path
		currentDir := environment.GetPwd(p.Runner)
		resolvedDir = filepath.Join(currentDir, searchDir)
	}

	// Read directory contents
	entries, err := osReadDir(resolvedDir)
	if err != nil {
		return []string{}
	}

	var completions []string
	for _, entry := range entries {
		// Skip directories and non-matching files
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), filePrefix) {
			continue
		}

		// Check if file is executable (simplified check)
		// In a more complete implementation, we'd check file permissions
		if info, err := entry.Info(); err == nil {
			// On Unix-like systems, check if any execute bit is set
			if info.Mode()&0111 != 0 {
				// Build the completion preserving the original path structure
				if strings.HasSuffix(pathPrefix, "/") {
					completions = append(completions, pathPrefix+entry.Name())
				} else {
					// Replace the filename part with the matched file
					completions = append(completions, filepath.Join(searchDir, entry.Name()))
				}
			}
		}
	}

	// Sort alphabetically for consistent ordering
	sort.Strings(completions)
	return completions
}

// getAvailableCommands returns available system commands that match the given prefix
func (p *ShellCompletionProvider) getAvailableCommands(prefix string) []string {
	// Use a map to avoid duplicates
	commands := make(map[string]bool)

	// First, add shell aliases
	aliasCompletions := p.getAliasCompletions(prefix)
	for _, alias := range aliasCompletions {
		commands[alias] = true
	}

	// Then, get PATH from environment for system commands
	pathEnv := os.Getenv("PATH")
	if pathEnv != "" {
		// Split PATH into directories
		pathDirs := strings.Split(pathEnv, string(os.PathListSeparator))

		// Search each directory in PATH
		for _, dir := range pathDirs {
			entries, err := osReadDir(dir)
			if err != nil {
				continue // Skip directories we can't read
			}

			for _, entry := range entries {
				// Only consider regular files that are executable
				if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
					// Check if file is executable (this is a simplified check)
					// In a real implementation, you'd want to check file permissions
					commands[entry.Name()] = true
				}
			}
		}
	}

	// Convert map to sorted slice
	var completions []string
	for cmd := range commands {
		completions = append(completions, cmd)
	}

	// Sort alphabetically for consistent ordering
	sort.Strings(completions)
	return completions
}

// getAliasCompletions returns shell aliases that match the given prefix
func (p *ShellCompletionProvider) getAliasCompletions(prefix string) []string {
	if p.Runner == nil {
		return []string{}
	}

	// Use reflection to access the unexported alias field
	runnerValue := reflect.ValueOf(p.Runner).Elem()
	aliasField := runnerValue.FieldByName("alias")

	if !aliasField.IsValid() || aliasField.IsNil() {
		return []string{}
	}

	// The alias field is a map[string]interp.alias
	// We need to iterate over the keys (alias names)
	var completions []string

	// Get the map keys using reflection
	for _, key := range aliasField.MapKeys() {
		aliasName := key.String()
		if strings.HasPrefix(aliasName, prefix) {
			completions = append(completions, aliasName)
		}
	}

	// Sort alphabetically for consistent ordering
	sort.Strings(completions)
	return completions
}

// getBuiltinCommandCompletions returns completions for built-in commands starting with @!
func (p *ShellCompletionProvider) getBuiltinCommandCompletions(prefix string) []string {
	builtinCommands := []string{
		"new",
		"tokens",
		"subagents",
		"reload-subagents",
		"subagent-info",
	}

	var completions []string
	prefixAfterBang := strings.TrimPrefix(prefix, "@!")

	for _, cmd := range builtinCommands {
		if strings.HasPrefix(cmd, prefixAfterBang) {
			completions = append(completions, "@!"+cmd)
		}
	}

	// Sort alphabetically for consistent ordering
	sort.Strings(completions)
	return completions
}

// getSubagentCompletions returns completions for subagents starting with @
func (p *ShellCompletionProvider) getSubagentCompletions(prefix string) []string {
	// If no subagent provider is available, return no completions
	if p.SubagentProvider == nil {
		return []string{}
	}

	var completions []string
	prefixAfterAt := strings.TrimPrefix(prefix, "@")

	// Get all available subagents
	subagents := p.SubagentProvider.GetAllSubagents()

	for id, subagent := range subagents {
		// Match against both ID and name
		if strings.HasPrefix(id, prefixAfterAt) {
			completions = append(completions, "@"+id)
		} else if strings.HasPrefix(strings.ToLower(subagent.Name), strings.ToLower(prefixAfterAt)) {
			// Also match against display name (case-insensitive)
			completions = append(completions, "@"+id)
		}
	}

	// Sort alphabetically for consistent ordering
	sort.Strings(completions)
	return completions
}

// GetHelpInfo returns help information for special commands like @!, @/, and @
func (p *ShellCompletionProvider) GetHelpInfo(line string, pos int) string {
	// Get the current word being completed
	start, end := p.getCurrentWordBoundary(line, pos)
	if start < 0 || end < 0 {
		return ""
	}

	currentWord := line[start:end]

	// Check if the current word starts with @! (agent controls)
	if strings.HasPrefix(currentWord, "@!") {
		command := strings.TrimPrefix(currentWord, "@!")
		return p.getBuiltinCommandHelp(command)
	}

	// Check if the current word starts with @/ (macros)
	if strings.HasPrefix(currentWord, "@/") {
		macroName := strings.TrimPrefix(currentWord, "@/")
		return p.getMacroHelp(macroName)
	}

	// Check if the current word starts with @ (subagents)
	if strings.HasPrefix(currentWord, "@") && !strings.HasPrefix(currentWord, "@/") && !strings.HasPrefix(currentWord, "@!") {
		subagentName := strings.TrimPrefix(currentWord, "@")
		return p.getSubagentHelp(subagentName)
	}

	// Also check if we're at the beginning of a potential prefix
	if start > 0 {
		// Find the start of the word that might contain our prefix
		wordStart := start
		for wordStart > 0 && !unicode.IsSpace(rune(line[wordStart-1])) {
			wordStart--
		}

		potentialWord := line[wordStart:end]
		if strings.HasPrefix(potentialWord, "@!") {
			command := strings.TrimPrefix(potentialWord, "@!")
			return p.getBuiltinCommandHelp(command)
		} else if strings.HasPrefix(potentialWord, "@/") {
			macroName := strings.TrimPrefix(potentialWord, "@/")
			return p.getMacroHelp(macroName)
		} else if strings.HasPrefix(potentialWord, "@") && !strings.HasPrefix(potentialWord, "@/") && !strings.HasPrefix(potentialWord, "@!") {
			subagentName := strings.TrimPrefix(potentialWord, "@")
			return p.getSubagentHelp(subagentName)
		}
	}

	return ""
}

// getBuiltinCommandHelp returns help information for built-in commands
func (p *ShellCompletionProvider) getBuiltinCommandHelp(command string) string {
	switch command {
	case "new":
		return "**@!new** - Start a new chat session with the agent\n\nThis command resets the conversation history and starts fresh."
	case "tokens":
		return "**@!tokens** - Display token usage statistics\n\nShows information about token consumption for the current chat session."
	case "subagents":
		return "**@!subagents** - List all available subagents and modes\n\nDisplays all configured Claude-style subagents and Roo Code-style modes with their descriptions and capabilities."
	case "reload-subagents":
		return "**@!reload-subagents** - Reload subagent configurations from disk\n\nRefreshes the subagent configurations by rescanning the .claude/agents/ and .roo/modes/ directories."
	case "subagent-info":
		return "**@!subagent-info <name>** - Show detailed information about a subagent\n\nDisplays comprehensive information about a specific subagent including tools, file restrictions, and configuration."
	case "":
		return "**Agent Controls** - Built-in commands for managing the agent\n\nAvailable commands:\n• **@!new** - Start a new chat session\n• **@!tokens** - Show token usage statistics\n• **@!subagents** - List available subagents\n• **@!reload-subagents** - Reload subagent configurations\n• **@!subagent-info <name>** - Show subagent details"
	default:
		// Check for partial matches
		builtinCommands := []string{"new", "tokens", "subagents", "reload-subagents", "subagent-info"}
		for _, cmd := range builtinCommands {
			if strings.HasPrefix(cmd, command) {
				// Partial match, show general help
				return "**Agent Controls** - Built-in commands for managing the agent\n\nAvailable commands:\n• **@!new** - Start a new chat session\n• **@!tokens** - Show token usage statistics\n• **@!subagents** - List available subagents\n• **@!reload-subagents** - Reload subagent configurations\n• **@!subagent-info <name>** - Show subagent details"
			}
		}
		return ""
	}
}

// getMacroHelp returns help information for macros
func (p *ShellCompletionProvider) getMacroHelp(macroName string) string {
	var macrosStr string
	if p.Runner != nil {
		macrosStr = p.Runner.Vars["GSH_AGENT_MACROS"].String()
	} else {
		// Fallback to environment variable for testing
		macrosStr = os.Getenv("GSH_AGENT_MACROS")
	}

	if macrosStr == "" {
		if macroName == "" {
			return "**Chat Macros** - Quick shortcuts for common agent messages\n\nNo macros are currently configured."
		}
		return ""
	}

	var macros map[string]interface{}
	if err := json.Unmarshal([]byte(macrosStr), &macros); err != nil {
		return ""
	}

	if macroName == "" {
		// Show general macro help
		var macroList []string
		for name := range macros {
			macroList = append(macroList, "• **@/"+name+"**")
		}
		sort.Strings(macroList)

		if len(macroList) == 0 {
			return "**Chat Macros** - Quick shortcuts for common agent messages\n\nNo macros are currently configured."
		}

		return "**Chat Macros** - Quick shortcuts for common agent messages\n\nAvailable macros:\n" + strings.Join(macroList, "\n")
	}

	// Check for exact match first
	if message, ok := macros[macroName]; ok {
		if msgStr, ok := message.(string); ok {
			return fmt.Sprintf("**@/%s** - Chat macro\n\n**Expands to:**\n%s", macroName, msgStr)
		}
	}

	// Check for partial matches
	var matches []string
	for name, message := range macros {
		if strings.HasPrefix(name, macroName) {
			if msgStr, ok := message.(string); ok {
				matches = append(matches, fmt.Sprintf("• **@/%s** - %s", name, msgStr))
			}
		}
	}

	if len(matches) > 0 {
		sort.Strings(matches)
		return "**Chat Macros** - Matching macros:\n\n" + strings.Join(matches, "\n")
	}

	return ""
}

// getSubagentHelp returns help information for subagents
func (p *ShellCompletionProvider) getSubagentHelp(subagentName string) string {
	// If no subagent manager is available, return generic help
	if p.SubagentProvider == nil {
		if subagentName == "" {
			return "**Subagents** - Specialized AI assistants with specific roles\n\nNo subagent manager configured. Use @<subagent-name> to invoke a subagent."
		}
		return ""
	}

	// Get all available subagents
	subagents := p.SubagentProvider.GetAllSubagents()

	if subagentName == "" {
		// Show general subagent help
		if len(subagents) == 0 {
			return "**Subagents** - Specialized AI assistants with specific roles\n\nNo subagents are currently configured."
		}

		var subagentList []string
		for id, subagent := range subagents {
			description := subagent.Description
			if description == "" {
				description = "No description available"
			}
			subagentList = append(subagentList, fmt.Sprintf("• **@%s** - %s", id, description))
		}
		sort.Strings(subagentList)

		return "**Subagents** - Specialized AI assistants with specific roles\n\nAvailable subagents:\n" + strings.Join(subagentList, "\n")
	}

	// Check for exact match first
	if subagent, ok := subagents[subagentName]; ok {
		var toolsStr string
		if len(subagent.AllowedTools) > 0 {
			toolsStr = fmt.Sprintf("\n**Tools:** %v", subagent.AllowedTools)
		}

		var fileRegexStr string
		if subagent.FileRegex != "" {
			fileRegexStr = fmt.Sprintf("\n**File Access:** %s", subagent.FileRegex)
		}

		var modelStr string
		if subagent.Model != "" && subagent.Model != "inherit" {
			modelStr = fmt.Sprintf("\n**Model:** %s", subagent.Model)
		}

		description := subagent.Description
		if description == "" {
			description = "No description available"
		}

		return fmt.Sprintf("**@%s** - %s\n\n%s%s%s%s",
			subagentName, subagent.Name, description, toolsStr, fileRegexStr, modelStr)
	}

	// Check for partial matches by ID
	var matches []string
	for id, subagent := range subagents {
		if strings.HasPrefix(id, subagentName) {
			description := subagent.Description
			if description == "" {
				description = "No description available"
			}
			matches = append(matches, fmt.Sprintf("• **@%s** - %s", id, description))
		}
	}

	// Also check for matches by name (case-insensitive)
	for id, subagent := range subagents {
		if strings.HasPrefix(strings.ToLower(subagent.Name), strings.ToLower(subagentName)) &&
			!strings.HasPrefix(id, subagentName) { // Don't duplicate ID matches
			description := subagent.Description
			if description == "" {
				description = "No description available"
			}
			matches = append(matches, fmt.Sprintf("• **@%s** (%s) - %s", id, subagent.Name, description))
		}
	}

	if len(matches) > 0 {
		sort.Strings(matches)
		return "**Subagents** - Matching subagents:\n\n" + strings.Join(matches, "\n")
	}

	return ""
}
