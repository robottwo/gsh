package shellinput

// CompletionProvider is the interface that provides completion suggestions
type CompletionProvider interface {
	// GetCompletions returns a list of completion suggestions for the current input
	// line and cursor position
	GetCompletions(line string, pos int) []string
}

// completionState tracks the state of completion suggestions
type completionState struct {
	active      bool
	suggestions []string
	selected    int
	prefix      string // the part of the word being completed
	startPos    int    // where in the input the completion should be inserted
}

func (cs *completionState) reset() {
	cs.active = false
	cs.suggestions = nil
	cs.selected = -1
	cs.prefix = ""
	cs.startPos = 0
}

func (cs *completionState) nextSuggestion() string {
	if !cs.active || len(cs.suggestions) == 0 {
		return ""
	}
	cs.selected = (cs.selected + 1) % len(cs.suggestions)
	return cs.suggestions[cs.selected]
}

func (cs *completionState) prevSuggestion() string {
	if !cs.active || len(cs.suggestions) == 0 {
		return ""
	}
	cs.selected--
	if cs.selected < 0 {
		cs.selected = len(cs.suggestions) - 1
	}
	return cs.suggestions[cs.selected]
}

func (cs *completionState) currentSuggestion() string {
	if !cs.active || cs.selected < 0 || cs.selected >= len(cs.suggestions) {
		return ""
	}
	return cs.suggestions[cs.selected]
}

