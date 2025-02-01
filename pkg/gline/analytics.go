package gline

type PredictionAnalytics interface {
	NewEntry(input string, prediction string, actual string) error
}

type NoopPredictionAnalytics struct{}

func (p *NoopPredictionAnalytics) NewEntry(input string, prediction string, actual string) error {
	return nil
}
