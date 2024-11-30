package main

import (
	"github.com/atinylittleshell/gsh/internal/core"
)

func main() {
	shell, err := core.NewShell()
	if err != nil {
		panic(err)
	}

	shell.Run()
}
