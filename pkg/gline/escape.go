package gline

const (
	ESC                  = "\033"
	BACKSPACE            = "\b \b"
	RESET_CURSOR         = ESC + "[H"
	RESET_CURSOR_COLUMN  = ESC + "[G"
	CLEAR_REMAINING_LINE = ESC + "[K"
	CLEAR_LINE           = RESET_CURSOR_COLUMN + ESC + "[2K"
	CLEAR_AFTER_CURSOR   = ESC + "[J"
	SAVE_CURSOR          = ESC + "[s"
	RESTORE_CURSOR       = ESC + "[u"
	GET_CURSOR_POS       = ESC + "[6n"
	MOVE_CURSOR          = ESC + "[%d;%dH"
)
