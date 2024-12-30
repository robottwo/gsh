package rag

type ContextRetriever interface {
	GetContext(options ContextRetrievalOptions) (string, error)
}
