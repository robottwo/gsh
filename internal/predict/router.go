package predict

type PredictRouter struct {
	PrefixPredictor    *LLMPrefixPredictor
	NullStatePredictor *LLMNullStatePredictor
}

func (p *PredictRouter) Predict(input string, directory string) (string, error) {
	if input == "" {
		return p.NullStatePredictor.Predict(input, directory)
	}
	return p.PrefixPredictor.Predict(input, directory)
}
