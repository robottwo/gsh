package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewLLMHttpClient(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer test-token",
		"Content-Type":  "application/json",
	}

	client := NewLLMHttpClient(headers)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		for k, v := range headers {
			if r.Header.Get(k) != v {
				t.Errorf("expected header %s to be %s, got %s", k, v, r.Header.Get(k))
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Make a request
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}