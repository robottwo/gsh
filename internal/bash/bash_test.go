package bash

import (
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

func TestParsingTypeset(t *testing.T) {
	parser := syntax.NewParser()

	command := "typeset -i x=1"
	reader := strings.NewReader(command)

	prog, err := parser.Parse(reader, "")
	if err != nil {
		t.Errorf("error parsing: %v", err)
	}

	if len(prog.Stmts) != 1 {
		t.Errorf("expected 1 statement, got %d", len(prog.Stmts))
	}

	decl, ok := prog.Stmts[0].Cmd.(*syntax.DeclClause)
	if !ok {
		t.Errorf("expected DeclClause, got %T", prog.Stmts[0].Cmd)
	}

	if decl.Variant.Value != "typeset" {
		t.Errorf("expected typeset, got %s", decl.Variant.Value)
	}
}
