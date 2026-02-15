package match

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/umbralcalc/stochadex/pkg/general"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

// smoothingKernelRange defines a [L, U] minute range for averaging.
type smoothingKernelRange struct{ L, U int }

// matchMinuteSmoothingKernel maps each match minute to the range of
// minutes used for computing a smoothed average. The 5-minute window
// is clamped at half boundaries (0-40 and 41-84).
var matchMinuteSmoothingKernel = map[int]smoothingKernelRange{
	0: {0, 5}, 1: {0, 5}, 2: {0, 5}, 3: {1, 6}, 4: {2, 7},
	5: {3, 8}, 6: {4, 9}, 7: {5, 10}, 8: {6, 11}, 9: {7, 12},
	10: {8, 13}, 11: {9, 14}, 12: {10, 15}, 13: {11, 16}, 14: {12, 17},
	15: {13, 18}, 16: {14, 19}, 17: {15, 20}, 18: {16, 21}, 19: {17, 22},
	20: {18, 23}, 21: {19, 24}, 22: {20, 25}, 23: {21, 26}, 24: {22, 27},
	25: {23, 28}, 26: {24, 29}, 27: {25, 30}, 28: {26, 31}, 29: {27, 32},
	30: {28, 33}, 31: {29, 34}, 32: {30, 35}, 33: {31, 36}, 34: {32, 37},
	35: {33, 38}, 36: {34, 39}, 37: {35, 40}, 38: {35, 40}, 39: {35, 40},
	40: {35, 40},
	41: {41, 46}, 42: {41, 46}, 43: {41, 46}, 44: {42, 47}, 45: {43, 48},
	46: {44, 49}, 47: {45, 50}, 48: {46, 51}, 49: {47, 52}, 50: {48, 53},
	51: {49, 54}, 52: {50, 55}, 53: {51, 56}, 54: {52, 57}, 55: {53, 58},
	56: {54, 59}, 57: {55, 60}, 58: {56, 61}, 59: {57, 62}, 60: {58, 63},
	61: {59, 64}, 62: {60, 65}, 63: {61, 66}, 64: {62, 67}, 65: {63, 68},
	66: {64, 69}, 67: {65, 70}, 68: {66, 71}, 69: {67, 72}, 70: {68, 73},
	71: {69, 74}, 72: {70, 75}, 73: {71, 76}, 74: {72, 77}, 75: {73, 78},
	76: {74, 79}, 77: {75, 80}, 78: {76, 81}, 79: {77, 82}, 80: {78, 83},
	81: {79, 84}, 82: {79, 84}, 83: {79, 84}, 84: {79, 84},
}

// rateEventTypes are the event type strings that map to rate-based events.
// Order matches rateEventIndices pairs: [home, away] for each.
var rateEventTypes = []string{"try", "penalty goal", "yellow card"}

// ComputeSmoothedBaselineRates computes per-minute baseline event rates
// from multi-game smoothed averages. For each rate event type, the raw
// per-minute counts across all games are smoothed with a 5-minute kernel,
// then divided by the number of games and split equally between home and
// away (any home/away asymmetry is captured by the model intercept).
//
// Returns [][]float64 indexed by minute, each row length RateEventWidth (6).
func ComputeSmoothedBaselineRates(eventsPath string) ([][]float64, error) {
	f, err := os.Open(eventsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %w", eventsPath, err)
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse CSV %s: %w", eventsPath, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV %s has no data rows", eventsPath)
	}

	header := records[0]
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[h] = i
	}

	// Count games and build per-minute raw event counts across all games.
	// rawCounts[minute][eventTypeIdx] = total count across all games.
	gameSet := make(map[int]struct{})
	maxMinute := 0
	type eventRecord struct {
		minute  int
		evtType string
	}
	var events []eventRecord

	for _, row := range records[1:] {
		gid, err := strconv.Atoi(row[colIdx["game_id"]])
		if err != nil {
			continue
		}
		gameSet[gid] = struct{}{}
		minute, err := strconv.Atoi(strings.TrimSuffix(row[colIdx["time"]], "'"))
		if err != nil {
			continue
		}
		evtType := row[colIdx["event_type"]]
		events = append(events, eventRecord{minute: minute, evtType: evtType})
		if minute > maxMinute {
			maxMinute = minute
		}
	}

	nGames := len(gameSet)
	if nGames == 0 {
		return nil, fmt.Errorf("no games found in %s", eventsPath)
	}

	// Build raw counts: rawCounts[minute][evtTypeIdx]
	rawCounts := make([][]float64, maxMinute+1)
	for t := range rawCounts {
		rawCounts[t] = make([]float64, len(rateEventTypes))
	}
	evtTypeIdx := make(map[string]int)
	for i, et := range rateEventTypes {
		evtTypeIdx[et] = i
	}
	for _, ev := range events {
		idx, ok := evtTypeIdx[ev.evtType]
		if !ok {
			continue
		}
		if ev.minute <= maxMinute {
			rawCounts[ev.minute][idx]++
		}
	}

	// Apply smoothing kernel and convert to per-game rates split home/away.
	rates := make([][]float64, maxMinute+1)
	for t := 0; t <= maxMinute; t++ {
		row := make([]float64, RateEventWidth)
		kernel, exists := matchMinuteSmoothingKernel[t]
		if !exists {
			kernel = smoothingKernelRange{L: t, U: t}
		}
		for evtIdx, evtType := range rateEventTypes {
			sum := 0.0
			count := 0
			for m := kernel.L; m <= kernel.U && m <= maxMinute; m++ {
				sum += rawCounts[m][evtIdx]
				count++
			}
			smoothedCount := 0.0
			if count > 0 {
				smoothedCount = sum / float64(count)
			}
			// Per-game rate, split equally between home and away.
			perGameRate := smoothedCount / float64(nGames)
			indices := smoothedEventTypeToRateIndices[evtType]
			row[indices[0]] = perGameRate / 2.0
			row[indices[1]] = perGameRate / 2.0
		}
		rates[t] = row
	}
	return rates, nil
}

// smoothedEventTypeToRateIndices maps event type strings to pairs of
// rate indices [home, away] in the RateEventWidth vector.
var smoothedEventTypeToRateIndices = map[string][2]int{
	"try":          {0, 1},
	"penalty goal": {2, 3},
	"yellow card":  {4, 5},
}

// NewBaselineRatesReplayPartition creates a partition that replays
// pre-computed baseline rates at each simulation step.
func NewBaselineRatesReplayPartition(rates [][]float64) *simulator.PartitionConfig {
	return &simulator.PartitionConfig{
		Name: "baseline_rates",
		Iteration: &general.FromStorageIteration{
			Data: rates,
		},
		Params:            simulator.NewParams(make(map[string][]float64)),
		InitStateValues:   make([]float64, RateEventWidth),
		StateHistoryDepth: 1,
		Seed:              0,
	}
}

// NewBaselineRatesConstantPartition creates a partition that returns
// zero baseline rates at every step. When the rate function sees a zero
// baseline, it falls back to the existing exp(intercept + covariates) model.
func NewBaselineRatesConstantPartition() *simulator.PartitionConfig {
	return &simulator.PartitionConfig{
		Name:      "baseline_rates",
		Iteration: &general.ParamValuesIteration{},
		Params: simulator.NewParams(map[string][]float64{
			"param_values": make([]float64, RateEventWidth),
		}),
		InitStateValues:   make([]float64, RateEventWidth),
		StateHistoryDepth: 1,
		Seed:              0,
	}
}
