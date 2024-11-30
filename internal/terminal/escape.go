package terminal

const (
	ESC                 = "\033"
	BACKSPACE           = "\b \b"
	RESET_CURSOR        = ESC + "[H"
	RESET_CURSOR_COLUMN = ESC + "[G"
	CLEAR_SCREEN        = RESET_CURSOR + ESC + "[2J"
)
