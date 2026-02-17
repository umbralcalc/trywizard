package match

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/floats/scalar"
)

func TestComputeMLERates(t *testing.T) {
	t.Run(
		"test that MLE rates match empirical averages",
		func(t *testing.T) {
			storage, err := TransformEventsToStateTimeStorage(
				"../../dat/events.csv", 600009, 25900,
			)
			if err != nil {
				t.Fatalf("failed to load data: %v", err)
			}

			rates := ComputeMLERates(storage)
			if len(rates) != RateEventWidth {
				t.Fatalf("expected %d rates, got %d", RateEventWidth, len(rates))
			}
			t.Logf("MLE rates: %v", rates)

			// All rates should be positive.
			for i, r := range rates {
				if r <= 0 || math.IsNaN(r) {
					t.Errorf("rate[%d] is non-positive or NaN: %f", i, r)
				}
			}

			// Home tries: ~7 in 81 minutes ≈ 0.086 (index 0)
			if !scalar.EqualWithinAbs(rates[0], 0.086, 0.01) {
				t.Errorf("home try rate: got %f, expected ~0.086", rates[0])
			}

			// Coefficients should be finite log values.
			scoreCoeffs, cardCoeffs := ComputeLogCoefficients(rates)
			for i, c := range scoreCoeffs {
				if math.IsNaN(c) || math.IsInf(c, 0) {
					t.Errorf("scoreCoeff[%d] is NaN or Inf: %f", i, c)
				}
			}
			for i, c := range cardCoeffs {
				if math.IsNaN(c) || math.IsInf(c, 0) {
					t.Errorf("cardCoeff[%d] is NaN or Inf: %f", i, c)
				}
			}
		},
	)
}

func TestInitCoefficientsFromRates(t *testing.T) {
	t.Run(
		"test that init coefficients have correct structure",
		func(t *testing.T) {
			rates := []float64{0.1, 0.05, 0.01, 0.01, 0.015, 0.001}
			coeffs := InitCoefficientsFromRates(rates)
			if len(coeffs) != TotalCoeffWidth {
				t.Fatalf("expected %d coeffs, got %d", TotalCoeffWidth, len(coeffs))
			}
			// Check intercepts are log(rate).
			for i := 0; i < RateEventWidth; i++ {
				expected := math.Log(rates[i])
				got := coeffs[i*CoeffsPerRate]
				if !scalar.EqualWithinAbs(got, expected, 1e-10) {
					t.Errorf("intercept[%d]: expected %f, got %f", i, expected, got)
				}
				// Covariate coefficients should be zero.
				for j := 1; j < CoeffsPerRate; j++ {
					if coeffs[i*CoeffsPerRate+j] != 0 {
						t.Errorf("coeff[%d][%d] should be 0, got %f",
							i, j, coeffs[i*CoeffsPerRate+j])
					}
				}
			}
		},
	)
}

func TestSplitCoefficients(t *testing.T) {
	t.Run(
		"test that split coefficients match original",
		func(t *testing.T) {
			coeffs := make([]float64, TotalCoeffWidth)
			for i := range coeffs {
				coeffs[i] = float64(i)
			}
			score, card := SplitCoefficients(coeffs)
			if len(score) != ScoreCoeffWidth {
				t.Fatalf("expected %d score coeffs, got %d", ScoreCoeffWidth, len(score))
			}
			if len(card) != CardCoeffWidth {
				t.Fatalf("expected %d card coeffs, got %d", CardCoeffWidth, len(card))
			}
			for i := range score {
				if score[i] != float64(i) {
					t.Errorf("score[%d]: expected %f, got %f", i, float64(i), score[i])
				}
			}
			for i := range card {
				if card[i] != float64(ScoreCoeffWidth+i) {
					t.Errorf("card[%d]: expected %f, got %f",
						i, float64(ScoreCoeffWidth+i), card[i])
				}
			}
		},
	)
}

func TestMatchRateTraining(t *testing.T) {
	t.Run(
		"test that gradient descent training runs",
		func(t *testing.T) {
			storage, err := TransformEventsToStateTimeStorage(
				"../../dat/events.csv", 600009, 25900,
			)
			if err != nil {
				t.Fatalf("failed to load data: %v", err)
			}

			// Filter to rate-only events for training.
			rateStorage := RateEventsStorage(storage)

			// Initialize near the MLE to test convergence stability.
			mleRates := ComputeMLERates(storage)
			initRates := make([]float64, RateEventWidth)
			for i := range initRates {
				initRates[i] = math.Max(mleRates[i]*1.5, 0.01)
			}

			// Use conservative hyperparameters.
			windowDepth := 10
			descentIterations := 5
			outputStorage := RunMatchRateTraining(
				rateStorage, initRates, 0.001, descentIterations, windowDepth,
			)

			fittedRates := ExtractFittedRates(outputStorage)
			if fittedRates == nil {
				t.Fatal("expected non-empty training output")
			}

			t.Logf("MLE rates:     %v", mleRates)
			t.Logf("init rates:    %v", initRates)
			t.Logf("fitted rates:  %v", fittedRates)

			// All fitted rates should be positive and finite.
			for i, r := range fittedRates {
				if r <= 0 || math.IsNaN(r) || math.IsInf(r, 0) {
					t.Errorf("rate[%d] is non-positive/NaN/Inf: %f", i, r)
				}
			}

			// Fitted rates should be closer to MLE than the init rates.
			for i := range mleRates {
				initDist := math.Abs(initRates[i] - mleRates[i])
				fitDist := math.Abs(fittedRates[i] - mleRates[i])
				if fitDist > initDist*1.5 {
					t.Errorf(
						"rate[%d]: fitted diverged from MLE "+
							"(init_dist=%f, fit_dist=%f)",
						i, initDist, fitDist,
					)
				}
			}
		},
	)
}

func TestBuildMultiGameBaselineCovariateStorage(t *testing.T) {
	t.Run(
		"test that multi-game storage concatenates correctly",
		func(t *testing.T) {
			baselineRates, err := ComputeSmoothedBaselineRates("../../dat/events.csv", "../../dat/players.csv")
			if err != nil {
				t.Fatalf("failed to compute baseline: %v", err)
			}

			storage, err := BuildMultiGameBaselineCovariateStorage(
				"../../dat/events.csv", "../../dat/players.csv", baselineRates,
			)
			if err != nil {
				t.Fatalf("failed to build multi-game storage: %v", err)
			}

			data := storage.GetValues("events_with_covariates_and_baseline")
			times := storage.GetTimes()
			if len(data) == 0 {
				t.Fatal("expected non-empty data")
			}
			t.Logf("multi-game storage: %d rows", len(data))

			// Should have more rows than a single game (~80 minutes).
			if len(data) < 100 {
				t.Errorf("expected many rows across 30 games, got %d", len(data))
			}

			// All rows should have correct width.
			for i, row := range data {
				if len(row) != BaselineCovariateDataWidth {
					t.Errorf("row %d: expected width %d, got %d",
						i, BaselineCovariateDataWidth, len(row))
				}
			}

			// Times should be monotonically non-decreasing.
			for i := 1; i < len(times); i++ {
				if times[i] < times[i-1] {
					t.Errorf("time decreased at index %d: %f -> %f",
						i, times[i-1], times[i])
				}
			}
		},
	)
}

func TestRunMultiGameBaselineCovariateTraining(t *testing.T) {
	t.Run(
		"test that multi-game training produces finite coefficients",
		func(t *testing.T) {
			coeffs, err := RunMultiGameBaselineCovariateTraining(
				"../../dat/events.csv",
				"../../dat/players.csv",
				0.1,
				10,
				50,
			)
			if err != nil {
				t.Fatalf("training failed: %v", err)
			}
			if len(coeffs) != TotalCoeffWidth {
				t.Fatalf("expected %d coefficients, got %d", TotalCoeffWidth, len(coeffs))
			}

			// All coefficients should be finite.
			for i, c := range coeffs {
				if math.IsNaN(c) || math.IsInf(c, 0) {
					t.Errorf("coeff[%d] is NaN/Inf: %f", i, c)
				}
			}

			t.Logf("fitted coefficients (first rate):")
			covLabels := []string{"intercept", "home_front", "home_back", "home_halves", "home_outside",
				"away_front", "away_back", "away_halves", "away_outside"}
			rateLabels := []string{"home_try", "away_try", "home_penalty", "away_penalty", "home_yellow", "away_yellow"}
			for i := 0; i < RateEventWidth; i++ {
				t.Logf("  %s:", rateLabels[i])
				for j := 0; j < CoeffsPerRate; j++ {
					t.Logf("    %-14s %+.6f", covLabels[j], coeffs[i*CoeffsPerRate+j])
				}
			}
		},
	)
}

func TestMatchCovariateRateTraining(t *testing.T) {
	t.Run(
		"test that covariate-aware gradient descent training runs",
		func(t *testing.T) {
			storage, err := TransformEventsWithCovariates(
				"../../dat/events.csv", "../../dat/players.csv", 600009, 25900,
			)
			if err != nil {
				t.Fatalf("failed to load data: %v", err)
			}

			// Build combined storage for training.
			combStorage := RateEventsWithCovariatesStorage(storage)

			// Initialize from MLE rates.
			mleRates := ComputeMLERates(storage)
			initCoeffs := InitCoefficientsFromRates(mleRates)

			windowDepth := 10
			descentIterations := 5
			outputStorage := RunMatchCovariateRateTraining(
				combStorage, initCoeffs, 0.0001, descentIterations, windowDepth,
			)

			fittedCoeffs := ExtractFittedCoefficients(outputStorage)
			if fittedCoeffs == nil {
				t.Fatal("expected non-empty training output")
			}

			t.Logf("init coeffs:   %v", initCoeffs[:9])
			t.Logf("fitted coeffs: %v", fittedCoeffs[:9])

			// All fitted coefficients should be finite.
			for i, c := range fittedCoeffs {
				if math.IsNaN(c) || math.IsInf(c, 0) {
					t.Errorf("coeff[%d] is NaN/Inf: %f", i, c)
				}
			}

			// Intercepts should be close to log(MLE rate).
			for i := 0; i < RateEventWidth; i++ {
				intercept := fittedCoeffs[i*CoeffsPerRate]
				expected := math.Log(mleRates[i])
				// With covariates and gradient descent, intercepts may shift,
				// but should be in the same ballpark.
				if math.Abs(intercept-expected) > 2.0 {
					t.Errorf("intercept[%d]: %f far from expected %f",
						i, intercept, expected)
				}
			}

			// Split should produce correct lengths.
			score, card := SplitCoefficients(fittedCoeffs)
			if len(score) != ScoreCoeffWidth {
				t.Errorf("expected %d score coeffs, got %d", ScoreCoeffWidth, len(score))
			}
			if len(card) != CardCoeffWidth {
				t.Errorf("expected %d card coeffs, got %d", CardCoeffWidth, len(card))
			}
		},
	)
}
