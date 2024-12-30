package gline

import "github.com/atinylittleshell/gsh/pkg/gline/keys"

type Options struct {
	Keybinds map[Command][]keys.KeyPress
}

type Command int

const (
	CommandExecute Command = iota
	CommandBackspace
	CommandHistoryPrevious
	CommandHistoryNext
	CommandCursorForward
	CommandCursorBackward
	CommandCursorDeleteToBeginningOfLine
	CommandCursorDeleteToEndOfLine
	CommandCursorMoveToBeginningOfLine
	CommandCursorMoveToEndOfLine
)

func NewOptions() *Options {
	return &Options{
		Keybinds: map[Command][]keys.KeyPress{
			CommandExecute:                       {{Code: keys.KeyEnter}, {Code: keys.KeyM, Ctrl: true}},
			CommandBackspace:                     {{Code: keys.KeyBackspace}},
			CommandHistoryPrevious:               {{Code: keys.KeyUp}, {Code: keys.KeyP, Ctrl: true}},
			CommandHistoryNext:                   {{Code: keys.KeyDown}, {Code: keys.KeyN, Ctrl: true}},
			CommandCursorForward:                 {{Code: keys.KeyRight}, {Code: keys.KeyF, Ctrl: true}},
			CommandCursorBackward:                {{Code: keys.KeyLeft}, {Code: keys.KeyB, Ctrl: true}},
			CommandCursorDeleteToBeginningOfLine: {{Code: keys.KeyU, Ctrl: true}},
			CommandCursorDeleteToEndOfLine:       {{Code: keys.KeyK, Ctrl: true}},
			CommandCursorMoveToBeginningOfLine:   {{Code: keys.KeyA, Ctrl: true}},
			CommandCursorMoveToEndOfLine:         {{Code: keys.KeyE, Ctrl: true}},
		},
	}
}
