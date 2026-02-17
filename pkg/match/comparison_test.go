package match

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/floats/scalar"
)

func TestRunMatchSimulations(t *testing.T) {
	t.Run(
		"test that simulations produce valid results",
		func(t *testing.T) {
			scoreCoeffs := []float64{-3.0, -3.0, -3.5, -3.5}
			cardCoeffs := []float64{-4.5, -4.5}
			convProbs := []float64{0.5, 0.5}
			nSims := 10
			nSteps := 80

			results := RunMatchSimulations(
				scoreCoeffs, cardCoeffs, convProbs, nil, nSims, nSteps, 1000,
			)

			if len(results) != nSims {
				t.Fatalf("expected %d results, got %d", nSims, len(results))
			}
			for i, r := range results {
				if len(r.ScoreTrajectory) == 0 {
					t.Errorf("sim %d: expected non-empty trajectory", i)
				}
				if len(r.EventTotals) != EventWidth {
					t.Errorf("sim %d: expected %d event totals, got %d",
						i, EventWidth, len(r.EventTotals))
				}
				for j, pt := range r.ScoreTrajectory {
					if pt.Home < 0 || math.IsNaN(pt.Home) {
						t.Errorf("sim %d step %d: invalid home score %f", i, j, pt.Home)
					}
					if pt.Away < 0 || math.IsNaN(pt.Away) {
						t.Errorf("sim %d step %d: invalid away score %f", i, j, pt.Away)
					}
				}
				for j, v := range r.EventTotals {
					if v < 0 || math.IsNaN(v) {
						t.Errorf("sim %d: event total[%d] invalid: %f", i, j, v)
					}
				}
				// Conversions should not exceed tries.
				if r.EventTotals[IdxHomeConv] > r.EventTotals[IdxHomeTry] {
					t.Errorf("sim %d: home conv (%f) > home tries (%f)",
						i, r.EventTotals[IdxHomeConv], r.EventTotals[IdxHomeTry])
				}
				if r.EventTotals[IdxAwayConv] > r.EventTotals[IdxAwayTry] {
					t.Errorf("sim %d: away conv (%f) > away tries (%f)",
						i, r.EventTotals[IdxAwayConv], r.EventTotals[IdxAwayTry])
				}
			}
		},
	)
}

func TestComputeScoreTrajectory(t *testing.T) {
	t.Run(
		"test that score trajectory matches known match",
		func(t *testing.T) {
			storage, err := TransformEventsToStateTimeStorage(
				"../../dat/events.csv", 600009, 25900,
			)
			if err != nil {
				t.Fatalf("failed to load data: %v", err)
			}

			trajectory := ComputeScoreTrajectory(storage)
			if len(trajectory) == 0 {
				t.Fatal("expected non-empty trajectory")
			}

			// Final score should be Home 41, Away 26.
			// (game 600009 has no penalties, so score is still tries*5 + conv*2)
			final := trajectory[len(trajectory)-1]
			if !scalar.EqualWithinAbs(final.Home, 41.0, 0.01) {
				t.Errorf("expected final home score 41, got %f", final.Home)
			}
			if !scalar.EqualWithinAbs(final.Away, 26.0, 0.01) {
				t.Errorf("expected final away score 26, got %f", final.Away)
			}

			// Scores should be monotonically non-decreasing.
			for i := 1; i < len(trajectory); i++ {
				if trajectory[i].Home < trajectory[i-1].Home {
					t.Errorf("home score decreased at minute %d: %f -> %f",
						i, trajectory[i-1].Home, trajectory[i].Home)
				}
				if trajectory[i].Away < trajectory[i-1].Away {
					t.Errorf("away score decreased at minute %d: %f -> %f",
						i, trajectory[i-1].Away, trajectory[i].Away)
				}
			}
		},
	)
}

func TestComputeEventTotals(t *testing.T) {
	t.Run(
		"test that event totals match known match",
		func(t *testing.T) {
			storage, err := TransformEventsToStateTimeStorage(
				"../../dat/events.csv", 600009, 25900,
			)
			if err != nil {
				t.Fatalf("failed to load data: %v", err)
			}

			totals := ComputeEventTotals(storage)
			if len(totals) != EventWidth {
				t.Fatalf("expected %d totals, got %d", EventWidth, len(totals))
			}

			// Home tries = 7 for this match.
			if !scalar.EqualWithinAbs(totals[IdxHomeTry], 7.0, 0.01) {
				t.Errorf("expected 7 home tries, got %f", totals[IdxHomeTry])
			}

			// No penalties in game 600009.
			if totals[IdxHomePenalty] != 0 {
				t.Errorf("expected 0 home penalties, got %f", totals[IdxHomePenalty])
			}

			// All totals should be non-negative.
			for i, v := range totals {
				if v < 0 {
					t.Errorf("total[%d] is negative: %f", i, v)
				}
			}
		},
	)
}

func TestRunStrategySimulations(t *testing.T) {
	t.Run(
		"test that strategy simulations produce valid results and win probs",
		func(t *testing.T) {
			baselineRates, err := ComputeSmoothedBaselineRates("../../dat/events.csv", "../../dat/players.csv")
			if err != nil {
				t.Fatalf("failed to compute baseline: %v", err)
			}

			// Use some plausible coefficients (small effects around zero
			// since baseline provides the rate level).
			scoreCoeffs := make([]float64, ScoreCoeffWidth)
			cardCoeffs := make([]float64, CardCoeffWidth)
			convProbs := []float64{0.7, 0.65}

			strategy := &SubstitutionStrategy{
				HomeSubs: [NumPositionGroups]int{50, 55, 60, 65},
				AwaySubs: [NumPositionGroups]int{48, 52, 0, 58},
			}

			nSims := 20
			nSteps := 80
			results := RunStrategySimulations(
				scoreCoeffs, cardCoeffs, convProbs,
				baselineRates, strategy,
				nSims, nSteps, 3000,
			)

			if len(results) != nSims {
				t.Fatalf("expected %d results, got %d", nSims, len(results))
			}

			for i, r := range results {
				if len(r.ScoreTrajectory) == 0 {
					t.Errorf("sim %d: empty trajectory", i)
				}
				if len(r.EventTotals) != EventWidth {
					t.Errorf("sim %d: expected %d event totals, got %d",
						i, EventWidth, len(r.EventTotals))
				}
			}

			probs := ComputeWinProbabilities(results)
			total := probs.HomeWin + probs.Draw + probs.AwayWin
			if !scalar.EqualWithinAbs(total, 1.0, 1e-10) {
				t.Errorf("probabilities don't sum to 1: %f", total)
			}
			t.Logf("win probs: home=%.2f draw=%.2f away=%.2f",
				probs.HomeWin, probs.Draw, probs.AwayWin)
		},
	)
}
