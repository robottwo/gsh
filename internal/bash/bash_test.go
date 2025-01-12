package bash

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

func TestParsingTypeset(t *testing.T) {
	parser := syntax.NewParser()

	command := "typeset -i x=1"
	reader := strings.NewReader(command)

	prog, err := parser.Parse(reader, "")
	assert.NoError(t, err, "error parsing")
	assert.Len(t, prog.Stmts, 1, "expected 1 statement")

	decl, ok := prog.Stmts[0].Cmd.(*syntax.DeclClause)
	assert.True(t, ok, "expected DeclClause, got %T", prog.Stmts[0].Cmd)
	assert.Equal(t, "typeset", decl.Variant.Value, "expected typeset")
}
