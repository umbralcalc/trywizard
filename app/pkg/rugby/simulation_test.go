package rugby

import (
	"math"
	"sync"
	"testing"

	"github.com/umbralcalc/stochadex/pkg/simulator"
)

// shortGen returns the dashboard's generator with termination clamped to
// `matches` matches, so individual tests stay sub-second.
func shortGen(matches int) *simulator.ConfigGenerator {
	gen := BuildRugbySimulation()
	gen.SetSimulation(&simulator.SimulationConfig{
		OutputCondition: &simulator.EveryStepOutputCondition{},
		TerminationCondition: &simulator.NumberOfStepsTerminationCondition{
			MaxNumberOfSteps: MatchMinutes * matches,
		},
		TimestepFunction: &simulator.ConstantTimestepFunction{Stepsize: 1.0},
		InitTimeValue:    0.0,
	})
	return gen
}

func TestBuildRugbySimulation(t *testing.T) {
	// Harness suite: NaN, state width, params mutation, history integrity,
	// statefulness residues — all from the stochadex side.
	t.Run("harness", func(t *testing.T) {
		settings, implementations := shortGen(2).GenerateConfigs()
		if err := simulator.RunWithHarnesses(settings, implementations); err != nil {
			t.Fatalf("harness failed: %v", err)
		}
	})

	// End-to-end: run the simulation through a few matches and check that
	// the match-cycling and histogram-aggregation invariants hold.
	t.Run("match cycling and histogram aggregation", func(t *testing.T) {
		settings, implementations := shortGen(3).GenerateConfigs()
		store := simulator.NewStateTimeStorage()
		implementations.OutputFunction = &simulator.StateTimeStorageOutputFunction{Store: store}
		coordinator := simulator.NewPartitionCoordinator(settings, implementations)
		var wg sync.WaitGroup
		for !coordinator.ReadyToTerminate() {
			coordinator.Step(&wg)
		}

		ms := store.GetValues("match_state")
		if len(ms) == 0 {
			t.Fatal("no match_state output collected")
		}
		last := ms[len(ms)-1]
		if last[MSIdxMatchesCompleted] < 2 {
			t.Errorf("expected ≥ 2 completed matches after 3 simulated matches, got %v",
				last[MSIdxMatchesCompleted])
		}

		// In-match scores must be non-negative — the baseline-subtraction
		// would silently go negative on a sign error.
		for _, row := range ms {
			if row[MSIdxHomeScore] < 0 || row[MSIdxAwayScore] < 0 {
				t.Errorf("negative in-match score at minute %v: home=%v away=%v",
					row[MSIdxMatchMinute], row[MSIdxHomeScore], row[MSIdxAwayScore])
				break
			}
		}

		// outcomes matches_completed mirrors match_state's.
		out := store.GetValues("outcomes")
		outLast := out[len(out)-1]
		if math.Abs(outLast[OutIdxMatches]-last[MSIdxMatchesCompleted]) > 0 {
			t.Errorf("outcomes matches=%v disagrees with match_state matches_completed=%v",
				outLast[OutIdxMatches], last[MSIdxMatchesCompleted])
		}
		// Win percentage must be in [0, 100].
		if outLast[OutIdxWinPct] < 0 || outLast[OutIdxWinPct] > 100 {
			t.Errorf("win pct out of range: %v", outLast[OutIdxWinPct])
		}
		// Bin counts sum to matches_completed.
		var binSum float64
		for i := 0; i < HistBins; i++ {
			binSum += outLast[OutIdxBinsStart+i]
		}
		if binSum != outLast[OutIdxMatches] {
			t.Errorf("histogram bin sum %v ≠ matches_completed %v",
				binSum, outLast[OutIdxMatches])
		}

		// histogram_bars: 4-aligned and never negative w/h (those would
		// silently disappear from the canvas).
		bars := store.GetValues("histogram_bars")
		barsLast := bars[len(bars)-1]
		if len(barsLast) != HistogramBarsLen {
			t.Errorf("histogram_bars width=%d, want %d", len(barsLast), HistogramBarsLen)
		}
		for i := 0; i+3 < len(barsLast); i += 4 {
			if barsLast[i+2] < 0 || barsLast[i+3] < 0 {
				t.Errorf("bar %d has negative size (w=%v, h=%v)",
					i/4, barsLast[i+2], barsLast[i+3])
			}
		}
	})
}
