package gline

import (
	"bufio"
	"io"
	"unicode/utf8"

	"github.com/atinylittleshell/gsh/pkg/gline/keys"
)

type state int

const (
	stateNormal state = iota
	stateEsc          // just saw '\x1b'
	stateCSI          // in a Control Sequence Introducer
	stateOSC          // sometimes used for function keys, etc.
)

type TerminalReader struct {
	reader        *bufio.Reader
	state         state
	partialEscape []byte
	partialUTF8   []byte
}

func NewTerminalReader(reader io.Reader) *TerminalReader {
	return &TerminalReader{
		reader: bufio.NewReaderSize(reader, 32),
		state:  stateNormal,
	}
}

func tryDecodeUTF8(buf *[]byte) (string, bool) {
	for {
		if len(*buf) == 0 {
			return "", false
		}
		r, size := utf8.DecodeRune(*buf)
		if r == utf8.RuneError && size == 1 {
			// Means either invalid rune or incomplete sequence;
			// wait for more bytes if incomplete.
			return "", false
		}

		// Remove the decoded bytes from the buffer.
		*buf = (*buf)[size:]

		return string(r), true
	}
}

// isTerminator checks if a byte indicates the end of a CSI/OSC sequence.
// Common terminators for CSI are letters in the range @-~.
func isTerminator(b byte) bool {
	// Very simplified check; real-world might handle a broader set
	// of terminators or sub-states (like 'm' for color codes, etc.).
	return (b >= '@' && b <= '~')
}

func (r *TerminalReader) Read() (string, keys.KeyPress, error) {
	for {
		buf := make([]byte, 1)
		n, err := r.reader.Read(buf)
		if err != nil || n == 0 {
			return "", keys.KeyPress{}, err
		}
		b := buf[0]

		switch r.state {
		case stateNormal:
			if b == 0x1b {
				// start of an escape sequence
				r.state = stateEsc
			} else if b < 0x20 || b == 0x7f {
				// control character
				k, ok := keys.GetKeyPressFromInput(string(b))
				if ok {
					return "", k, nil
				}
			} else {
				// Possibly a UTF-8 character
				r.partialUTF8 = append(r.partialUTF8, b)
				text, ok := tryDecodeUTF8(&r.partialUTF8)
				if ok {
					return text, keys.KeyPress{}, nil
				}
			}

		case stateEsc:
			// We just saw ESC. Let's see what the next byte is.
			if b == '[' {
				// ESC + '[' => CSI sequence
				r.partialEscape = []byte{'\x01', '['}
				r.state = stateCSI
			} else if b == 'O' {
				// ESC + 'O' => OSC or function key
				r.partialEscape = []byte{'\x01', 'O'}
				r.state = stateOSC
			} else {
				// ESC + something else (maybe just ESC alone, or ESC + letter)
				r.state = stateNormal
				r.partialUTF8 = []byte{b}
				text, ok := tryDecodeUTF8(&r.partialUTF8)
				if ok {
					return text, keys.KeyPress{}, nil
				}
			}

		case stateCSI:
			// We’re accumulating bytes in a CSI sequence, e.g. ESC+[1;5A
			r.partialEscape = append(r.partialEscape, b)
			if isTerminator(b) {
				k, ok := keys.GetKeyPressFromInput(string(r.partialEscape))
				r.partialEscape = nil
				r.state = stateNormal
				if ok {
					return "", k, nil
				}
			}

		case stateOSC:
			// We’re accumulating bytes for OSC or function key sequences.
			r.partialEscape = append(r.partialEscape, b)
			if isTerminator(b) {
				k, ok := keys.GetKeyPressFromInput(string(r.partialEscape))
				r.partialEscape = nil
				r.state = stateNormal
				if ok {
					return "", k, nil
				}
			}
		}
	}
}
