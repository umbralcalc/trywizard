package sportdevs

type MatchesIncidentsResponse struct {
	MatchID   int        `json:"match_id"`
	Incidents []Incident `json:"incidents"`
}

type Incident struct {
	Time               int     `json:"time"`
	Type               string  `json:"type"`
	Class              string  `json:"class,omitempty"`
	IsHome             *bool   `json:"is_home,omitempty"`
	AwayScore          int     `json:"away_score"`
	HomeScore          int     `json:"home_score"`
	ReversedPeriodTime *int    `json:"reversed_period_time,omitempty"`
	Text               *string `json:"text,omitempty"`
	IsLive             *bool   `json:"is_live,omitempty"`
	AddedTime          *int    `json:"added_time,omitempty"`
}

type Match struct {
	ID           int    `json:"id"`
	HomeTeamName string `json:"home_team_name"`
	AwayTeamName string `json:"away_team_name"`
	StartTime    string `json:"start_time"`
}
