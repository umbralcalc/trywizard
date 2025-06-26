package sportdevs

type MatchByDateResponse struct {
	Date    string         `json:"date"`
	Matches []MatchDetails `json:"matches"`
}

type MatchDetails struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	LeagueID     int    `json:"league_id"`
	SeasonID     int    `json:"season_id"`
	StatusType   string `json:"status_type"`
	HomeTeamID   int    `json:"home_team_id"`
	AwayTeamID   int    `json:"away_team_id"`
	HomeTeamName string `json:"home_team_name"`
	AwayTeamName string `json:"away_team_name"`
	AwayScore    int    `json:"away_team_score"`
	HomeScore    int    `json:"home_team_score"`
	StartTime    string `json:"start_time"`
}

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

type MatchStatisticsResponse struct {
	MatchID    int                  `json:"match_id"`
	Statistics []MatchStatisticItem `json:"statistics"`
}

type MatchStatisticItem struct {
	Type     string `json:"type"`
	Period   string `json:"period"`
	Category string `json:"category"`
	AwayTeam string `json:"away_team"`
	HomeTeam string `json:"home_team"`
}

type MatchPlayerStatisticsResponse []MatchPlayerStatistics

type MatchPlayerStatistics struct {
	MatchID       int `json:"match_id"`
	TeamID        int `json:"team_id"`
	PlayerID      int `json:"player_id"`
	Carries       int `json:"carries"`
	CleanBreaks   int `json:"clean_breaks"`
	MetersRun     int `json:"meters_run"`
	Offloads      int `json:"offloads"`
	Passes        int `json:"passes"`
	TacklesMissed int `json:"tackles_missed"`
	Tackles       int `json:"tackles"`
	TryAssists    int `json:"try_assists"`
	TurnoversWon  int `json:"turnovers_won"`
}

type MatchLineupResponse struct {
	ID        int  `json:"id"`
	Confirmed bool `json:"confirmed"`
	HomeTeam  Team `json:"home_team"`
	AwayTeam  Team `json:"away_team"`
}

type Team struct {
	Players            []Player `json:"players"`
	PlayerColorNumber  string   `json:"player_color_number"`
	PlayerColorPrimary string   `json:"player_color_primary"`
}

type Player struct {
	PlayerID           int    `json:"player_id"`
	Substitute         bool   `json:"substitute"`
	PlayerName         string `json:"player_name"`
	ShirtNumber        int    `json:"shirt_number"`
	JerseyNumber       string `json:"jersey_number"`
	PlayerHashImage    string `json:"player_hash_image"`
	PlayerStatisticsID int    `json:"player_statistics_id"`
}
