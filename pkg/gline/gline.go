package gline

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/atinylittleshell/gsh/pkg/debounce"
	"golang.org/x/term"
)

const (
	EXIT_COMMAND = "exit"
)

type glineContext struct {
	predictor Predictor

	prompt         string
	promptRow      int
	userInput      string
	predictedInput string
	stateId        atomic.Int64

	generatePredictedInputDebounced func()
}

// NextLine starts a new prompt and waits for user input
func NextLine(prompt string, predictor Predictor, options Options) (string, error) {
	// enter raw mode and exit it when done
	origTTY, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), origTTY)

	if options.ClearScreen {
		fmt.Print(CLEAR_SCREEN)
	}

	row, _, err := GetCursorPos()
	if err != nil {
		return "", err
	}

	g := &glineContext{
		predictor:      predictor,
		prompt:         prompt,
		promptRow:      row,
		userInput:      "",
		predictedInput: "",
		stateId:        atomic.Int64{},
	}
	g.generatePredictedInputDebounced = debounce.Debounce(200*time.Millisecond, func() {
		g.generatePredictedInput()
	})

	g.redrawLine()

	// Read user input
	input, err := g.readCommand()
	if err != nil {
		return "", err
	}

	return input, nil
}

func (g *glineContext) redrawLine() error {
	fmt.Print(CLEAR_LINE)

	// Prompt
	fmt.Print(WHITE)
	fmt.Print(g.prompt)
	fmt.Print(RESET_COLOR)

	// User input
	fmt.Print(WHITE)
	fmt.Print(g.userInput)
	fmt.Print(RESET_COLOR)

	// Predicted input
	if len(g.predictedInput) > 0 && strings.HasPrefix(g.predictedInput, g.userInput) {
		fmt.Print(SAVE_CURSOR)

		fmt.Print(GRAY)
		fmt.Print(g.predictedInput[len(g.userInput):])
		fmt.Print(RESET_COLOR)

		fmt.Print(RESTORE_CURSOR)
	}

	return nil
}

func (g *glineContext) readCommand() (string, error) {
	g.userInput = ""
	g.predictedInput = ""

	buffer := make([]byte, 1)

	for {
		_, err := os.Stdin.Read(buffer)
		if err != nil {
			return "", err
		}

		char := buffer[0]
		// increment stateId
		g.stateId.Add(1)

		if char == '\n' || char == '\r' {
			// Enter
			fmt.Println()
			fmt.Print(RESET_CURSOR_COLUMN)

			result := strings.TrimSpace(g.userInput)

			g.userInput = ""
			g.predictedInput = ""

			return result, nil
		}

		if char == 127 {
			// Backspace
			if len(g.userInput) > 0 {
				g.userInput = g.userInput[:len(g.userInput)-1]
			}
		} else {
			// Normal character
			g.userInput += string(char)
		}

		g.redrawLine()
		g.predictInput()
	}
}

func (g *glineContext) predictInput() {
	if len(g.userInput) == 0 || g.predictor == nil {
		g.predictedInput = ""
		return
	}

	if strings.HasPrefix(g.predictedInput, g.userInput) {
		return
	}

	g.generatePredictedInputDebounced()
}

func (g *glineContext) generatePredictedInput() {
	startStateId := g.stateId.Load()
	userInput := g.userInput

	go func() {
		predicted, err := g.predictor.Predict(userInput)

		newStateId := g.stateId.Load()
		if startStateId != newStateId {
			// if the stateId has changed, then the user has made more input,
			// so we should discard this prediction
			return
		}

		if err != nil {
			g.predictedInput = ""
		} else {
			g.predictedInput = predicted
		}

		g.redrawLine()
	}()
}
