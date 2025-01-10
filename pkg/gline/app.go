package gline

import (
	"fmt"
	"strings"
	"time"

	"github.com/atinylittleshell/gsh/pkg/shellinput"
	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
)

type appModel struct {
	predictor Predictor
	explainer Explainer
	logger    *zap.Logger
	options   Options

	textInput         shellinput.Model
	dirty             bool
	prediction        string
	explanation       string
	predictionStateId int

	historyValues []string
	result        string
	appState      appState

	explanationStyle lipgloss.Style
}

type attemptPredictionMsg struct {
	stateId int
}

type setPredictionMsg struct {
	stateId    int
	prediction string
}

type attemptExplanationMsg struct {
	stateId    int
	prediction string
}

type setExplanationMsg struct {
	stateId     int
	explanation string
}

type terminateMsg struct{}

func terminate() tea.Msg {
	return terminateMsg{}
}

type appState int

const (
	Active appState = iota
	Terminated
)

func initialModel(
	prompt string,
	historyValues []string,
	explanation string,
	predictor Predictor,
	explainer Explainer,
	logger *zap.Logger,
	options Options,
) appModel {
	textInput := shellinput.New()
	textInput.Prompt = prompt
	textInput.SetHistoryValues(historyValues)
	textInput.Cursor.SetMode(cursor.CursorStatic)
	textInput.ShowSuggestions = true
	textInput.Focus()

	return appModel{
		predictor: predictor,
		explainer: explainer,
		logger:    logger,
		options:   options,

		textInput:     textInput,
		dirty:         false,
		prediction:    "",
		explanation:   explanation,
		historyValues: historyValues,
		result:        "",
		appState:      Active,

		predictionStateId: 0,

		explanationStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")),
	}
}

func (m appModel) Init() tea.Cmd {
	return func() tea.Msg {
		return attemptPredictionMsg{
			stateId: m.predictionStateId,
		}
	}
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.textInput.Width = msg.Width
		m.explanationStyle = m.explanationStyle.Width(max(1, msg.Width-2))
		return m, nil

	case terminateMsg:
		m.appState = Terminated
		return m, nil

	case attemptPredictionMsg:
		return m.attemptPrediction(msg)

	case setPredictionMsg:
		return m.setPrediction(msg.stateId, msg.prediction)

	case attemptExplanationMsg:
		return m.attemptExplanation(msg)

	case setExplanationMsg:
		return m.setExplanation(msg)

	case tea.KeyMsg:
		switch msg.String() {

		// TODO: replace with custom keybindings
		case "backspace":
			// if the input is already empty, we should clear prediction
			if m.textInput.Value() == "" {
				m.dirty = true
				m.predictionStateId++
				m.clearPrediction()
				return m, nil
			}

		case "enter":
			m.result = m.textInput.Value()
			return m, tea.Sequence(terminate, tea.Quit)

		case "ctrl+c":
			return m, tea.Sequence(terminate, tea.Quit)
		}
	}

	return m.updateTextInput(msg)
}

func (m appModel) View() string {
	// Once terminated, render nothing
	if m.appState == Terminated {
		return ""
	}

	// Render normal state
	s := m.textInput.View()
	if m.explanation != "" {
		s += "\n"
		s += m.explanationStyle.Render(m.explanation)
	}

	numLines := strings.Count(s, "\n")
	if numLines < m.options.MinHeight {
		s += strings.Repeat("\n", m.options.MinHeight-numLines)
	}

	return s
}

func (m appModel) getFinalOutput() string {
	m.textInput.SetValue(m.result)
	m.textInput.SetSuggestions([]string{})
	m.textInput.Blur()
	m.textInput.ShowSuggestions = false

	s := m.textInput.View()
	return s
}

func (m appModel) updateTextInput(msg tea.Msg) (appModel, tea.Cmd) {
	oldVal := m.textInput.Value()
	updatedTextInput, cmd := m.textInput.Update(msg)
	newVal := updatedTextInput.Value()

	textUpdated := oldVal != newVal
	m.textInput = updatedTextInput

	// if the text input has changed, we want to attempt a prediction
	if textUpdated && m.predictor != nil {
		m.predictionStateId++

		userInput := updatedTextInput.Value()

		// whenever the user has typed something, mark the model as dirty
		if len(userInput) > 0 {
			m.dirty = true
		}

		if len(userInput) == 0 && m.dirty {
			// if the model was dirty earlier, but now the user has cleared the input,
			// we should clear the prediction
			m.clearPrediction()
		} else if len(userInput) > 0 && strings.HasPrefix(m.prediction, userInput) {
			// if the prediction already starts with the user input, we don't need to predict again
			m.logger.Debug("gline existing predicted input already starts with user input", zap.String("userInput", userInput))
		} else {
			// in other cases, we should kick off a debounced prediction after clearing the current one
			m.clearPrediction()

			cmd = tea.Batch(cmd, tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
				return attemptPredictionMsg{
					stateId: m.predictionStateId,
				}
			}))
		}
	}

	return m, cmd
}

func (m *appModel) clearPrediction() {
	m.prediction = ""
	m.explanation = ""
	m.textInput.SetSuggestions([]string{})
}

func (m appModel) setPrediction(stateId int, prediction string) (appModel, tea.Cmd) {
	if stateId != m.predictionStateId {
		m.logger.Debug(
			"gline discarding prediction",
			zap.Int("startStateId", stateId),
			zap.Int("newStateId", m.predictionStateId),
		)
		return m, nil
	}

	m.prediction = prediction
	m.textInput.SetSuggestions([]string{prediction})
	m.explanation = ""
	return m, tea.Cmd(func() tea.Msg {
		return attemptExplanationMsg{stateId: m.predictionStateId, prediction: prediction}
	})
}

func (m appModel) attemptPrediction(msg attemptPredictionMsg) (tea.Model, tea.Cmd) {
	if m.predictor == nil {
		return m, nil
	}
	if msg.stateId != m.predictionStateId {
		return m, nil
	}

	return m, tea.Cmd(func() tea.Msg {
		prediction, err := m.predictor.Predict(m.textInput.Value())
		if err != nil {
			m.logger.Error("gline prediction failed", zap.Error(err))
			return nil
		}

		m.logger.Debug(
			"gline predicted input",
			zap.Int("stateId", msg.stateId),
			zap.String("prediction", prediction),
		)
		return setPredictionMsg{stateId: msg.stateId, prediction: prediction}
	})
}

func (m appModel) attemptExplanation(msg attemptExplanationMsg) (tea.Model, tea.Cmd) {
	if m.explainer == nil {
		return m, nil
	}
	if msg.stateId != m.predictionStateId {
		return m, nil
	}

	return m, tea.Cmd(func() tea.Msg {
		explanation, err := m.explainer.Explain(msg.prediction)
		if err != nil {
			m.logger.Error("gline explanation failed", zap.Error(err))
			return nil
		}

		m.logger.Debug(
			"gline explained prediction",
			zap.Int("stateId", msg.stateId),
			zap.String("explanation", explanation),
		)
		return setExplanationMsg{stateId: msg.stateId, explanation: explanation}
	})
}

func (m appModel) setExplanation(msg setExplanationMsg) (tea.Model, tea.Cmd) {
	if msg.stateId != m.predictionStateId {
		m.logger.Debug(
			"gline discarding explanation",
			zap.Int("startStateId", msg.stateId),
			zap.Int("newStateId", m.predictionStateId),
		)
		return m, nil
	}

	m.explanation = msg.explanation
	return m, nil
}

func Gline(
	prompt string,
	historyValues []string,
	explanation string,
	predictor Predictor,
	explainer Explainer,
	logger *zap.Logger,
	options Options,
) (string, error) {
	p := tea.NewProgram(
		initialModel(prompt, historyValues, explanation, predictor, explainer, logger, options),
	)

	m, err := p.Run()
	if err != nil {
		return "", err
	}

	appModel, ok := m.(appModel)
	if !ok {
		logger.Error("Gline resulted in an unexpected app model")
		panic("Gline resulted in an unexpected app model")
	}

	fmt.Print(RESET_CURSOR_COLUMN + appModel.getFinalOutput() + "\n")

	return appModel.result, nil
}
