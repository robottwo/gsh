package retrievers

import (
	"fmt"
	"strings"

	"github.com/atinylittleshell/gsh/internal/bash"
	"github.com/atinylittleshell/gsh/internal/rag"
	"mvdan.cc/sh/v3/interp"
)

type GitContextRetriever struct {
	Runner *interp.Runner
}

func (r GitContextRetriever) GetContext(options rag.ContextRetrievalOptions) (string, error) {
	revParseOut, _, err := bash.RunBashCommandInSubShell(r.Runner, "git rev-parse --show-toplevel")
	if err != nil {
		return "", err
	}
	statusOut, _, err := bash.RunBashCommandInSubShell(r.Runner, "git status")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("<git_status>Project root: %s\n%s</git_status>", strings.TrimSpace(revParseOut), statusOut), nil
}
