package completion

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileCompletions(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "completion_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some test files and directories
	files := []string{
		"file1.txt",
		"file2.txt",
		"folder1/",
		"folder2/",
		"folder1/inside.txt",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if filepath.Ext(path) == "" {
			// It's a directory
			err = os.MkdirAll(path, 0755)
		} else {
			// It's a file
			err = os.MkdirAll(filepath.Dir(path), 0755)
			if err == nil {
				err = os.WriteFile(path, []byte("test"), 0644)
			}
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name        string
		prefix      string
		currentDir  string
		expected    []string
		shouldMatch bool // true for exact match, false for contains
	}{
		{
			name:        "empty prefix lists all files",
			prefix:      "",
			currentDir:  tmpDir,
			expected:    []string{"file1.txt", "file2.txt", "folder1/", "folder2/"},
			shouldMatch: true,
		},
		{
			name:        "prefix matches start of filename",
			prefix:      "file",
			currentDir:  tmpDir,
			expected:    []string{"file1.txt", "file2.txt"},
			shouldMatch: true,
		},
		{
			name:        "prefix matches directories",
			prefix:      "folder",
			currentDir:  tmpDir,
			expected:    []string{"folder1/", "folder2/"},
			shouldMatch: true,
		},
		{
			name:        "absolute path prefix",
			prefix:      filepath.Join(tmpDir, "folder1") + "/",
			currentDir:  "/some/other/dir",
			expected:    []string{filepath.Join(tmpDir, "folder1/inside.txt")},
			shouldMatch: true,
		},
		{
			name:        "relative path in subdirectory",
			prefix:      "folder1/i",
			currentDir:  tmpDir,
			expected:    []string{"folder1/inside.txt"},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := getFileCompletions(tt.prefix, tt.currentDir)
			if tt.shouldMatch {
				assert.ElementsMatch(t, tt.expected, results)
			} else {
				for _, exp := range tt.expected {
					found := false
					for _, res := range results {
						if filepath.Base(res) == exp {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected to find %s in results, but got %s", exp, results)
				}
			}
		})
	}
}

