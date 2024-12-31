package gline

type Predictor interface {
	Predict(input string) (string, string, error)
}

type NoopPredictor struct{}

func (p *NoopPredictor) Predict(input string) (string, string, error) {
	return "", "", nil
}
