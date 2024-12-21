package gline

import (
	"fmt"
	"os"
	"strings"
)

func GetCursorPos() (int, int, error) {
	// Request cursor position
	_, err := fmt.Print(GET_CURSOR_POS)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to write request: %w", err)
	}

	var buf [32]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read response: %w", err)
	}

	response := string(buf[:n])
	// The expected format is ESC [ row ; col R
	if !strings.HasPrefix(response, ESC+"[") || !strings.HasSuffix(response, "R") {
		return 0, 0, fmt.Errorf("invalid response format: %q", response)
	}

	inner := response[2 : len(response)-1] // remove "\033[" from start and "R" from end
	var row, col int
	_, err = fmt.Sscanf(inner, "%d;%d", &row, &col)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse row/col: %w", err)
	}
	return row, col, nil
}
