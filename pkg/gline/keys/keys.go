package keys

import "strings"

type KeyPress struct {
	Code  KeyCode
	Ctrl  bool
	Alt   bool
	Shift bool
}

// Sequence mappings.
var sequenceMapping = map[string]KeyPress{
	// ANSI control codes
	"\x00": {Code: KeyNull, Alt: false, Shift: false, Ctrl: false},
	"\x01": {Code: KeyA, Alt: false, Shift: false, Ctrl: true},
	"\x02": {Code: KeyB, Alt: false, Shift: false, Ctrl: true},
	"\x03": {Code: KeyBreak, Alt: false, Shift: false, Ctrl: false},
	"\x04": {Code: KeyD, Alt: false, Shift: false, Ctrl: true},
	"\x05": {Code: KeyE, Alt: false, Shift: false, Ctrl: true},
	"\x06": {Code: KeyF, Alt: false, Shift: false, Ctrl: true},
	"\x07": {Code: KeyG, Alt: false, Shift: false, Ctrl: true},
	"\x08": {Code: KeyBackspace, Alt: false, Shift: false, Ctrl: false},
	"\x09": {Code: KeyTab, Alt: false, Shift: false, Ctrl: false},
	"\x0a": {Code: KeyJ, Alt: false, Shift: false, Ctrl: true},
	"\x0b": {Code: KeyK, Alt: false, Shift: false, Ctrl: true},
	"\x0c": {Code: KeyL, Alt: false, Shift: false, Ctrl: true},
	"\x0d": {Code: KeyEnter, Alt: false, Shift: false, Ctrl: false},
	"\x0e": {Code: KeyN, Alt: false, Shift: false, Ctrl: true},
	"\x0f": {Code: KeyO, Alt: false, Shift: false, Ctrl: true},
	"\x10": {Code: KeyP, Alt: false, Shift: false, Ctrl: true},
	"\x11": {Code: KeyQ, Alt: false, Shift: false, Ctrl: true},
	"\x12": {Code: KeyR, Alt: false, Shift: false, Ctrl: true},
	"\x13": {Code: KeyS, Alt: false, Shift: false, Ctrl: true},
	"\x14": {Code: KeyT, Alt: false, Shift: false, Ctrl: true},
	"\x15": {Code: KeyU, Alt: false, Shift: false, Ctrl: true},
	"\x16": {Code: KeyV, Alt: false, Shift: false, Ctrl: true},
	"\x17": {Code: KeyW, Alt: false, Shift: false, Ctrl: true},
	"\x18": {Code: KeyX, Alt: false, Shift: false, Ctrl: true},
	"\x19": {Code: KeyY, Alt: false, Shift: false, Ctrl: true},
	"\x1a": {Code: KeyZ, Alt: false, Shift: false, Ctrl: true},
	"\x1b": {Code: KeyEscape, Alt: false, Shift: false, Ctrl: false},
	"\x1c": {Code: KeyBackslash, Alt: false, Shift: false, Ctrl: true},
	"\x1d": {Code: KeyCloseBracket, Alt: false, Shift: false, Ctrl: true},
	"\x1e": {Code: KeyCaret, Alt: false, Shift: false, Ctrl: true},
	"\x1f": {Code: KeyUnderscore, Alt: false, Shift: false, Ctrl: true},
	"\x7f": {Code: KeyBackspace, Alt: false, Shift: false, Ctrl: false},
	// Arrow keys
	"\x1b[A":    {Code: KeyUp, Alt: false, Shift: false, Ctrl: false},
	"\x1b[B":    {Code: KeyDown, Alt: false, Shift: false, Ctrl: false},
	"\x1b[C":    {Code: KeyRight, Alt: false, Shift: false, Ctrl: false},
	"\x1b[D":    {Code: KeyLeft, Alt: false, Shift: false, Ctrl: false},
	"\x1b[1;2A": {Code: KeyUp, Alt: false, Shift: true, Ctrl: false},
	"\x1b[1;2B": {Code: KeyDown, Alt: false, Shift: true, Ctrl: false},
	"\x1b[1;2C": {Code: KeyRight, Alt: false, Shift: true, Ctrl: false},
	"\x1b[1;2D": {Code: KeyLeft, Alt: false, Shift: true, Ctrl: false},
	"\x1b[OA":   {Code: KeyUp, Alt: false, Shift: true, Ctrl: false},
	"\x1b[OB":   {Code: KeyDown, Alt: false, Shift: true, Ctrl: false},
	"\x1b[OC":   {Code: KeyRight, Alt: false, Shift: true, Ctrl: false},
	"\x1b[OD":   {Code: KeyLeft, Alt: false, Shift: true, Ctrl: false},
	"\x1b[a":    {Code: KeyUp, Alt: false, Shift: true, Ctrl: false},
	"\x1b[b":    {Code: KeyDown, Alt: false, Shift: true, Ctrl: false},
	"\x1b[c":    {Code: KeyRight, Alt: false, Shift: true, Ctrl: false},
	"\x1b[d":    {Code: KeyLeft, Alt: false, Shift: true, Ctrl: false},

	"\x1b[1;3A": {Code: KeyUp, Alt: true, Shift: false, Ctrl: false},
	"\x1b[1;3B": {Code: KeyDown, Alt: true, Shift: false, Ctrl: false},
	"\x1b[1;3C": {Code: KeyRight, Alt: true, Shift: false, Ctrl: false},
	"\x1b[1;3D": {Code: KeyLeft, Alt: true, Shift: false, Ctrl: false},

	"\x1b[1;4A": {Code: KeyUp, Alt: true, Shift: true, Ctrl: false},
	"\x1b[1;4B": {Code: KeyDown, Alt: true, Shift: true, Ctrl: false},
	"\x1b[1;4C": {Code: KeyRight, Alt: true, Shift: true, Ctrl: false},
	"\x1b[1;4D": {Code: KeyLeft, Alt: true, Shift: true, Ctrl: false},

	"\x1b[1;5A": {Code: KeyUp, Alt: false, Shift: false, Ctrl: true},
	"\x1b[1;5B": {Code: KeyDown, Alt: false, Shift: false, Ctrl: true},
	"\x1b[1;5C": {Code: KeyRight, Alt: false, Shift: false, Ctrl: true},
	"\x1b[1;5D": {Code: KeyLeft, Alt: false, Shift: false, Ctrl: true},

	"\x1b[Oa": {Code: KeyUp, Alt: true, Shift: false, Ctrl: true},
	"\x1b[Ob": {Code: KeyDown, Alt: true, Shift: false, Ctrl: true},
	"\x1b[Oc": {Code: KeyRight, Alt: true, Shift: false, Ctrl: true},
	"\x1b[Od": {Code: KeyLeft, Alt: true, Shift: false, Ctrl: true},

	"\x1b[1;6A": {Code: KeyUp, Alt: false, Shift: true, Ctrl: true},
	"\x1b[1;6B": {Code: KeyDown, Alt: false, Shift: true, Ctrl: true},
	"\x1b[1;6C": {Code: KeyRight, Alt: false, Shift: true, Ctrl: true},
	"\x1b[1;6D": {Code: KeyLeft, Alt: false, Shift: true, Ctrl: true},

	"\x1b[1;7A": {Code: KeyUp, Alt: true, Shift: false, Ctrl: true},
	"\x1b[1;7B": {Code: KeyDown, Alt: true, Shift: false, Ctrl: true},
	"\x1b[1;7C": {Code: KeyRight, Alt: true, Shift: false, Ctrl: true},
	"\x1b[1;7D": {Code: KeyLeft, Alt: true, Shift: false, Ctrl: true},

	"\x1b[1;8A": {Code: KeyUp, Alt: true, Shift: true, Ctrl: true},
	"\x1b[1;8B": {Code: KeyDown, Alt: true, Shift: true, Ctrl: true},
	"\x1b[1;8C": {Code: KeyRight, Alt: true, Shift: true, Ctrl: true},
	"\x1b[1;8D": {Code: KeyLeft, Alt: true, Shift: true, Ctrl: true},

	// Others
	"\x1b[Z": {Code: KeyTab, Alt: false, Shift: true, Ctrl: false},

	"\x1b[2~":   {Code: KeyInsert, Alt: false, Shift: false, Ctrl: false},
	"\x1b[3;2~": {Code: KeyInsert, Alt: false, Shift: false, Ctrl: false},

	"\x1b[3~":   {Code: KeyDelete, Alt: false, Shift: false, Ctrl: false},
	"\x1b[3;3~": {Code: KeyDelete, Alt: false, Shift: false, Ctrl: false},

	"\x1b[5~":   {Code: KeyPgUp, Alt: false, Shift: false, Ctrl: false},
	"\x1b[5;3~": {Code: KeyPgUp, Alt: false, Shift: false, Ctrl: false},
	"\x1b[5;5~": {Code: KeyPgUp, Alt: false, Shift: false, Ctrl: false},
	"\x1b[5^":   {Code: KeyPgUp, Alt: false, Shift: false, Ctrl: false},
	"\x1b[5;7~": {Code: KeyPgUp, Alt: false, Shift: false, Ctrl: false},

	"\x1b[6~":   {Code: KeyPgDown, Alt: false, Shift: false, Ctrl: false},
	"\x1b[6;3~": {Code: KeyPgDown, Alt: false, Shift: false, Ctrl: false},
	"\x1b[6;5~": {Code: KeyPgDown, Alt: false, Shift: false, Ctrl: false},
	"\x1b[6^":   {Code: KeyPgDown, Alt: false, Shift: false, Ctrl: false},
	"\x1b[6;7~": {Code: KeyPgDown, Alt: false, Shift: false, Ctrl: false},

	"\x1b[1~":   {Code: KeyHome, Alt: false, Shift: false, Ctrl: false},
	"\x1b[H":    {Code: KeyHome, Alt: false, Shift: false, Ctrl: false},
	"\x1b[1;3H": {Code: KeyHome, Alt: true, Shift: false, Ctrl: false},
	"\x1b[1;5H": {Code: KeyHome, Alt: false, Shift: false, Ctrl: true},
	"\x1b[1;7H": {Code: KeyHome, Alt: true, Shift: false, Ctrl: true},
	"\x1b[1;2H": {Code: KeyHome, Alt: false, Shift: true, Ctrl: false},
	"\x1b[1;4H": {Code: KeyHome, Alt: true, Shift: true, Ctrl: false},
	"\x1b[1;6H": {Code: KeyHome, Alt: false, Shift: true, Ctrl: true},
	"\x1b[1;8H": {Code: KeyHome, Alt: true, Shift: true, Ctrl: true},

	"\x1b[4~":   {Code: KeyEnd, Alt: false, Shift: false, Ctrl: false},
	"\x1b[F":    {Code: KeyEnd, Alt: false, Shift: false, Ctrl: false},
	"\x1b[1;3F": {Code: KeyEnd, Alt: true, Shift: false, Ctrl: false},
	"\x1b[1;5F": {Code: KeyEnd, Alt: false, Shift: false, Ctrl: true},
	"\x1b[1;7F": {Code: KeyEnd, Alt: true, Shift: false, Ctrl: true},
	"\x1b[1;2F": {Code: KeyEnd, Alt: false, Shift: true, Ctrl: false},
	"\x1b[1;4F": {Code: KeyEnd, Alt: true, Shift: true, Ctrl: false},
	"\x1b[1;6F": {Code: KeyEnd, Alt: false, Shift: true, Ctrl: true},
	"\x1b[1;8F": {Code: KeyEnd, Alt: true, Shift: true, Ctrl: true},

	"\x1b[7~": {Code: KeyHome, Alt: false, Shift: false, Ctrl: false},
	"\x1b[7^": {Code: KeyHome, Alt: false, Shift: false, Ctrl: true},
	"\x1b[7$": {Code: KeyHome, Alt: false, Shift: true, Ctrl: false},
	"\x1b[7@": {Code: KeyHome, Alt: false, Shift: true, Ctrl: true},

	"\x1b[8~": {Code: KeyEnd, Alt: false, Shift: false, Ctrl: false},
	"\x1b[8^": {Code: KeyEnd, Alt: false, Shift: false, Ctrl: true},
	"\x1b[8$": {Code: KeyEnd, Alt: false, Shift: true, Ctrl: false},
	"\x1b[8@": {Code: KeyEnd, Alt: false, Shift: true, Ctrl: true},

	// Function keys, Linux console
	"\x1b[[A": {Code: KeyF1, Alt: false, Shift: false, Ctrl: false},
	"\x1b[[B": {Code: KeyF2, Alt: false, Shift: false, Ctrl: false},
	"\x1b[[C": {Code: KeyF3, Alt: false, Shift: false, Ctrl: false},
	"\x1b[[D": {Code: KeyF4, Alt: false, Shift: false, Ctrl: false},
	"\x1b[[E": {Code: KeyF5, Alt: false, Shift: false, Ctrl: false},

	// Function keys, X11 / vt100 / xterm
	"\x1bOP": {Code: KeyF1, Alt: false, Shift: false, Ctrl: false},
	"\x1bOQ": {Code: KeyF2, Alt: false, Shift: false, Ctrl: false},
	"\x1bOR": {Code: KeyF3, Alt: false, Shift: false, Ctrl: false},
	"\x1bOS": {Code: KeyF4, Alt: false, Shift: false, Ctrl: false},

	"\x1b[1;3P": {Code: KeyF1, Alt: true, Shift: false, Ctrl: false},
	"\x1b[1;3Q": {Code: KeyF2, Alt: true, Shift: false, Ctrl: false},
	"\x1b[1;3R": {Code: KeyF3, Alt: true, Shift: false, Ctrl: false},
	"\x1b[1;3S": {Code: KeyF4, Alt: true, Shift: false, Ctrl: false},

	"\x1b[11~": {Code: KeyF1, Alt: false, Shift: false, Ctrl: false},
	"\x1b[12~": {Code: KeyF2, Alt: false, Shift: false, Ctrl: false},
	"\x1b[13~": {Code: KeyF3, Alt: false, Shift: false, Ctrl: false},
	"\x1b[14~": {Code: KeyF4, Alt: false, Shift: false, Ctrl: false},

	"\x1b[15~":   {Code: KeyF5, Alt: false, Shift: false, Ctrl: false},
	"\x1b[15;3~": {Code: KeyF5, Alt: true, Shift: false, Ctrl: false},

	"\x1b[17~": {Code: KeyF6, Alt: false, Shift: false, Ctrl: false},
	"\x1b[18~": {Code: KeyF7, Alt: false, Shift: false, Ctrl: false},
	"\x1b[19~": {Code: KeyF8, Alt: false, Shift: false, Ctrl: false},
	"\x1b[20~": {Code: KeyF9, Alt: false, Shift: false, Ctrl: false},
	"\x1b[21~": {Code: KeyF10, Alt: false, Shift: false, Ctrl: false},

	"\x1b[17;3~": {Code: KeyF6, Alt: true, Shift: false, Ctrl: false},
	"\x1b[18;3~": {Code: KeyF7, Alt: true, Shift: false, Ctrl: false},
	"\x1b[19;3~": {Code: KeyF8, Alt: true, Shift: false, Ctrl: false},
	"\x1b[20;3~": {Code: KeyF9, Alt: true, Shift: false, Ctrl: false},
	"\x1b[21;3~": {Code: KeyF10, Alt: true, Shift: false, Ctrl: false},

	"\x1b[23~": {Code: KeyF11, Alt: false, Shift: false, Ctrl: false},
	"\x1b[24~": {Code: KeyF12, Alt: false, Shift: false, Ctrl: false},

	"\x1b[23;3~": {Code: KeyF11, Alt: true, Shift: false, Ctrl: false},
	"\x1b[24;3~": {Code: KeyF12, Alt: true, Shift: false, Ctrl: false},

	"\x1b[1;2P":  {Code: KeyF13, Alt: false, Shift: false, Ctrl: false},
	"\x1b[1;2Q":  {Code: KeyF14, Alt: false, Shift: false, Ctrl: false},
	"\x1b[25~":   {Code: KeyF13, Alt: false, Shift: false, Ctrl: false},
	"\x1b[26~":   {Code: KeyF14, Alt: false, Shift: false, Ctrl: false},
	"\x1b[25;3~": {Code: KeyF13, Alt: true, Shift: false, Ctrl: false},
	"\x1b[26;3~": {Code: KeyF14, Alt: true, Shift: false, Ctrl: false},

	"\x1b[1;2R":  {Code: KeyF15, Alt: false, Shift: false, Ctrl: false},
	"\x1b[1;2S":  {Code: KeyF16, Alt: false, Shift: false, Ctrl: false},
	"\x1b[28~":   {Code: KeyF15, Alt: false, Shift: false, Ctrl: false},
	"\x1b[29~":   {Code: KeyF16, Alt: false, Shift: false, Ctrl: false},
	"\x1b[28;3~": {Code: KeyF15, Alt: true, Shift: false, Ctrl: false},
	"\x1b[29;3~": {Code: KeyF16, Alt: true, Shift: false, Ctrl: false},

	"\x1b[15;2~": {Code: KeyF17, Alt: false, Shift: false, Ctrl: false},
	"\x1b[17;2~": {Code: KeyF18, Alt: false, Shift: false, Ctrl: false},
	"\x1b[18;2~": {Code: KeyF19, Alt: false, Shift: false, Ctrl: false},
	"\x1b[19;2~": {Code: KeyF20, Alt: false, Shift: false, Ctrl: false},
	"\x1b[31~":   {Code: KeyF17, Alt: false, Shift: false, Ctrl: false},
	"\x1b[32~":   {Code: KeyF18, Alt: false, Shift: false, Ctrl: false},
	"\x1b[33~":   {Code: KeyF19, Alt: false, Shift: false, Ctrl: false},
	"\x1b[34~":   {Code: KeyF20, Alt: false, Shift: false, Ctrl: false},

	// Powershell sequences.
	"\x1bOA": {Code: KeyUp, Alt: false, Shift: false, Ctrl: false},
	"\x1bOB": {Code: KeyDown, Alt: false, Shift: false, Ctrl: false},
	"\x1bOC": {Code: KeyRight, Alt: false, Shift: false, Ctrl: false},
	"\x1bOD": {Code: KeyLeft, Alt: false, Shift: false, Ctrl: false},
}

func (k KeyPress) String() string {
	var keys []string
	if k.Ctrl {
		keys = append(keys, "CTRL")
	}
	if k.Alt {
		keys = append(keys, "ALT")
	}
	if k.Shift {
		keys = append(keys, "SHIFT")
	}
	keys = append(keys, k.Code.String())
	return strings.Join(keys, "+")
}

func GetKeyPressFromInput(s string) (KeyPress, bool) {
	k, ok := sequenceMapping[s]
	return k, ok
}

func getAllSupportedKeyPresses() map[string]KeyPress {
	unique := make(map[string]KeyPress)
	for _, k := range sequenceMapping {
		unique[strings.ToUpper(k.String())] = k
	}
	return unique
}

var allSupportedKeyPresses = getAllSupportedKeyPresses()

func ParseKeybind(s string) (KeyPress, bool) {
	k, ok := allSupportedKeyPresses[strings.ToUpper(s)]
	return k, ok
}
