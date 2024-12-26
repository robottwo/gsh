package keys

type KeyCode int

const (
	KeyNull   KeyCode = ansiNUL
	KeyBreak  KeyCode = ansiETX
	KeyEnter  KeyCode = ansiCR
	KeyTab    KeyCode = ansiHT
	KeyEscape KeyCode = ansiESC

	// use ascii code for the following keys
	KeySpace            KeyCode = 32
	KeyExclamationMark  KeyCode = 33 // !
	KeyDoubleQuote      KeyCode = 34 // "
	KeyHash             KeyCode = 35 // #
	KeyDollar           KeyCode = 36 // $
	KeyPercent          KeyCode = 37 // %
	KeyAmpersand        KeyCode = 38 // &
	KeySingleQuote      KeyCode = 39 // '
	KeyOpenParentheses  KeyCode = 40 // (
	KeyCloseParentheses KeyCode = 41 // )
	KeyAsterisk         KeyCode = 42 // *
	KeyPlus             KeyCode = 43 // +
	KeyComma            KeyCode = 44 // ,
	KeyMinus            KeyCode = 45 // -
	KeyPeriod           KeyCode = 46 // .
	KeySlash            KeyCode = 47 // /

	Key0 KeyCode = 48
	Key1 KeyCode = 49
	Key2 KeyCode = 50
	Key3 KeyCode = 51
	Key4 KeyCode = 52
	Key5 KeyCode = 53
	Key6 KeyCode = 54
	Key7 KeyCode = 55
	Key8 KeyCode = 56
	Key9 KeyCode = 57

	KeyColon        KeyCode = 58
	KeySemicolon    KeyCode = 59
	KeyLessThan     KeyCode = 60
	KeyEquals       KeyCode = 61
	KeyGreaterThan  KeyCode = 62
	KeyQuestionMark KeyCode = 63
	KeyAt           KeyCode = 64

	// skip 65-90 for upper case A-Z

	KeyOpenBracket  KeyCode = 91 // [
	KeyBackslash    KeyCode = 92 // \
	KeyCloseBracket KeyCode = 93 // ]
	KeyCaret        KeyCode = 94 // ^
	KeyUnderscore   KeyCode = 95 // _
	KeyGraveAccent  KeyCode = 96 // `

	KeyA KeyCode = 97
	KeyB KeyCode = 98
	KeyC KeyCode = 99
	KeyD KeyCode = 100
	KeyE KeyCode = 101
	KeyF KeyCode = 102
	KeyG KeyCode = 103
	KeyH KeyCode = 104
	KeyI KeyCode = 105
	KeyJ KeyCode = 106
	KeyK KeyCode = 107
	KeyL KeyCode = 108
	KeyM KeyCode = 109
	KeyN KeyCode = 110
	KeyO KeyCode = 111
	KeyP KeyCode = 112
	KeyQ KeyCode = 113
	KeyR KeyCode = 114
	KeyS KeyCode = 115
	KeyT KeyCode = 116
	KeyU KeyCode = 117
	KeyV KeyCode = 118
	KeyW KeyCode = 119
	KeyX KeyCode = 120
	KeyY KeyCode = 121
	KeyZ KeyCode = 122

	KeyOpenCurlyBrace  KeyCode = 123 // {
	KeyVerticalBar     KeyCode = 124 // |
	KeyCloseCurlyBrace KeyCode = 125 // }
	KeyTilde           KeyCode = 126 // ~
	KeyBackspace       KeyCode = 127
)

const (
	keyRunes KeyCode = -(iota + 1)
	KeyUp
	KeyDown
	KeyRight
	KeyLeft
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDown
	KeyDelete
	KeyInsert
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
)

var keyNames = map[KeyCode]string{
	KeyNull:   "NUL",
	KeyBreak:  "BREAK",
	KeyEnter:  "ENTER",
	KeyTab:    "TAB",
	KeyEscape: "ESCAPE",

	KeySpace:            "SPACE",
	KeyExclamationMark:  "!",
	KeyDoubleQuote:      "\"",
	KeyHash:             "#",
	KeyDollar:           "$",
	KeyPercent:          "%",
	KeyAmpersand:        "&",
	KeySingleQuote:      "'",
	KeyOpenParentheses:  "(",
	KeyCloseParentheses: ")",
	KeyAsterisk:         "*",
	KeyPlus:             "+",
	KeyComma:            ",",
	KeyMinus:            "-",
	KeyPeriod:           ".",
	KeySlash:            "/",

	Key0: "0",
	Key1: "1",
	Key2: "2",
	Key3: "3",
	Key4: "4",
	Key5: "5",
	Key6: "6",
	Key7: "7",
	Key8: "8",
	Key9: "9",

	KeyColon:        ":",
	KeySemicolon:    ";",
	KeyLessThan:     "<",
	KeyEquals:       "=",
	KeyGreaterThan:  ">",
	KeyQuestionMark: "?",
	KeyAt:           "@",

	KeyOpenBracket:  "[",
	KeyBackslash:    "\\",
	KeyCloseBracket: "]",
	KeyCaret:        "^",
	KeyUnderscore:   "_",
	KeyGraveAccent:  "`",

	KeyA: "A",
	KeyB: "B",
	KeyC: "C",
	KeyD: "D",
	KeyE: "E",
	KeyF: "F",
	KeyG: "G",
	KeyH: "H",
	KeyI: "I",
	KeyJ: "J",
	KeyK: "K",
	KeyL: "L",
	KeyM: "M",
	KeyN: "N",
	KeyO: "O",
	KeyP: "P",
	KeyQ: "Q",
	KeyR: "R",
	KeyS: "S",
	KeyT: "T",
	KeyU: "U",
	KeyV: "V",
	KeyW: "W",
	KeyX: "X",
	KeyY: "Y",
	KeyZ: "Z",

	KeyOpenCurlyBrace:  "{",
	KeyVerticalBar:     "|",
	KeyCloseCurlyBrace: "}",
	KeyTilde:           "~",
	KeyBackspace:       "BACKSPACE",

	KeyUp:     "UP",
	KeyDown:   "DOWN",
	KeyRight:  "RIGHT",
	KeyLeft:   "LEFT",
	KeyHome:   "HOME",
	KeyEnd:    "END",
	KeyPgUp:   "PGUP",
	KeyPgDown: "PGDOWN",
	KeyDelete: "DELETE",
	KeyInsert: "INSERT",

	KeyF1:  "F1",
	KeyF2:  "F2",
	KeyF3:  "F3",
	KeyF4:  "F4",
	KeyF5:  "F5",
	KeyF6:  "F6",
	KeyF7:  "F7",
	KeyF8:  "F8",
	KeyF9:  "F9",
	KeyF10: "F10",
	KeyF11: "F11",
	KeyF12: "F12",
	KeyF13: "F13",
	KeyF14: "F14",
	KeyF15: "F15",
	KeyF16: "F16",
	KeyF17: "F17",
	KeyF18: "F18",
	KeyF19: "F19",
	KeyF20: "F20",
}

func (k KeyCode) String() string {
	if name, ok := keyNames[k]; ok {
		return name
	}
	return "UNKNOWN"
}
