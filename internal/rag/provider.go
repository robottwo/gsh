package rag

import (
	"strings"

	"go.uber.org/zap"
)

type ContextProvider struct {
	Logger     *zap.Logger
	Retrievers []ContextRetriever
}

func (p *ContextProvider) GetContext() string {
	var result string
	for _, retriever := range p.Retrievers {
		output, err := retriever.GetContext()
		if err != nil {
			p.Logger.Error("error getting context", zap.Error(err))
			continue
		}

		result += output
		if !strings.HasSuffix(result, "\n") {
			result += "\n"
		}
	}
	return result
}
