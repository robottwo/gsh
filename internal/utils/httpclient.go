package utils

import (
	"fmt"
	"net/http"
	"os"
)

type llmTransport struct {
	Headers map[string]string
}

func (t *llmTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Headers {
		req.Header.Add(k, v)
	}
	resp, err := http.DefaultTransport.RoundTrip(req)
	fmt.Fprintf(os.Stderr, "response status code: %d\n", resp.StatusCode)
	return resp, err
}

func NewLLMHttpClient(headers map[string]string) *http.Client {
	return &http.Client{
		Transport: &llmTransport{Headers: headers},
	}
}
