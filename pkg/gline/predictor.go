package gline

type Predictor interface {
	Predict(input string, directory string) (string, error)
}

type NoopPredictor struct{}

func (p *NoopPredictor) Predict(input string, directory string) (string, error) {
	return "", nil
}
