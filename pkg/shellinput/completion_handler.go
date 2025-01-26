package shellinput

import (
	"unicode"
)

// getWordBoundary returns the start and end position of the word at the cursor
func (m *Model) getWordBoundary() (start, end int) {
	value := m.Value()
	if len(value) == 0 {
		return 0, 0
	}

	// Get cursor position
	pos := m.Position()

	// Find start of word
	start = pos
	for start > 0 && !unicode.IsSpace(rune(value[start-1])) {
		start--
	}

	// Find end of word
	end = pos
	for end < len(value) && !unicode.IsSpace(rune(value[end])) {
		end++
	}

	return start, end
}

// handleCompletion handles the TAB key press for completion
func (m *Model) handleCompletion() {
	if m.CompletionProvider == nil {
		return
	}

	if !m.completion.active {
		// Start a new completion
		start, _ := m.getWordBoundary()
		suggestions := m.CompletionProvider.GetCompletions(m.Value(), m.Position())
		if len(suggestions) == 0 {
			m.resetCompletion() // Ensure completion state is reset
			return
		}

		m.completion.active = true
		m.completion.suggestions = suggestions
		m.completion.selected = -1
		m.completion.prefix = m.Value()[start:m.Position()]
		m.completion.startPos = 0 // Always replace from the start for full completions
	}

	// Get next suggestion (this works for both initial and subsequent TAB presses)
	suggestion := m.completion.nextSuggestion()
	if suggestion == "" {
		return
	}

	// Apply the suggestion
	m.applySuggestion(suggestion)
}

// handleBackwardCompletion handles the Shift+TAB key press for completion
func (m *Model) handleBackwardCompletion() {
	if m.CompletionProvider == nil || !m.completion.active {
		return
	}

	suggestion := m.completion.prevSuggestion()
	if suggestion == "" {
		return
	}

	m.applySuggestion(suggestion)
}

// applySuggestion replaces the current word with the suggestion
func (m *Model) applySuggestion(suggestion string) {
	value := m.Value()
	if m.completion.startPos > len(value) {
		return
	}

	_, end := m.getWordBoundary()
	if end > len(value) {
		end = len(value)
	}

	// Replace the current word with the suggestion
	newValue := value[:m.completion.startPos] + suggestion
	if end < len(value) {
		newValue += value[end:]
	}
	m.SetValue(newValue)

	// Move cursor to end of inserted suggestion
	m.SetCursor(m.completion.startPos + len(suggestion))
}

// resetCompletion resets the completion state
func (m *Model) resetCompletion() {
	m.completion.reset()
}

