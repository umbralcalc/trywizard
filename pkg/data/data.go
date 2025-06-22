package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Summary struct {
	Drives []struct {
		Plays []struct {
			Text   string `json:"text"`
			Clock  string `json:"clock"`
			Period int    `json:"period"`
		} `json:"plays"`
	} `json:"drives"`
}

func fetchCommentary(gameID int) ([]string, error) {
	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/rugby/premiership/summary?event=%d", gameID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var sum Summary
	if err := json.Unmarshal(body, &sum); err != nil {
		return nil, err
	}

	var comments []string
	for _, d := range sum.Drives {
		for _, p := range d.Plays {
			comments = append(comments, fmt.Sprintf("%s — Q%d %s", p.Text, p.Period, p.Clock))
		}
	}
	return comments, nil
}
