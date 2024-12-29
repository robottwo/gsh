package rag

type ContextRetriever interface {
	GetContext() (string, error)
}
