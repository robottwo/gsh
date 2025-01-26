package completion

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletionManager(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		manager := NewCompletionManager()

		// Test adding a spec
		spec := CompletionSpec{
			Command: "test-cmd",
			Type:    "W",
			Value:   "foo bar baz",
		}
		manager.AddSpec(spec)

		// Test getting the spec
		retrieved, exists := manager.GetSpec("test-cmd")
		assert.True(t, exists)
		assert.Equal(t, spec, retrieved)

		// Test getting non-existent spec
		_, exists = manager.GetSpec("non-existent")
		assert.False(t, exists)

		// Test listing specs
		specs := manager.ListSpecs()
		assert.Len(t, specs, 1)
		assert.Equal(t, spec, specs[0])

		// Test removing spec
		manager.RemoveSpec("test-cmd")
		_, exists = manager.GetSpec("test-cmd")
		assert.False(t, exists)
		specs = manager.ListSpecs()
		assert.Empty(t, specs)
	})
}

func TestCompleteCommandHandler(t *testing.T) {
	t.Run("completion specifications", func(t *testing.T) {
		manager := NewCompletionManager()
		handler := NewCompleteCommandHandler(manager)

		// Create a mock next handler that just returns nil
		nextHandler := func(ctx context.Context, args []string) error {
			return nil
		}

		wrappedHandler := handler(nextHandler)

		// Test adding word list completion
		var captured []string
		oldPrintf := printf
		printf = func(format string, a ...any) (int, error) {
			captured = append(captured, fmt.Sprintf(format, a...))
			return len(format), nil
		}
		defer func() { printf = oldPrintf }()

		// Test word list completion
		err := wrappedHandler(context.Background(), []string{"complete", "-W", "foo bar", "mycmd"})
		assert.NoError(t, err)

		// Verify the word list spec was added correctly
		spec, exists := manager.GetSpec("mycmd")
		assert.True(t, exists)
		assert.Equal(t, WordListCompletion, spec.Type)
		assert.Equal(t, "foo bar", spec.Value)

		// Test function completion
		err = wrappedHandler(context.Background(), []string{"complete", "-F", "_mycmd_completion", "mycmd2"})
		assert.NoError(t, err)

		// Verify the function spec was added correctly
		spec, exists = manager.GetSpec("mycmd2")
		assert.True(t, exists)
		assert.Equal(t, FunctionCompletion, spec.Type)
		assert.Equal(t, "_mycmd_completion", spec.Value)

		// Test complete -p
		captured = []string{}
		err = wrappedHandler(context.Background(), []string{"complete", "-p"})
		assert.NoError(t, err)
		assert.Contains(t, captured, "complete -W \"foo bar\" mycmd\n")
		assert.Contains(t, captured, "complete -F _mycmd_completion mycmd2\n")

		// Test complete -p mycmd
		captured = []string{}
		err = wrappedHandler(context.Background(), []string{"complete", "-p", "mycmd"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"complete -W \"foo bar\" mycmd\n"}, captured)

		// Test complete -r mycmd
		err = wrappedHandler(context.Background(), []string{"complete", "-r", "mycmd"})
		assert.NoError(t, err)
		_, exists = manager.GetSpec("mycmd")
		assert.False(t, exists)
	})

	t.Run("error cases", func(t *testing.T) {
		manager := NewCompletionManager()
		handler := NewCompleteCommandHandler(manager)
		nextHandler := func(ctx context.Context, args []string) error {
			return nil
		}
		wrappedHandler := handler(nextHandler)

		testCases := []struct {
			name    string
			args    []string
			wantErr string
		}{
			{
				name:    "missing word list",
				args:    []string{"complete", "-W"},
				wantErr: "option -W requires a word list",
			},
			{
				name:    "unknown option",
				args:    []string{"complete", "-x", "mycmd"},
				wantErr: "unknown option: -x",
			},
			{
				name:    "no command specified",
				args:    []string{"complete", "-W", "foo bar"},
				wantErr: "no command specified",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := wrappedHandler(context.Background(), tc.args)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			})
		}
	})

	t.Run("pass through non-complete commands", func(t *testing.T) {
		manager := NewCompletionManager()
		handler := NewCompleteCommandHandler(manager)

		nextCalled := false
		nextHandler := func(ctx context.Context, args []string) error {
			nextCalled = true
			return nil
		}
		wrappedHandler := handler(nextHandler)

		err := wrappedHandler(context.Background(), []string{"echo", "hello"})
		assert.NoError(t, err)
		assert.True(t, nextCalled)
	})
}

