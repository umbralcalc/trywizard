package match

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/umbralcalc/stochadex/pkg/simulator"
)

const (
	IdxHomeTry     = 0
	IdxAwayTry     = 1
	IdxHomePenalty = 2
	IdxAwayPenalty = 3
	IdxHomeConv    = 4
	IdxAwayConv    = 5
	IdxHomeYellow  = 6
	IdxAwayYellow  = 7
	EventWidth     = 8
)

// Position group constants for substitution covariates.
const (
	SubCovWidth       = 8 // 4 groups × 2 teams
	NumPositionGroups = 4
	GrpFrontRow       = 0
	GrpBackRow        = 1
	GrpHalves         = 2
	GrpOutsideBacks   = 3
)

// positionToGroup maps lowercase position name to group index.
var positionToGroup = map[string]int{
	"prop":       GrpFrontRow,
	"hooker":     GrpFrontRow,
	"lock":       GrpBackRow,
	"flanker":    GrpBackRow,
	"no. 8":      GrpBackRow,
	"scrum-half": GrpHalves,
	"fly-half":   GrpHalves,
	"centre":     GrpOutsideBacks,
	"wing":       GrpOutsideBacks,
	"fullback":   GrpOutsideBacks,
}

// eventKey maps (event_type, is_home) to state vector index.
var eventKey = map[string]map[bool]int{
	"try":          {true: IdxHomeTry, false: IdxAwayTry},
	"penalty goal": {true: IdxHomePenalty, false: IdxAwayPenalty},
	"conversion":   {true: IdxHomeConv, false: IdxAwayConv},
	"yellow card":  {true: IdxHomeYellow, false: IdxAwayYellow},
}

// ComputeConversionProbabilities computes the per-team probability of
// converting a try, estimated as the ratio of conversions to tries.
// Returns [homeProb, awayProb]. Defaults to 0.5 if a team has no tries.
func ComputeConversionProbabilities(storage *simulator.StateTimeStorage) []float64 {
	events := storage.GetValues("events")
	var homeTries, awayTries, homeConv, awayConv float64
	for _, ev := range events {
		homeTries += ev[IdxHomeTry]
		awayTries += ev[IdxAwayTry]
		homeConv += ev[IdxHomeConv]
		awayConv += ev[IdxAwayConv]
	}
	homeProb, awayProb := 0.5, 0.5
	if homeTries > 0 {
		homeProb = homeConv / homeTries
	}
	if awayTries > 0 {
		awayProb = awayConv / awayTries
	}
	return []float64{homeProb, awayProb}
}

// TransformEventsToStateTimeStorage reads events.csv and produces a
// StateTimeStorage with one partition ("events") containing per-minute
// event counts. The state vector per minute has width EventWidth:
// [home_try, away_try, home_penalty, away_penalty, home_conv, away_conv, home_yellow, away_yellow].
//
// The gameID filters to a single match. The homeTeamID identifies which
// team_id is the home team, used to split events into home/away.
func TransformEventsToStateTimeStorage(
	filePath string,
	gameID int,
	homeTeamID int,
) (*simulator.StateTimeStorage, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %w", filePath, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse CSV %s: %w", filePath, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV %s has no data rows", filePath)
	}

	// Parse header to find column indices.
	header := records[0]
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[h] = i
	}

	// Find the max minute to determine full range (filtered by gameID).
	maxMinute := 0
	for _, row := range records[1:] {
		gid, err := strconv.Atoi(row[colIdx["game_id"]])
		if err != nil || gid != gameID {
			continue
		}
		minute, err := parseMinute(row[colIdx["time"]])
		if err != nil {
			continue
		}
		if minute > maxMinute {
			maxMinute = minute
		}
	}

	// Build per-minute counts.
	counts := make([][]float64, maxMinute+1)
	for i := range counts {
		counts[i] = make([]float64, EventWidth)
	}

	for _, row := range records[1:] {
		gid, err := strconv.Atoi(row[colIdx["game_id"]])
		if err != nil || gid != gameID {
			continue
		}
		minute, err := parseMinute(row[colIdx["time"]])
		if err != nil {
			continue
		}
		eventType := row[colIdx["event_type"]]
		teamID, err := strconv.Atoi(row[colIdx["team_id"]])
		if err != nil {
			continue
		}
		isHome := teamID == homeTeamID
		mapping, ok := eventKey[eventType]
		if !ok {
			continue // skip event types we don't model (substitutions etc.)
		}
		idx, ok := mapping[isHome]
		if !ok {
			continue
		}
		counts[minute][idx] += 1.0
	}

	// Build StateTimeStorage.
	storage := simulator.NewStateTimeStorage()
	for minute := 0; minute <= maxMinute; minute++ {
		storage.Append("events", float64(minute), counts[minute])
	}
	return storage, nil
}

func parseMinute(timeStr string) (int, error) {
	s := strings.TrimSuffix(timeStr, "'")
	return strconv.Atoi(s)
}

// LoadPlayerPositions reads players.csv and returns a map from player_id
// to their position group index. Players with position "replacement"
// are excluded since they have no starting position group.
func LoadPlayerPositions(playersPath string, gameID int) (map[int]int, error) {
	f, err := os.Open(playersPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %w", playersPath, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse CSV %s: %w", playersPath, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV %s has no data rows", playersPath)
	}

	header := records[0]
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[h] = i
	}

	positions := make(map[int]int)
	for _, row := range records[1:] {
		gid, err := strconv.Atoi(row[colIdx["game_id"]])
		if err != nil || gid != gameID {
			continue
		}
		playerID, err := strconv.Atoi(row[colIdx["player_id"]])
		if err != nil {
			continue
		}
		pos := strings.ToLower(row[colIdx["position"]])
		grp, ok := positionToGroup[pos]
		if !ok {
			continue // skip "replacement" and unknown positions
		}
		positions[playerID] = grp
	}
	return positions, nil
}

// BuildSubstitutionCovariates constructs per-minute binary substitution
// covariate vectors from events.csv and players.csv for a single game.
// Returns [maxMinute+1][SubCovWidth] where each row is a binary vector
// indicating whether any substitution has been made in each position
// group by that minute.
// Layout: [home_front_row, home_back_row, home_halves, home_outside_backs,
//
//	away_front_row, away_back_row, away_halves, away_outside_backs]
func BuildSubstitutionCovariates(
	eventsPath string,
	playersPath string,
	gameID int,
	homeTeamID int,
	maxMinute int,
) ([][]float64, error) {
	playerPositions, err := LoadPlayerPositions(playersPath, gameID)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(eventsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %w", eventsPath, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse CSV %s: %w", eventsPath, err)
	}

	header := records[0]
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[h] = i
	}

	// Track the first minute each group was substituted.
	// -1 means no substitution yet.
	firstSubMinute := make([]int, SubCovWidth)
	for i := range firstSubMinute {
		firstSubMinute[i] = -1
	}

	for _, row := range records[1:] {
		gid, err := strconv.Atoi(row[colIdx["game_id"]])
		if err != nil || gid != gameID {
			continue
		}
		if row[colIdx["event_type"]] != "player substituted" {
			continue
		}
		minute, err := parseMinute(row[colIdx["time"]])
		if err != nil {
			continue
		}
		playerID, err := strconv.Atoi(row[colIdx["player_id"]])
		if err != nil {
			continue
		}
		grp, ok := playerPositions[playerID]
		if !ok {
			continue // player not found or is a replacement
		}
		teamID, err := strconv.Atoi(row[colIdx["team_id"]])
		if err != nil {
			continue
		}
		covIdx := grp
		if teamID != homeTeamID {
			covIdx += NumPositionGroups
		}
		if firstSubMinute[covIdx] == -1 || minute < firstSubMinute[covIdx] {
			firstSubMinute[covIdx] = minute
		}
	}

	// Build the binary covariate matrix.
	covariates := make([][]float64, maxMinute+1)
	for t := 0; t <= maxMinute; t++ {
		row := make([]float64, SubCovWidth)
		for j := 0; j < SubCovWidth; j++ {
			if firstSubMinute[j] >= 0 && t >= firstSubMinute[j] {
				row[j] = 1.0
			}
		}
		covariates[t] = row
	}
	return covariates, nil
}

// GameInfo holds the metadata for a single game needed for multi-game training.
type GameInfo struct {
	GameID     int
	HomeTeamID int
}

// ListGames reads players.csv and returns a GameInfo for every unique game,
// identifying the home team from the home_away column.
func ListGames(playersPath string) ([]GameInfo, error) {
	f, err := os.Open(playersPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %w", playersPath, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse CSV %s: %w", playersPath, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV %s has no data rows", playersPath)
	}

	header := records[0]
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[h] = i
	}

	// Collect home team ID per game.
	homeTeams := make(map[int]int) // gameID → homeTeamID
	for _, row := range records[1:] {
		gid, err := strconv.Atoi(row[colIdx["game_id"]])
		if err != nil {
			continue
		}
		if _, seen := homeTeams[gid]; seen {
			continue
		}
		if strings.ToLower(row[colIdx["home_away"]]) == "home" {
			tid, err := strconv.Atoi(row[colIdx["team_id"]])
			if err != nil {
				continue
			}
			homeTeams[gid] = tid
		}
	}

	games := make([]GameInfo, 0, len(homeTeams))
	for gid, tid := range homeTeams {
		games = append(games, GameInfo{GameID: gid, HomeTeamID: tid})
	}
	return games, nil
}

// TransformEventsWithCovariates extends TransformEventsToStateTimeStorage
// by also adding substitution covariates as a "sub_covariates" partition
// in the StateTimeStorage.
func TransformEventsWithCovariates(
	eventsPath string,
	playersPath string,
	gameID int,
	homeTeamID int,
) (*simulator.StateTimeStorage, error) {
	storage, err := TransformEventsToStateTimeStorage(eventsPath, gameID, homeTeamID)
	if err != nil {
		return nil, err
	}

	events := storage.GetValues("events")
	maxMinute := len(events) - 1

	covariates, err := BuildSubstitutionCovariates(
		eventsPath, playersPath, gameID, homeTeamID, maxMinute,
	)
	if err != nil {
		return nil, err
	}

	times := storage.GetTimes()
	for i, t := range times {
		storage.Append("sub_covariates", t, covariates[i])
	}
	return storage, nil
}
