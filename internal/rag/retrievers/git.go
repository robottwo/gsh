package retrievers

import (
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/internal/bash"
	"github.com/atinylittleshell/gsh/internal/rag"
	"go.uber.org/zap"
	"mvdan.cc/sh/v3/interp"
)

type GitContextRetriever struct {
	Runner *interp.Runner
	Logger *zap.Logger
}

func (r GitContextRetriever) Name() string {
	return "git"
}

func (r GitContextRetriever) GetContext(options rag.ContextRetrievalOptions) (string, error) {
	revParseOut, _, err := bash.RunBashCommandInSubShell(r.Runner, "git rev-parse --show-toplevel")
	if err != nil {
		r.Logger.Debug("error running `git rev-parse --show-toplevel`", zap.Error(err))
		return "", nil
	}
	statusOut, _, err := bash.RunBashCommandInSubShell(r.Runner, "git status")
	if err != nil {
		r.Logger.Debug("error running `git status`", zap.Error(err))
		return "", nil
	}

	return fmt.Sprintf("<git_status>Project root: %s\n%s</git_status>", strings.TrimSpace(revParseOut), statusOut), nil
}
