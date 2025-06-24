package sportdevs

import (
	"encoding/json"
	"fmt"
)

func FetchMatchList() ([]Match, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.Get("https://rugby.sportdevs.com/matches")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var arr []Match
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, fmt.Errorf("invalid match list JSON: %w", err)
	}
	return arr, nil
}

func FetchMatchIncidents(matchID int) (*MatchesIncidentsResponse, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://rugby.sportdevs.com/matches-incidents?match_id=eq.%d", matchID)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var arr []MatchesIncidentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("no incidents found")
	}
	return &arr[0], nil
}
