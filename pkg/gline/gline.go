package gline

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/atinylittleshell/gsh/pkg/debounce"
	"go.uber.org/zap"
	"golang.org/x/term"
)

const (
	EXIT_COMMAND = "exit"
)

type glineContext struct {
	predictor Predictor
	logger    *zap.Logger

	prompt         string
	promptRow      int
	directory      string
	userInput      string
	predictedInput string
	stateId        atomic.Int64

	generatePredictedInputDebounced func(input predictionInput)
}

type predictionInput struct {
	userInput string
	directory string
	stateId   int64
}

// NextLine starts a new prompt and waits for user input
func NextLine(prompt string, directory string, predictor Predictor, logger *zap.Logger, options Options) (string, error) {
	// enter raw mode and exit it when done
	origTTY, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		logger.Error("failed to turn on raw terminal mode", zap.Error(err))
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), origTTY)

	if options.ClearScreen {
		fmt.Print(CLEAR_SCREEN)
	}

	row, _, err := GetCursorPos()
	if err != nil {
		logger.Error("failed to get cursor position", zap.Error(err))
		return "", err
	}

	g := &glineContext{
		predictor:      predictor,
		logger:         logger,
		prompt:         prompt,
		promptRow:      row,
		directory:      directory,
		userInput:      "",
		predictedInput: "",
		stateId:        atomic.Int64{},
	}
	g.generatePredictedInputDebounced = debounce.DebounceWithParam(200*time.Millisecond, func(input predictionInput) {
		g.generatePredictedInput(input)
	})

	g.redrawLine()

	// Read user input
	input, err := g.readCommand()
	if err != nil {
		logger.Error("failed to read command", zap.Error(err))
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
			g.logger.Error("gline failed to read from stdin", zap.Error(err))
			return "", err
		}

		char := buffer[0]
		g.logger.Debug("gline received character", zap.String("char", string(char)))

		// increment stateId
		g.stateId.Add(1)
		g.logger.Debug("gline state id", zap.Int64("id", g.stateId.Load()))

		if char == '\n' || char == '\r' {
			// Enter
			fmt.Print(CLEAR_REMAINING_LINE)
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
		g.logger.Debug("gline existing predicted input already starts with user input")
		return
	}

	predictionInput := predictionInput{
		userInput: g.userInput,
		stateId:   g.stateId.Load(),
		directory: g.directory,
	}
	g.generatePredictedInputDebounced(predictionInput)
}

func (g *glineContext) generatePredictedInput(input predictionInput) {
	startStateId := input.stateId

	g.logger.Debug("gline predicting input", zap.Int64("stateId", startStateId), zap.String("input", input.userInput))

	go func() {
		predicted, err := g.predictor.Predict(input.userInput, input.directory)
		if err != nil {
			g.logger.Error("gline prediction failed", zap.Error(err))
		}

		newStateId := g.stateId.Load()
		g.logger.Debug("gline predicted input", zap.Int64("stateId", newStateId), zap.String("predicted", predicted))

		if startStateId != newStateId {
			// if the stateId has changed, then the user has made more input,
			// so we should discard this prediction
			g.logger.Debug(
				"gline discarding prediction",
				zap.Int64("startStateId", startStateId),
				zap.Int64("newStateId", newStateId),
			)
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
