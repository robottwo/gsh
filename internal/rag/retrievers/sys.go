package retrievers

import (
	"fmt"
	"runtime"

	"github.com/atinylittleshell/gsh/internal/rag"
	"mvdan.cc/sh/v3/interp"
)

type SystemInfoContextRetriever struct {
	Runner *interp.Runner
}

func (r SystemInfoContextRetriever) Name() string {
	return "system_info"
}

func (r SystemInfoContextRetriever) GetContext(options rag.ContextRetrievalOptions) (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	return fmt.Sprintf("<system_info>OS: %s, Arch: %s</system_info>", osName, arch), nil
}
