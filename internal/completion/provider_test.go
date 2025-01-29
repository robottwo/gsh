package completion

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"mvdan.cc/sh/v3/interp"
)

// Mock getFileCompletions for testing
var mockGetFileCompletions fileCompleter = func(prefix, currentDirectory string) []string {
	switch prefix {
	case "some/pa":
		return []string{"some/path.txt", "some/path2.txt"}
	case "/usr/local/b":
		return []string{"/usr/local/bin", "/usr/local/bin/"}
	case "'my documents/som":
		return []string{"my documents/something.txt", "my documents/somefile.txt"}
	case "":
		// Empty prefix means list everything in current directory
		return []string{"folder1/", "folder2/", "file1.txt", "file2.txt"}
	case "foo/bar/b":
		return []string{"foo/bar/baz", "foo/bar/bin"}
	case "other/path/te":
		return []string{"other/path/test.txt", "other/path/temp.txt"}
	default:
		// No match found
		return []string{}
	}
}

// mockCompletionManager mocks the CompletionManager for testing
type mockCompletionManager struct {
	mock.Mock
}

func (m *mockCompletionManager) GetSpec(command string) (CompletionSpec, bool) {
	args := m.Called(command)
	return args.Get(0).(CompletionSpec), args.Bool(1)
}

func (m *mockCompletionManager) ExecuteCompletion(ctx context.Context, runner *interp.Runner, spec CompletionSpec, args []string) ([]string, error) {
	callArgs := m.Called(ctx, runner, spec, args)
	return callArgs.Get(0).([]string), callArgs.Error(1)
}

func TestGetCompletions(t *testing.T) {
	// Replace getFileCompletions with mock for testing
	origGetFileCompletions := getFileCompletions
	getFileCompletions = mockGetFileCompletions
	defer func() {
		getFileCompletions = origGetFileCompletions
	}()
	runner := &interp.Runner{}
	manager := &mockCompletionManager{}
	provider := NewShellCompletionProvider(manager, runner)

	tests := []struct {
		name     string
		line     string
		pos      int
		setup    func()
		expected []string
	}{
		{
			name: "empty line returns no completions",
			line: "",
			pos:  0,
			setup: func() {
				// no setup needed
			},
			expected: []string{},
		},
		{
			name: "command with no completion spec returns no completions",
			line: "unknown-command arg1",
			pos:  20,
			setup: func() {
				manager.On("GetSpec", "unknown-command").Return(CompletionSpec{}, false)
			},
			expected: []string{},
		},
		{
			name: "command with word list completion returns suggestions",
			line: "git ch",
			pos:  6,
			setup: func() {
				spec := CompletionSpec{
					Command: "git",
					Type:    WordListCompletion,
					Value:   "checkout cherry-pick",
				}
				manager.On("GetSpec", "git").Return(spec, true)
				manager.On("ExecuteCompletion", mock.Anything, runner, spec, []string{"git", "ch"}).
					Return([]string{"checkout", "cherry-pick"}, nil)
			},
			expected: []string{"checkout", "cherry-pick"},
		},
		{
			name: "cursor position in middle of line only uses text up to cursor",
			line: "git checkout master",
			pos:  6, // cursor after "git ch"
			setup: func() {
				spec := CompletionSpec{
					Command: "git",
					Type:    WordListCompletion,
					Value:   "checkout cherry-pick",
				}
				manager.On("GetSpec", "git").Return(spec, true)
				manager.On("ExecuteCompletion", mock.Anything, runner, spec, []string{"git", "ch"}).
					Return([]string{"checkout", "cherry-pick"}, nil)
			},
			expected: []string{"checkout", "cherry-pick"},
		},
		{
			name: "file completion preserves command and path prefix",
			line: "cat some/pa",
			pos:  11,
			setup: func() {
				manager.On("GetSpec", "cat").Return(CompletionSpec{}, false)
			},
			expected: []string{"cat some/path.txt", "cat some/path2.txt"},
		},
		{
			name: "file completion with multiple path segments",
			line: "vim /usr/local/bi",
			pos:  16,
			setup: func() {
				manager.On("GetSpec", "vim").Return(CompletionSpec{}, false)
			},
			expected: []string{"vim /usr/local/bin", "vim /usr/local/bin/"},
		},
		{
			name: "file completion with spaces in path",
			line: "less 'my documents/some",
			pos:  22,
			setup: func() {
				manager.On("GetSpec", "less").Return(CompletionSpec{}, false)
			},
			expected: []string{"less \"my documents/something.txt\"", "less \"my documents/somefile.txt\""},
		},
		{
			name: "file completion after command with space",
			line: "cd ",
			pos:  3,
			setup: func() {
				manager.On("GetSpec", "cd").Return(CompletionSpec{}, false)
			},
			expected: []string{"cd folder1/", "cd folder2/", "cd file1.txt", "cd file2.txt"},
		},
		{
			name: "file completion after command with multiple spaces",
			line: "cd   ",
			pos:  5,
			setup: func() {
				manager.On("GetSpec", "cd").Return(CompletionSpec{}, false)
			},
			expected: []string{"cd folder1/", "cd folder2/", "cd file1.txt", "cd file2.txt"},
		},
		{
			name: "file completion with multiple path segments should only replace last segment",
			line: "ls foo/bar/b",
			pos:  12,
			setup: func() {
				manager.On("GetSpec", "ls").Return(CompletionSpec{}, false)
			},
			expected: []string{"ls foo/bar/baz", "ls foo/bar/bin"},
		},
		{
			name: "file completion with multiple arguments should preserve earlier arguments",
			line: "ls some/path other/path/te",
			pos:  26,
			setup: func() {
				manager.On("GetSpec", "ls").Return(CompletionSpec{}, false)
			},
			expected: []string{"ls some/path other/path/test.txt", "ls some/path other/path/temp.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.ExpectedCalls = nil
			manager.Calls = nil
			tt.setup()

			result := provider.GetCompletions(tt.line, tt.pos)
			assert.Equal(t, tt.expected, result)
			manager.AssertExpectations(t)
		})
	}
}
