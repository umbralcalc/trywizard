package espn

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func FetchPremiershipSummary(gameID string) (*Summary, error) {
	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/rugby/267979/summary?event=%s", gameID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var sum Summary
	if err := json.Unmarshal(body, &sum); err != nil {
		return nil, err
	}
	return &sum, nil
}
