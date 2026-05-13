package rugby

import (
	"github.com/umbralcalc/stochadex/pkg/simulator"
	"github.com/umbralcalc/trywizard/pkg/match"
)

// MatchMinutes is the in-simulation length of one rugby match. The
// simulation timestep is 1.0 (one minute) so the wall-clock minute index
// inside a given match is `int(t) % MatchMinutes`.
const MatchMinutes = 80

// Match-cycling state vector layout (length MatchStateWidth). The line
// chart binds to state[0], so the in-match score diff sits there.
const (
	MSIdxScoreDiff           = 0
	MSIdxHomeScore           = 1
	MSIdxAwayScore           = 2
	MSIdxHomeActiveYellow    = 3
	MSIdxAwayActiveYellow    = 4
	MSIdxHalf                = 5
	MSIdxMatchMinute         = 6
	MSIdxMatchID             = 7
	MSIdxBaseHomeTry         = 8
	MSIdxBaseAwayTry         = 9
	MSIdxBaseHomePen         = 10
	MSIdxBaseAwayPen         = 11
	MSIdxBaseHomeConv        = 12
	MSIdxBaseAwayConv        = 13
	MSIdxBaseHomeYellow      = 14
	MSIdxBaseAwayYellow      = 15
	MSIdxLastCompletedMargin = 16
	MSIdxHomeWins            = 17
	MSIdxMatchesCompleted    = 18
	MatchStateWidth          = 19
)

// Outcomes histogram bin layout. Margins outside [HistMin, HistMax] are
// clamped into the boundary bins.
const (
	HistMin     = -50.0
	HistMax     = 50.0
	HistBins    = 20
	HistBinSize = (HistMax - HistMin) / HistBins
)

// Outcomes state layout: [home_win_pct, matches_completed, ...bin counts...].
const (
	OutIdxWinPct     = 0
	OutIdxMatches    = 1
	OutIdxBinsStart  = 2
	OutcomesWidth    = OutIdxBinsStart + HistBins
	HistogramBarsLen = HistBins * 4
)

// Canvas geometry shared by the visualization and HistogramBarsIteration.
// Keep these in sync with rugby.go's WithCanvas/AddRectangleSet calls.
const (
	HistCanvasX0     = 60
	HistCanvasY0     = 170
	HistCanvasWidth  = 520
	HistCanvasHeight = 200
)

// HomeSubSnapshotIteration is the action-state partition the sliders write
// to. It snapshots the four home-team substitution minutes at the start of
// each in-simulation match (m == 0) and holds them steady until the next
// match starts, so mid-match slider movement doesn't corrupt the current
// trajectory — it only takes effect on the following match.
//
// action_state_values is a 4-vector: [front_row, back_row, halves, outside_backs]
// minutes. State output mirrors that 4-vector but latched.
type HomeSubSnapshotIteration struct{}

func (h *HomeSubSnapshotIteration) Configure(int, *simulator.Settings) {}

func (h *HomeSubSnapshotIteration) Iterate(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) []float64 {
	t := timestepsHistory.Values.AtVec(0)
	m := int(t) % MatchMinutes

	out := make([]float64, match.NumPositionGroups)
	if m == 0 {
		actions, ok := params.GetOk("action_state_values")
		if ok {
			for i := 0; i < match.NumPositionGroups && i < len(actions); i++ {
				out[i] = actions[i]
			}
		}
		return out
	}
	prev := stateHistories[partitionIndex].CopyStateRow(0)
	copy(out, prev[:match.NumPositionGroups])
	return out
}

// SubCovariatesFromActionIteration assembles the 8-wide binary substitution
// covariate vector from the latched home minutes (upstream as
// "home_sub_minutes") and a fixed away strategy (own param "away_sub_minutes").
// The covariate flips on at m >= sub_minute for each (team, group).
type SubCovariatesFromActionIteration struct{}

func (s *SubCovariatesFromActionIteration) Configure(int, *simulator.Settings) {}

func (s *SubCovariatesFromActionIteration) Iterate(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) []float64 {
	t := timestepsHistory.Values.AtVec(0)
	m := int(t) % MatchMinutes

	home, _ := params.GetOk("home_sub_minutes")
	away, _ := params.GetOk("away_sub_minutes")

	out := make([]float64, match.SubCovWidth)
	for i := 0; i < match.NumPositionGroups; i++ {
		if i < len(home) && home[i] > 0 && float64(m) >= home[i] {
			out[i] = 1.0
		}
		if i < len(away) && away[i] > 0 && float64(m) >= away[i] {
			out[match.NumPositionGroups+i] = 1.0
		}
	}
	return out
}

// BaselineRatesByMinuteIteration replays the embedded smoothed baseline
// rates indexed by in-match minute, so each cycle of the simulation
// sees the same time-of-match baseline curve.
//
// Param "rates_by_minute" is a flattened nMinutes × RateEventWidth slice;
// param "n_minutes" is its row count.
type BaselineRatesByMinuteIteration struct{}

func (b *BaselineRatesByMinuteIteration) Configure(int, *simulator.Settings) {}

func (b *BaselineRatesByMinuteIteration) Iterate(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) []float64 {
	flat := params.Get("rates_by_minute")
	nMinutes := int(params.Get("n_minutes")[0])
	t := timestepsHistory.Values.AtVec(0)
	m := int(t) % MatchMinutes
	if m >= nMinutes {
		m = nMinutes - 1
	}

	out := make([]float64, match.RateEventWidth)
	copy(out, flat[m*match.RateEventWidth:(m+1)*match.RateEventWidth])
	return out
}

// MatchCyclingStateIteration derives the in-match state from the
// cumulative score/card/conversion partitions by subtracting baselines
// snapshotted at each match start. At m == 0 (and t > 0) it also records
// the just-completed match's margin and increments win counters.
//
// Upstream params:
//
//	score_values      score_events latest state (4 tries+penalties)
//	conversion_values conversion_events latest state (2 conversions)
//	card_values       card_events latest state (2 yellows)
//
// Own param: "card_partition" — partition index of card_events (for the
// rolling yellow-card window, same convention as match.MatchStateFunction).
type MatchCyclingStateIteration struct{}

func (m *MatchCyclingStateIteration) Configure(int, *simulator.Settings) {}

func (mc *MatchCyclingStateIteration) Iterate(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) []float64 {
	scoreValues := params.Get("score_values")
	convValues := params.Get("conversion_values")
	cardValues := params.Get("card_values")

	cumHomeTry := scoreValues[match.IdxHomeTry]
	cumAwayTry := scoreValues[match.IdxAwayTry]
	cumHomePen := scoreValues[match.IdxHomePenalty]
	cumAwayPen := scoreValues[match.IdxAwayPenalty]
	cumHomeConv := convValues[0]
	cumAwayConv := convValues[1]
	cumHomeYel := cardValues[0]
	cumAwayYel := cardValues[1]

	t := timestepsHistory.Values.AtVec(0)
	minute := int(t) % MatchMinutes
	matchID := int(t) / MatchMinutes

	prev := stateHistories[partitionIndex].CopyStateRow(0)

	out := make([]float64, MatchStateWidth)
	out[MSIdxMatchMinute] = float64(minute)
	out[MSIdxMatchID] = float64(matchID)
	if minute >= 40 {
		out[MSIdxHalf] = 1.0
	}

	switch {
	case t == 0:
		// First step of the very first match. Baselines are the current
		// cumulatives (zeros, modulo any seed noise) and there is no
		// previously-completed match to record.
		out[MSIdxBaseHomeTry] = cumHomeTry
		out[MSIdxBaseAwayTry] = cumAwayTry
		out[MSIdxBaseHomePen] = cumHomePen
		out[MSIdxBaseAwayPen] = cumAwayPen
		out[MSIdxBaseHomeConv] = cumHomeConv
		out[MSIdxBaseAwayConv] = cumAwayConv
		out[MSIdxBaseHomeYellow] = cumHomeYel
		out[MSIdxBaseAwayYellow] = cumAwayYel
	case minute == 0:
		// Start of a new match (not the first). Record the margin of the
		// match that just ended and refresh baselines.
		lastMargin := prev[MSIdxScoreDiff]
		out[MSIdxLastCompletedMargin] = lastMargin
		out[MSIdxMatchesCompleted] = prev[MSIdxMatchesCompleted] + 1
		out[MSIdxHomeWins] = prev[MSIdxHomeWins]
		if lastMargin > 0 {
			out[MSIdxHomeWins] += 1
		}
		out[MSIdxBaseHomeTry] = cumHomeTry
		out[MSIdxBaseAwayTry] = cumAwayTry
		out[MSIdxBaseHomePen] = cumHomePen
		out[MSIdxBaseAwayPen] = cumAwayPen
		out[MSIdxBaseHomeConv] = cumHomeConv
		out[MSIdxBaseAwayConv] = cumAwayConv
		out[MSIdxBaseHomeYellow] = cumHomeYel
		out[MSIdxBaseAwayYellow] = cumAwayYel
	default:
		// Mid-match: inherit baselines and counters from previous step.
		out[MSIdxLastCompletedMargin] = prev[MSIdxLastCompletedMargin]
		out[MSIdxHomeWins] = prev[MSIdxHomeWins]
		out[MSIdxMatchesCompleted] = prev[MSIdxMatchesCompleted]
		out[MSIdxBaseHomeTry] = prev[MSIdxBaseHomeTry]
		out[MSIdxBaseAwayTry] = prev[MSIdxBaseAwayTry]
		out[MSIdxBaseHomePen] = prev[MSIdxBaseHomePen]
		out[MSIdxBaseAwayPen] = prev[MSIdxBaseAwayPen]
		out[MSIdxBaseHomeConv] = prev[MSIdxBaseHomeConv]
		out[MSIdxBaseAwayConv] = prev[MSIdxBaseAwayConv]
		out[MSIdxBaseHomeYellow] = prev[MSIdxBaseHomeYellow]
		out[MSIdxBaseAwayYellow] = prev[MSIdxBaseAwayYellow]
	}

	inHomeTry := cumHomeTry - out[MSIdxBaseHomeTry]
	inAwayTry := cumAwayTry - out[MSIdxBaseAwayTry]
	inHomePen := cumHomePen - out[MSIdxBaseHomePen]
	inAwayPen := cumAwayPen - out[MSIdxBaseAwayPen]
	inHomeConv := cumHomeConv - out[MSIdxBaseHomeConv]
	inAwayConv := cumAwayConv - out[MSIdxBaseAwayConv]
	inHomeYel := cumHomeYel - out[MSIdxBaseHomeYellow]
	inAwayYel := cumAwayYel - out[MSIdxBaseAwayYellow]

	homeScore := inHomeTry*5.0 + inHomeConv*2.0 + inHomePen*3.0
	awayScore := inAwayTry*5.0 + inAwayConv*2.0 + inAwayPen*3.0

	// Active yellow cards: events in the last YellowCardMinutes minutes,
	// computed against the card_events history (same shape as the original
	// match.MatchStateFunction).
	cardPartitionIndices := params.Get("card_partition")
	cardHistory := stateHistories[int(cardPartitionIndices[0])]
	historyDepth := cardHistory.StateHistoryDepth
	lookback := match.YellowCardMinutes
	if lookback > historyDepth-1 {
		lookback = historyDepth - 1
	}

	out[MSIdxScoreDiff] = homeScore - awayScore
	out[MSIdxHomeScore] = homeScore
	out[MSIdxAwayScore] = awayScore
	out[MSIdxHomeActiveYellow] = inHomeYel - (cardHistory.Values.At(lookback, 0) - out[MSIdxBaseHomeYellow])
	out[MSIdxAwayActiveYellow] = inAwayYel - (cardHistory.Values.At(lookback, 1) - out[MSIdxBaseAwayYellow])
	if out[MSIdxHomeActiveYellow] < 0 {
		out[MSIdxHomeActiveYellow] = 0
	}
	if out[MSIdxAwayActiveYellow] < 0 {
		out[MSIdxAwayActiveYellow] = 0
	}
	return out
}

// OutcomesIteration aggregates completed-match margins into a fixed-bin
// histogram. State layout (length OutcomesWidth):
//
//	state[0]            home win percentage so far
//	state[1]            number of matches completed
//	state[2..21]        per-bin count (HistBins=20 bins)
//
// On each step it compares match_state's matches_completed counter to its
// own. A change means a fresh match just finished; bin its margin.
type OutcomesIteration struct{}

func (o *OutcomesIteration) Configure(int, *simulator.Settings) {}

func (o *OutcomesIteration) Iterate(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) []float64 {
	ms := params.Get("match_state")
	prev := stateHistories[partitionIndex].CopyStateRow(0)

	out := make([]float64, OutcomesWidth)
	copy(out, prev)

	matchesNow := ms[MSIdxMatchesCompleted]
	if matchesNow > prev[OutIdxMatches] {
		margin := ms[MSIdxLastCompletedMargin]
		bin := int((margin - HistMin) / HistBinSize)
		if bin < 0 {
			bin = 0
		}
		if bin >= HistBins {
			bin = HistBins - 1
		}
		out[OutIdxBinsStart+bin] = prev[OutIdxBinsStart+bin] + 1
		out[OutIdxMatches] = matchesNow
	}

	if out[OutIdxMatches] > 0 {
		out[OutIdxWinPct] = 100.0 * ms[MSIdxHomeWins] / out[OutIdxMatches]
	}
	return out
}

// HistogramBarsIteration projects the OutcomesIteration's bin counts onto
// canvas-space (x, y, w, h) quadruples consumed by AddRectangleSet. State
// width is HistogramBarsLen = HistBins * 4.
type HistogramBarsIteration struct{}

func (h *HistogramBarsIteration) Configure(int, *simulator.Settings) {}

func (h *HistogramBarsIteration) Iterate(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) []float64 {
	outcomes := params.Get("outcomes_values")

	out := make([]float64, HistogramBarsLen)
	maxCount := 0.0
	for i := 0; i < HistBins; i++ {
		c := outcomes[OutIdxBinsStart+i]
		if c > maxCount {
			maxCount = c
		}
	}
	if maxCount == 0 {
		return out
	}

	binStride := float64(HistCanvasWidth) / HistBins
	barW := binStride - 2
	baseline := float64(HistCanvasY0 + HistCanvasHeight)

	for i := 0; i < HistBins; i++ {
		c := outcomes[OutIdxBinsStart+i]
		if c == 0 {
			continue
		}
		h := (c / maxCount) * float64(HistCanvasHeight)
		if h < 3 {
			h = 3
		}
		out[i*4+0] = float64(HistCanvasX0) + float64(i)*binStride + 1
		out[i*4+1] = baseline - h
		out[i*4+2] = barW
		out[i*4+3] = h
	}
	return out
}
