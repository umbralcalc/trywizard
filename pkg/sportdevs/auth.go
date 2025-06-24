package sportdevs

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func NewClient() (*http.Client, error) {
	apiKey := os.Getenv("SPORTDEVS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing SPORTDEVS_API_KEY")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	client.Transport = roundTripperWithAuth{apiKey, http.DefaultTransport}
	return client, nil
}

type roundTripperWithAuth struct {
	apiKey string
	rt     http.RoundTripper
}

func (r roundTripperWithAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	return r.rt.RoundTrip(req)
}
