package match

import (
	"math"
	"testing"
)

func TestComputeSmoothedBaselineRates(t *testing.T) {
	t.Run(
		"test that smoothed baseline rates have correct structure",
		func(t *testing.T) {
			rates, err := ComputeSmoothedBaselineRates("../../dat/events.csv", "../../dat/players.csv")
			if err != nil {
				t.Fatalf("failed to compute baseline rates: %v", err)
			}
			if len(rates) == 0 {
				t.Fatal("expected non-empty rates")
			}

			for minute, row := range rates {
				if len(row) != RateEventWidth {
					t.Fatalf("minute %d: expected width %d, got %d",
						minute, RateEventWidth, len(row))
				}
				for i, r := range row {
					if r < 0 || math.IsNaN(r) || math.IsInf(r, 0) {
						t.Errorf("minute %d rate[%d]: invalid value %f", minute, i, r)
					}
				}
			}

			// Home and away rates are fitted separately and may differ.
			anyHomeTry, anyAwayTry := false, false
			for _, row := range rates {
				if row[0] > 0 {
					anyHomeTry = true
				}
				if row[1] > 0 {
					anyAwayTry = true
				}
			}
			if !anyHomeTry {
				t.Error("all home try rates are zero")
			}
			if !anyAwayTry {
				t.Error("all away try rates are zero")
			}

			// At least some try rates should be non-zero.
			anyNonZero := false
			for _, row := range rates {
				if row[0] > 0 {
					anyNonZero = true
					break
				}
			}
			if !anyNonZero {
				t.Error("all try rates are zero")
			}
		},
	)
}

func TestRunMatchSimulationsWithBaseline(t *testing.T) {
	t.Run(
		"test that simulations with baseline produce valid results",
		func(t *testing.T) {
			baselineRates, err := ComputeSmoothedBaselineRates("../../dat/events.csv", "../../dat/players.csv")
			if err != nil {
				t.Fatalf("failed to compute baseline rates: %v", err)
			}

			// With baseline, intercepts should be ~0 (no adjustment).
			scoreCoeffs := make([]float64, ScoreRateWidth)
			cardCoeffs := make([]float64, CardRateWidth)
			convProbs := []float64{0.5, 0.5}
			nSims := 5
			nSteps := len(baselineRates) - 1

			results := RunMatchSimulations(
				scoreCoeffs, cardCoeffs, convProbs,
				baselineRates, nSims, nSteps, 2000,
			)

			if len(results) != nSims {
				t.Fatalf("expected %d results, got %d", nSims, len(results))
			}
			for i, r := range results {
				if len(r.ScoreTrajectory) == 0 {
					t.Errorf("sim %d: expected non-empty trajectory", i)
				}
				for j, pt := range r.ScoreTrajectory {
					if pt.Home < 0 || math.IsNaN(pt.Home) {
						t.Errorf("sim %d step %d: invalid home score %f",
							i, j, pt.Home)
					}
					if pt.Away < 0 || math.IsNaN(pt.Away) {
						t.Errorf("sim %d step %d: invalid away score %f",
							i, j, pt.Away)
					}
				}
			}
		},
	)
}

func TestMatchBaselineCovariateRateTraining(t *testing.T) {
	t.Run(
		"test that baseline-aware covariate training runs",
		func(t *testing.T) {
			storage, err := TransformEventsWithCovariates(
				"../../dat/events.csv", "../../dat/players.csv", 600009, 25900,
			)
			if err != nil {
				t.Fatalf("failed to load data: %v", err)
			}

			baselineRates, err := ComputeSmoothedBaselineRates("../../dat/events.csv", "../../dat/players.csv")
			if err != nil {
				t.Fatalf("failed to compute baseline rates: %v", err)
			}

			combStorage := RateEventsWithCovariatesAndBaselineStorage(
				storage, baselineRates,
			)

			initCoeffs := InitCoefficientsWithBaseline()

			windowDepth := 10
			descentIterations := 1
			outputStorage := RunMatchBaselineCovariateRateTraining(
				combStorage, initCoeffs, 0.001, descentIterations, windowDepth, false,
			)

			fittedCoeffs := ExtractFittedBaselineCovariateCoefficients(outputStorage)
			if fittedCoeffs == nil {
				t.Fatal("expected non-empty training output")
			}

			t.Logf("fitted coeffs (first 9): %v", fittedCoeffs[:9])

			// All fitted coefficients should be finite.
			for i, c := range fittedCoeffs {
				if math.IsNaN(c) || math.IsInf(c, 0) {
					t.Errorf("coeff[%d] is NaN/Inf: %f", i, c)
				}
			}

			// Split should produce correct lengths.
			score, card := SplitCoefficients(fittedCoeffs)
			if len(score) != ScoreCoeffWidth {
				t.Errorf("expected %d score coeffs, got %d",
					ScoreCoeffWidth, len(score))
			}
			if len(card) != CardCoeffWidth {
				t.Errorf("expected %d card coeffs, got %d",
					CardCoeffWidth, len(card))
			}
		},
	)
}
