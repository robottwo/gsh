package environment

import (
	"context"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

const (
	DEFAULT_PROMPT = "gsh> "
)

func GetHistoryContextLimit(runner *interp.Runner, logger *zap.Logger) int {
	historyContextLimit, err := strconv.ParseInt(
		runner.Vars["GSH_PAST_COMMANDS_CONTEXT_LIMIT"].String(), 10, 32)
	if err != nil {
		logger.Debug("error parsing GSH_PAST_COMMANDS_CONTEXT_LIMIT", zap.Error(err))
		historyContextLimit = 30
	}
	return int(historyContextLimit)
}

func GetLogLevel(runner *interp.Runner) zap.AtomicLevel {
	logLevel, err := zap.ParseAtomicLevel(runner.Vars["GSH_LOG_LEVEL"].String())
	if err != nil {
		logLevel = zap.NewAtomicLevel()
	}
	return logLevel
}

func ShouldCleanLogFile(runner *interp.Runner) bool {
	cleanLogFile := strings.ToLower(runner.Vars["GSH_CLEAN_LOG_FILE"].String())
	return cleanLogFile == "1" || cleanLogFile == "true"
}

func GetPwd(runner *interp.Runner) string {
	return runner.Vars["PWD"].String()
}

func GetPrompt(runner *interp.Runner, logger *zap.Logger) string {
	promptUpdater := runner.Funcs["GSH_UPDATE_PROMPT"]
	if promptUpdater != nil {
		err := runner.Run(context.Background(), promptUpdater)
		if err != nil {
			logger.Warn("error updating prompt", zap.Error(err))
		}
	}

	prompt := runner.Vars["GSH_PROMPT"].String()
	if prompt != "" {
		return prompt
	}
	return DEFAULT_PROMPT
}

func GetAgentContextWindowTokens(runner *interp.Runner, logger *zap.Logger) int {
	agentContextWindow, err := strconv.ParseInt(
		runner.Vars["GSH_AGENT_CONTEXT_WINDOW_TOKENS"].String(), 10, 32)
	if err != nil {
		logger.Debug("error parsing GSH_AGENT_CONTEXT_WINDOW_TOKENS", zap.Error(err))
		agentContextWindow = 32768
	}
	return int(agentContextWindow)
}

func GetMinimumLines(runner *interp.Runner, logger *zap.Logger) int {
	minimumLines, err := strconv.ParseInt(
		runner.Vars["GSH_MINIMUM_HEIGHT"].String(), 10, 32)
	if err != nil {
		logger.Debug("error parsing GSH_MINIMUM_HEIGHT", zap.Error(err))
		minimumLines = 8
	}
	return int(minimumLines)
}

func getContextTypes(runner *interp.Runner, key string) []string {
	contextTypes := strings.ToLower(runner.Vars[key].String())
	return lo.Map(strings.Split(contextTypes, ","), func(s string, _ int) string {
		return strings.TrimSpace(s)
	})
}

func GetContextTypesForAgent(runner *interp.Runner, logger *zap.Logger) []string {
	return getContextTypes(runner, "GSH_CONTEXT_TYPES_FOR_AGENT")
}

func GetContextTypesForPredictionWithPrefix(runner *interp.Runner, logger *zap.Logger) []string {
	return getContextTypes(runner, "GSH_CONTEXT_TYPES_FOR_PREDICTION_WITH_PREFIX")
}

func GetContextTypesForPredictionWithoutPrefix(runner *interp.Runner, logger *zap.Logger) []string {
	return getContextTypes(runner, "GSH_CONTEXT_TYPES_FOR_PREDICTION_WITHOUT_PREFIX")
}

func GetContextTypesForExplanation(runner *interp.Runner, logger *zap.Logger) []string {
	return getContextTypes(runner, "GSH_CONTEXT_TYPES_FOR_EXPLANATION")
}

func GetContextNumHistoryConcise(runner *interp.Runner, logger *zap.Logger) int {
	numHistoryConcise, err := strconv.ParseInt(
		runner.Vars["GSH_CONTEXT_NUM_HISTORY_CONCISE"].String(), 10, 32)
	if err != nil {
		logger.Debug("error parsing GSH_CONTEXT_NUM_HISTORY_CONCISE", zap.Error(err))
		numHistoryConcise = 30
	}
	return int(numHistoryConcise)
}

func GetContextNumHistoryVerbose(runner *interp.Runner, logger *zap.Logger) int {
	numHistoryVerbose, err := strconv.ParseInt(
		runner.Vars["GSH_CONTEXT_NUM_HISTORY_VERBOSE"].String(), 10, 32)
	if err != nil {
		logger.Debug("error parsing GSH_CONTEXT_NUM_HISTORY_VERBOSE", zap.Error(err))
		numHistoryVerbose = 30
	}
	return int(numHistoryVerbose)
}
