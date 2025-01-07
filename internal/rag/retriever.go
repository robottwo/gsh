package rag

type ContextRetriever interface {
	Name() string

	GetContext(options ContextRetrievalOptions) (string, error)
}
