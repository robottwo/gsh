package rag

import (
	"strings"

	"go.uber.org/zap"
)

type ContextProvider struct {
	Logger     *zap.Logger
	Retrievers []ContextRetriever
}

type ContextRetrievalOptions struct {
	Concise bool
}

func (p *ContextProvider) GetContext(options ContextRetrievalOptions) string {
	var result string
	for _, retriever := range p.Retrievers {
		output, err := retriever.GetContext(options)
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
