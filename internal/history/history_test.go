package history

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"go.uber.org/zap"
)

func Test(t *testing.T) {
	logger := zap.NewNop()

	historyManager, err := NewHistoryManager(":memory:", logger)
	assert.NoError(t, err, "Failed to create history manager")

	entry, err := historyManager.StartCommand("echo hello", "/")
	if err != nil {
		t.Errorf("Failed to start command: %v", err)
	}
	assert.False(t, entry.CreatedAt.IsZero(), "Expected CreatedAt to be set")
	assert.False(t, entry.UpdatedAt.IsZero(), "Expected UpdatedAt to be set")

	entry, err = historyManager.FinishCommand(entry, 0)
	if err != nil {
		t.Errorf("Failed to finish command: %v", err)
	}

	entry, err = historyManager.StartCommand("echo world", "/")
	if err != nil {
		t.Errorf("Failed to start command: %v", err)
	}

	entry, err = historyManager.FinishCommand(entry, 0)
	if err != nil {
		t.Errorf("Failed to finish command: %v", err)
	}

	allEntries, err := historyManager.GetRecentEntries("", 3)
	if err != nil {
		t.Errorf("Failed to get recent entries: %v", err)
	}

	assert.Len(t, allEntries, 2, "Expected 2 entries")

	assert.Equal(t, "echo hello", allEntries[0].Command, "Expected most recent command to be 'echo hello'")

	targetEntries, err := historyManager.GetRecentEntries("/", 3)
	assert.Len(t, targetEntries, 2, "Expected 2 entries")

	nonTargetEntries, err := historyManager.GetRecentEntries("/tmp", 3)
	assert.Len(t, nonTargetEntries, 0, "Expected 0 entries")
}
