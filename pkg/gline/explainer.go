package gline

type Explainer interface {
	Explain(input string) (string, error)
}

type NoopExplainer struct{}

func (e *NoopExplainer) Explain(input string) (string, error) {
	return "", nil
}
