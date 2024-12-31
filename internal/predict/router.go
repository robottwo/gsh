package predict

type PredictRouter struct {
	PrefixPredictor    *LLMPrefixPredictor
	NullStatePredictor *LLMNullStatePredictor
}

func (p *PredictRouter) Predict(input string) (string, string, error) {
	if input == "" {
		return p.NullStatePredictor.Predict(input)
	}
	return p.PrefixPredictor.Predict(input)
}
