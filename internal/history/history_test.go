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

	entry, err := historyManager.StartCommand("echo hello", "/")
	if err != nil {
		t.Errorf("Failed to start command: %v", err)
	}
	if entry.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be set")
	}
	if entry.UpdatedAt.IsZero() {
		t.Errorf("Expected UpdatedAt to be set")
	}

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

	if len(allEntries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(allEntries))
	}

	if allEntries[0].Command != "echo hello" {
		t.Errorf("Expected most recent command to be 'echo hello', got '%s'", allEntries[0].Command)
	}

	targetEntries, err := historyManager.GetRecentEntries("/", 3)
	if len(targetEntries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(allEntries))
	}

	nonTargetEntries, err := historyManager.GetRecentEntries("/tmp", 3)
	if len(nonTargetEntries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(allEntries))
	}
}
