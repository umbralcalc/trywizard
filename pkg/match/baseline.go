package match

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/umbralcalc/stochadex/pkg/general"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

const (
	// pilotBandwidth is the half-width used in the first pass to estimate
	// local variance at each minute.
	pilotBandwidth = 5
	// minBandwidth and maxBandwidth clamp the adaptive half-width.
	minBandwidth = 5
	maxBandwidth = 14
	// bandwidthScale controls how strongly local variance widens the kernel.
	// Adaptive half-width = clamp(bandwidthScale * localStdDev, min, max).
	bandwidthScale = 5.0
)

// computeAdaptiveKernelRanges returns a per-minute [L, U] smoothing range
// that adapts to local data variability. Where the raw counts fluctuate
// more, the window is wider. The kernel is clamped at each half boundary.
func computeAdaptiveKernelRanges(rawCounts [][]float64, halfBoundary int) []smoothingKernelRange {
	n := len(rawCounts)
	nEvt := len(rawCounts[0])
	ranges := make([]smoothingKernelRange, n)

	for t := 0; t < n; t++ {
		// Determine which half this minute belongs to.
		halfL, halfU := 0, halfBoundary
		if t > halfBoundary {
			halfL, halfU = halfBoundary+1, n-1
		}

		// Pilot pass: compute local mean and variance over a small fixed window.
		pL := t - pilotBandwidth
		if pL < halfL {
			pL = halfL
		}
		pU := t + pilotBandwidth
		if pU > halfU {
			pU = halfU
		}
		maxVar := 0.0
		for evtIdx := 0; evtIdx < nEvt; evtIdx++ {
			sum, sumSq := 0.0, 0.0
			cnt := 0
			for m := pL; m <= pU; m++ {
				v := rawCounts[m][evtIdx]
				sum += v
				sumSq += v * v
				cnt++
			}
			if cnt > 1 {
				mean := sum / float64(cnt)
				variance := sumSq/float64(cnt) - mean*mean
				if variance > maxVar {
					maxVar = variance
				}
			}
		}

		// Adaptive half-width: wider where variance is higher.
		hw := int(math.Round(bandwidthScale * math.Sqrt(maxVar)))
		if hw < minBandwidth {
			hw = minBandwidth
		}
		if hw > maxBandwidth {
			hw = maxBandwidth
		}

		l := t - hw
		if l < halfL {
			l = halfL
		}
		u := t + hw
		if u > halfU {
			u = halfU
		}
		ranges[t] = smoothingKernelRange{L: l, U: u}
	}
	return ranges
}

// smoothingKernelRange defines a [L, U] minute range for averaging.
type smoothingKernelRange struct{ L, U int }

// rateEventTypes are the event type strings that map to rate-based events.
// Order matches rateEventIndices pairs: [home, away] for each.
var rateEventTypes = []string{"try", "penalty goal", "yellow card"}

// ComputeSmoothedBaselineRates computes per-minute baseline event rates
// from multi-game smoothed averages, fitting home and away rates separately.
// For each rate event type, the raw per-minute counts are split by home/away
// team, smoothed independently with an adaptive kernel, then divided by the
// number of games.
//
// Returns [][]float64 indexed by minute, each row length RateEventWidth (6).
func ComputeSmoothedBaselineRates(eventsPath, playersPath string) ([][]float64, error) {
	// Get home team IDs per game from players data.
	games, err := ListGames(playersPath)
	if err != nil {
		return nil, fmt.Errorf("listing games: %w", err)
	}
	homeTeamByGame := make(map[int]int, len(games))
	for _, g := range games {
		homeTeamByGame[g.GameID] = g.HomeTeamID
	}

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

	// Count games and build per-minute raw event counts split by home/away.
	gameSet := make(map[int]struct{})
	maxMinute := 0
	type eventRecord struct {
		minute  int
		evtType string
		isHome  bool
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
		teamID, err := strconv.Atoi(row[colIdx["team_id"]])
		if err != nil {
			continue
		}
		evtType := row[colIdx["event_type"]]
		isHome := teamID == homeTeamByGame[gid]
		events = append(events, eventRecord{minute: minute, evtType: evtType, isHome: isHome})
		if minute > maxMinute {
			maxMinute = minute
		}
	}

	nGames := len(gameSet)
	if nGames == 0 {
		return nil, fmt.Errorf("no games found in %s", eventsPath)
	}

	// Build raw counts separately for home and away.
	nEvtTypes := len(rateEventTypes)
	rawCountsHome := make([][]float64, maxMinute+1)
	rawCountsAway := make([][]float64, maxMinute+1)
	for t := range rawCountsHome {
		rawCountsHome[t] = make([]float64, nEvtTypes)
		rawCountsAway[t] = make([]float64, nEvtTypes)
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
			if ev.isHome {
				rawCountsHome[ev.minute][idx]++
			} else {
				rawCountsAway[ev.minute][idx]++
			}
		}
	}

	// Compute adaptive smoothing kernel ranges from total raw data.
	rawCountsTotal := make([][]float64, maxMinute+1)
	for t := range rawCountsTotal {
		rawCountsTotal[t] = make([]float64, nEvtTypes)
		for j := 0; j < nEvtTypes; j++ {
			rawCountsTotal[t][j] = rawCountsHome[t][j] + rawCountsAway[t][j]
		}
	}
	kernelRanges := computeAdaptiveKernelRanges(rawCountsTotal, 40)

	// Apply smoothing kernel and convert to per-game rates.
	rates := make([][]float64, maxMinute+1)
	for t := 0; t <= maxMinute; t++ {
		row := make([]float64, RateEventWidth)
		kernel := kernelRanges[t]
		for evtIdx, evtType := range rateEventTypes {
			sumHome, sumAway := 0.0, 0.0
			count := 0
			for m := kernel.L; m <= kernel.U && m <= maxMinute; m++ {
				sumHome += rawCountsHome[m][evtIdx]
				sumAway += rawCountsAway[m][evtIdx]
				count++
			}
			indices := smoothedEventTypeToRateIndices[evtType]
			if count > 0 {
				row[indices[0]] = (sumHome / float64(count)) / float64(nGames)
				row[indices[1]] = (sumAway / float64(count)) / float64(nGames)
			}
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
