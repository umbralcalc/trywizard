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

func FetchMatchesByDate(date string) ([]MatchDetails, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://rugby.sportdevs.com/matches-by-date?date=eq.%s", date)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var arr []MatchByDateResponse
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(arr) == 0 {
		return nil, nil // no matches found
	}
	return arr[0].Matches, nil
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

func FetchMatchStatistics(matchID int) (*MatchStatisticsResponse, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://rugby.sportdevs.com/matches-statistics?match_id=eq.%d", matchID)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var arr []MatchStatisticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, fmt.Errorf("invalid match statistics JSON: %w", err)
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("no match statistics found")
	}
	return &arr[0], nil
}

func FetchMatchPlayersStatistics(matchID int) ([]MatchPlayerStatistics, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://rugby.sportdevs.com/matches-players-statistics?match_id=eq.%d", matchID)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var arr []MatchPlayerStatistics
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, fmt.Errorf("invalid match players statistics JSON: %w", err)
	}
	return arr, nil
}

func FetchMatchLineups(matchID int) (*MatchLineupResponse, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://rugby.sportdevs.com/matches-lineups?match_id=eq.%d", matchID)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var arr MatchLineupResponse
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return nil, fmt.Errorf("invalid match lineup JSON: %w", err)
	}
	return &arr, nil
}
