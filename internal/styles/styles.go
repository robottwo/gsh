package styles

import (
	"os"

	"github.com/muesli/termenv"
)

var (
	stdout = termenv.NewOutput(os.Stdout)

	ERROR = func(s string) string {
		return stdout.String(s).
			Foreground(stdout.Color("9")).
			String()
	}
	AGENT_MESSAGE = func(s string) string {
		return stdout.String(s).
			Foreground(stdout.Color("12")).
			String()
	}
	AGENT_QUESTION = func(s string) string {
		return stdout.String(s).
			Foreground(stdout.Color("11")).
			Bold().
			String()
	}
)
