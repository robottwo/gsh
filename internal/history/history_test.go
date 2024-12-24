package history

import (
	"testing"

	"go.uber.org/zap"
)

func Test(t *testing.T) {
	logger := zap.NewNop()

	historyManager, err := NewHistoryManager(":memory:", logger)
	if err != nil {
		t.Errorf("Failed to create history manager: %v", err)
	}

	entry, err := historyManager.StartCommand("echo hello")
	if err != nil {
		t.Errorf("Failed to start command: %v", err)
	}
	if entry.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be set")
	}
	if entry.UpdatedAt.IsZero() {
		t.Errorf("Expected UpdatedAt to be set")
	}

	entry, err = historyManager.FinishCommand(entry, "hello\n", "", 0)
	if err != nil {
		t.Errorf("Failed to finish command: %v", err)
	}

	entry, err = historyManager.StartCommand("echo world")
	if err != nil {
		t.Errorf("Failed to start command: %v", err)
	}

	entry, err = historyManager.FinishCommand(entry, "world\n", "", 0)
	if err != nil {
		t.Errorf("Failed to finish command: %v", err)
	}

	entries, err := historyManager.GetRecentEntries(3)
	if err != nil {
		t.Errorf("Failed to get recent entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	if entries[0].Command != "echo world" {
		t.Errorf("Expected most recent command to be 'echo world', got '%s'", entries[0].Command)
	}
}
