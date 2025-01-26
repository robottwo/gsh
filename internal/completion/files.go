package completion

import (
	"os"
	"path/filepath"
	"strings"
)

// fileCompleter is the function type for file completion
type fileCompleter func(prefix string, currentDirectory string) []string

// getFileCompletions is the default implementation of file completion
var getFileCompletions fileCompleter = func(prefix string, currentDirectory string) []string {
	prefixIsAbs := filepath.IsAbs(prefix)
	// If prefix is empty, use current directory
	dir := currentDirectory
	filePrefix := ""

	if prefix != "" {
		// If prefix is absolute path, use it as is
		if prefixIsAbs {
			dir = filepath.Dir(prefix)
		} else {
			// For relative paths, join with current directory
			fullPath := filepath.Join(currentDirectory, prefix)
			dir = filepath.Dir(fullPath)
		}
		filePrefix = filepath.Base(prefix)

		// If the prefix ends with '/', we're looking for contents of that directory
		if strings.HasSuffix(prefix, "/") {
			if prefixIsAbs {
				dir = prefix
			} else {
				dir = filepath.Join(currentDirectory, prefix)
			}
			filePrefix = ""
		}
	}

	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}

	// Filter and format matches
	var matches []string = make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, filePrefix) {
			continue
		}

		// Build full path
		var fullPath string
		if prefixIsAbs {
			fullPath = filepath.Join(dir, name)
		} else {
			// For relative paths, make them relative to current directory
			relPath := filepath.Join(filepath.Dir(prefix), name)
			if relPath == "." {
				relPath = name
			}
			fullPath = relPath
		}

		// Add trailing slash for directories
		if entry.IsDir() {
			fullPath += "/"
		}

		matches = append(matches, fullPath)
	}

	return matches
}
