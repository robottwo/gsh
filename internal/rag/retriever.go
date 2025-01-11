package rag

type ContextRetriever interface {
	Name() string

	GetContext() (string, error)
}
