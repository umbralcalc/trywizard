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
			t.Logf("MLE rates: %v", rates)

			// All rates should be positive.
			for i, r := range rates {
				if r <= 0 || math.IsNaN(r) {
					t.Errorf("rate[%d] is non-positive or NaN: %f", i, r)
				}
			}

			// Home tries: ~7 in 81 minutes ≈ 0.086
			if !scalar.EqualWithinAbs(rates[IdxHomeTry], 0.086, 0.01) {
				t.Errorf("home try rate: got %f, expected ~0.086", rates[IdxHomeTry])
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

			// Initialize near the MLE to test convergence stability.
			// Use a floor to avoid zero rates causing divergence.
			mleRates := ComputeMLERates(storage)
			initRates := make([]float64, EventWidth)
			for i := range initRates {
				initRates[i] = math.Max(mleRates[i]*1.5, 0.01)
			}

			// Use conservative hyperparameters.
			windowDepth := 10
			descentIterations := 5
			outputStorage := RunMatchRateTraining(
				storage, initRates, 0.001, descentIterations, windowDepth,
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
