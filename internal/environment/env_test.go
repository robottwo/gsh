package environment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

func TestAppendToAuthorizedCommands(t *testing.T) {
	// Create a temporary config directory for testing
	tempConfigDir := filepath.Join(os.TempDir(), "gsh_test_config")
	tempAuthorizedFile := filepath.Join(tempConfigDir, "authorized_commands")

	// Override the global variables for testing
	oldConfigDir := configDir
	oldAuthorizedFile := authorizedCommandsFile
	configDir = tempConfigDir
	authorizedCommandsFile = tempAuthorizedFile
	defer func() {
		configDir = oldConfigDir
		authorizedCommandsFile = oldAuthorizedFile
		os.RemoveAll(tempConfigDir)
	}()

	// Test appending a command
	err := AppendToAuthorizedCommands("ls.*")
	assert.NoError(t, err)

	// Check if file was created
	_, err = os.Stat(authorizedCommandsFile)
	assert.NoError(t, err)

	// Check file contents
	content, err := os.ReadFile(authorizedCommandsFile)
	assert.NoError(t, err)
	assert.Equal(t, "ls.*\n", string(content))

	// Test appending another command
	err = AppendToAuthorizedCommands("git.*")
	assert.NoError(t, err)

	// Check file contents again
	content, err = os.ReadFile(authorizedCommandsFile)
	assert.NoError(t, err)
	assert.Equal(t, "ls.*\ngit.*\n", string(content))
}

func TestAppendToAuthorizedCommandsSecurePermissions(t *testing.T) {
	// Create a temporary config directory for testing
	tempConfigDir := filepath.Join(os.TempDir(), "gsh_test_config_secure")
	tempAuthorizedFile := filepath.Join(tempConfigDir, "authorized_commands")

	// Override the global variables for testing
	oldConfigDir := configDir
	oldAuthorizedFile := authorizedCommandsFile
	configDir = tempConfigDir
	authorizedCommandsFile = tempAuthorizedFile
	defer func() {
		configDir = oldConfigDir
		authorizedCommandsFile = oldAuthorizedFile
		os.RemoveAll(tempConfigDir)
	}()

	t.Run("New directory and file have secure permissions", func(t *testing.T) {
		// Test appending a command to a new file
		err := AppendToAuthorizedCommands("ls.*")
		assert.NoError(t, err)

		// Check directory permissions
		dirInfo, err := os.Stat(configDir)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0700), dirInfo.Mode()&0777, "Directory should have 0700 permissions")

		// Check file permissions
		fileInfo, err := os.Stat(authorizedCommandsFile)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), fileInfo.Mode()&0777, "File should have 0600 permissions")

		// Verify no group or other access
		assert.Equal(t, os.FileMode(0), fileInfo.Mode()&0077, "File should not be accessible by group or others")
	})

	t.Run("Existing insecure files get permissions fixed", func(t *testing.T) {
		// Clean up from previous test
		os.RemoveAll(tempConfigDir)

		// Create directory and file with insecure permissions
		err := os.MkdirAll(tempConfigDir, 0755) // Insecure directory permissions
		assert.NoError(t, err)

		err = os.WriteFile(tempAuthorizedFile, []byte("existing.*\n"), 0644) // Insecure file permissions
		assert.NoError(t, err)

		// Verify they start with insecure permissions
		dirInfo, err := os.Stat(tempConfigDir)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0755), dirInfo.Mode()&0777, "Directory should start with 0755 permissions")

		fileInfo, err := os.Stat(tempAuthorizedFile)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0644), fileInfo.Mode()&0777, "File should start with 0644 permissions")

		// Append to the existing file - this should fix permissions
		err = AppendToAuthorizedCommands("new.*")
		assert.NoError(t, err)

		// Check that permissions were fixed
		dirInfo, err = os.Stat(tempConfigDir)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0700), dirInfo.Mode()&0777, "Directory permissions should be fixed to 0700")

		fileInfo, err = os.Stat(tempAuthorizedFile)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), fileInfo.Mode()&0777, "File permissions should be fixed to 0600")

		// Verify content is correct
		content, err := os.ReadFile(tempAuthorizedFile)
		assert.NoError(t, err)
		assert.Equal(t, "existing.*\nnew.*\n", string(content))
	})

	t.Run("Permission errors are handled gracefully", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("Skipping permission error test when running as root")
		}

		// Clean up from previous test
		os.RemoveAll(tempConfigDir)

		// Create a directory we can't write to
		err := os.MkdirAll(tempConfigDir, 0555) // Read and execute only
		assert.NoError(t, err)

		// Try to append - this may or may not fail depending on the system
		// The important thing is that it doesn't panic
		err = AppendToAuthorizedCommands("test.*")
		// On some systems this might succeed, on others it might fail
		// We just want to ensure no panic occurs
		if err != nil {
			// If it fails, it should be a permission-related error
			assert.True(t,
				strings.Contains(err.Error(), "permission") ||
					strings.Contains(err.Error(), "failed to set") ||
					strings.Contains(err.Error(), "failed to open"),
				"Error should be permission-related: %v", err)
		}
	})
}

func TestLoadAuthorizedCommandsFromFile(t *testing.T) {
	// Create a temporary config directory for testing
	tempConfigDir := filepath.Join(os.TempDir(), "gsh_test_config")
	tempAuthorizedFile := filepath.Join(tempConfigDir, "authorized_commands")

	// Override the global variables for testing
	oldConfigDir := configDir
	oldAuthorizedFile := authorizedCommandsFile
	configDir = tempConfigDir
	authorizedCommandsFile = tempAuthorizedFile
	defer func() {
		configDir = oldConfigDir
		authorizedCommandsFile = oldAuthorizedFile
		os.RemoveAll(tempConfigDir)
	}()

	// Test with non-existent file
	patterns, err := LoadAuthorizedCommandsFromFile()
	assert.NoError(t, err)
	assert.Equal(t, []string{}, patterns)

	// Create file with some patterns
	err = os.MkdirAll(tempConfigDir, 0700)
	assert.NoError(t, err)

	err = AppendToAuthorizedCommands("ls.*")
	assert.NoError(t, err)

	err = AppendToAuthorizedCommands("git.*")
	assert.NoError(t, err)

	// Test loading patterns
	patterns, err = LoadAuthorizedCommandsFromFile()
	assert.NoError(t, err)
	assert.Equal(t, []string{"ls.*", "git.*"}, patterns)
}

func TestGetApprovedBashCommandRegex(t *testing.T) {
	// Create a temporary config directory for testing
	tempConfigDir := filepath.Join(os.TempDir(), "gsh_test_config")
	tempAuthorizedFile := filepath.Join(tempConfigDir, "authorized_commands")

	// Override the global variables for testing
	oldConfigDir := configDir
	oldAuthorizedFile := authorizedCommandsFile
	configDir = tempConfigDir
	authorizedCommandsFile = tempAuthorizedFile
	defer func() {
		configDir = oldConfigDir
		authorizedCommandsFile = oldAuthorizedFile
		os.RemoveAll(tempConfigDir)
	}()

	// Create logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create a test runner
	env := expand.ListEnviron(os.Environ()...)
	runner, err := interp.New(interp.Env(env))
	assert.NoError(t, err)

	// Test with no environment patterns and no file patterns
	patterns := GetApprovedBashCommandRegex(runner, logger)
	assert.Equal(t, []string{}, patterns)

	// Add file patterns
	err = os.MkdirAll(tempConfigDir, 0700)
	assert.NoError(t, err)

	err = AppendToAuthorizedCommands("ls.*")
	assert.NoError(t, err)

	err = AppendToAuthorizedCommands("git.*")
	assert.NoError(t, err)

	// Test with file patterns only
	patterns = GetApprovedBashCommandRegex(runner, logger)
	assert.Equal(t, []string{"ls.*", "git.*"}, patterns)
}
