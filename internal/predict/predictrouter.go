package predict

type PredictRouter struct {
	PrefixPredictor    *LLMPrefixPredictor
	NullStatePredictor *LLMNullStatePredictor
}

func (p *PredictRouter) UpdateContext(context *map[string]string) {
	if p.PrefixPredictor != nil {
		p.PrefixPredictor.UpdateContext(context)
	}

	if p.NullStatePredictor != nil {
		p.NullStatePredictor.UpdateContext(context)
	}
}

func (p *PredictRouter) Predict(input string) (string, error) {
	if input == "" {
		return p.NullStatePredictor.Predict(input)
	}
	return p.PrefixPredictor.Predict(input)
}
