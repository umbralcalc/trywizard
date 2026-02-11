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
	IdxHomeTry    = 0
	IdxAwayTry    = 1
	IdxHomeConv   = 2
	IdxAwayConv   = 3
	IdxHomeYellow = 4
	IdxAwayYellow = 5
	EventWidth    = 6
)

// eventKey maps (event_type, is_home) to state vector index.
var eventKey = map[string]map[bool]int{
	"try":         {true: IdxHomeTry, false: IdxAwayTry},
	"conversion":  {true: IdxHomeConv, false: IdxAwayConv},
	"yellow card": {true: IdxHomeYellow, false: IdxAwayYellow},
}

// TransformEventsToStateTimeStorage reads events.csv and produces a
// StateTimeStorage with one partition ("events") containing per-minute
// event counts. The state vector per minute has width EventWidth:
// [home_try, away_try, home_conv, away_conv, home_yellow, away_yellow].
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
		storage.ConcurrentAppend("events", float64(minute), counts[minute])
	}
	return storage, nil
}

func parseMinute(timeStr string) (int, error) {
	s := strings.TrimSuffix(timeStr, "'")
	return strconv.Atoi(s)
}
