package gline

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/atinylittleshell/gsh/pkg/debounce"
	"github.com/atinylittleshell/gsh/pkg/gline/keys"
	"go.uber.org/zap"
	"golang.org/x/term"
)

const (
	EXIT_COMMAND = "exit"
)

type glineContext struct {
	predictor Predictor
	logger    *zap.Logger
	options   Options

	prompt         string
	promptRow      int
	directory      string
	cursorPosition int
	userInput      string
	predictedInput string
	stateId        atomic.Int64

	generatePredictionDebounced func(input predictionInput)
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
		predictor: predictor,
		logger:    logger,
		options:   options,

		prompt:         prompt,
		promptRow:      row,
		directory:      directory,
		cursorPosition: 0,
		userInput:      "",
		predictedInput: "",
		stateId:        atomic.Int64{},
	}
	g.generatePredictionDebounced = debounce.DebounceWithParam(200*time.Millisecond, func(input predictionInput) {
		g.generatePrediction(input)
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
	// first print the part of the user input before the cursor
	fmt.Print(g.userInput[:g.cursorPosition])
	// then save the cursor position
	fmt.Print(SAVE_CURSOR)
	// then print the part of the user input after the cursor
	fmt.Print(g.userInput[g.cursorPosition:])
	fmt.Print(RESET_COLOR)

	// Predicted input
	if len(g.predictedInput) > 0 && strings.HasPrefix(g.predictedInput, g.userInput) {
		fmt.Print(GRAY)
		fmt.Print(g.predictedInput[len(g.userInput):])
		fmt.Print(RESET_COLOR)
	}

	// Restore cursor to the saved position
	fmt.Print(RESTORE_CURSOR)

	return nil
}

func (g *glineContext) readCommand() (string, error) {
	g.userInput = ""
	g.predictedInput = ""

	g.generatePrediction(predictionInput{
		userInput: "",
		stateId:   g.stateId.Load(),
		directory: g.directory,
	})

	reader := NewTerminalReader(os.Stdin, g.logger)

	var keybindMapping = make(map[keys.KeyPress]Command)
	for command, keypresses := range g.options.Keybinds {
		for _, keypress := range keypresses {
			keybindMapping[keypress] = command
		}
	}

	for {
		text, key, err := reader.Read()
		if err != nil {
			g.logger.Error("gline failed to read from stdin", zap.Error(err))
			return "", err
		}

		g.logger.Debug("gline read", zap.String("text", text), zap.Any("key", key))

		// increment stateId
		g.stateId.Add(1)
		g.logger.Debug("gline state id", zap.Int64("id", g.stateId.Load()))

		if key.Code == keys.KeyNull {
			// Normal text input
			g.userInput = g.userInput[:g.cursorPosition] + text + g.userInput[g.cursorPosition:]
			g.cursorPosition += len(text)
		} else {
			// Keybind
			command, ok := keybindMapping[key]
			if !ok {
				g.logger.Debug("gline unknown key", zap.Any("key", key))
				continue
			}

			switch command {
			case CommandExecute:
				fmt.Print(CLEAR_REMAINING_LINE)
				fmt.Println()
				fmt.Print(RESET_CURSOR_COLUMN)

				result := strings.TrimSpace(g.userInput)

				g.userInput = ""
				g.predictedInput = ""
				g.cursorPosition = 0

				return result, nil
			case CommandBackspace:
				if g.cursorPosition > 0 {
					g.userInput = g.userInput[:g.cursorPosition-1] + g.userInput[g.cursorPosition:]
					g.cursorPosition--
				}
			case CommandHistoryPrevious:
				// TODO
			case CommandHistoryNext:
				// TODO
			case CommandCursorForward:
				if g.cursorPosition < len(g.userInput) {
					g.cursorPosition++
				} else {
					// if the cursor is at the end of the input, and there is a predicted input,
					// then move the cursor to the end of the predicted input
					if strings.HasPrefix(g.predictedInput, g.userInput) {
						g.userInput = g.predictedInput
						g.cursorPosition = len(g.userInput)
					}
				}
			case CommandCursorBackward:
				if g.cursorPosition > 0 {
					g.cursorPosition--
				}
			case CommandCursorDeleteToBeginningOfLine:
				// TODO
			case CommandCursorDeleteToEndOfLine:
				// TODO
			case CommandCursorMoveToBeginningOfLine:
				// TODO
			case CommandCursorMoveToEndOfLine:
				// TODO
			}
		}

		g.redrawLine()
		g.predictOnUserInput()
	}
}

func (g *glineContext) predictOnUserInput() {
	if len(g.userInput) > 0 && strings.HasPrefix(g.predictedInput, g.userInput) {
		g.logger.Debug("gline existing predicted input already starts with user input", zap.String("userInput", g.userInput))
		return
	}

	predictionInput := predictionInput{
		userInput: g.userInput,
		stateId:   g.stateId.Load(),
		directory: g.directory,
	}
	g.generatePredictionDebounced(predictionInput)
}

func (g *glineContext) generatePrediction(input predictionInput) {
	if g.predictor == nil {
		return
	}

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
