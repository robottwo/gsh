package utils

import "net/http"

type llmTransport struct {
	Headers map[string]string
}

func (t *llmTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Headers {
		req.Header.Add(k, v)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func NewLLMHttpClient(headers map[string]string) *http.Client {
	return &http.Client{
		Transport: &llmTransport{Headers: headers},
	}
}
