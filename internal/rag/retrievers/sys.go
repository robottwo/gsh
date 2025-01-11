package retrievers

import (
	"fmt"
	"runtime"

	"mvdan.cc/sh/v3/interp"
)

type SystemInfoContextRetriever struct {
	Runner *interp.Runner
}

func (r SystemInfoContextRetriever) Name() string {
	return "system_info"
}

func (r SystemInfoContextRetriever) GetContext() (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	return fmt.Sprintf("<system_info>OS: %s, Arch: %s</system_info>", osName, arch), nil
}
