package match

import (
	"math"

	"github.com/umbralcalc/stochadex/pkg/analysis"
	"github.com/umbralcalc/stochadex/pkg/inference"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

// ComputeMLERates computes the maximum likelihood Poisson rates from
// observed per-minute event counts. For the intercept-only model, the
// MLE is simply the mean count per minute for each event type.
// A floor is applied to prevent zero rates (which cause numerical issues).
func ComputeMLERates(storage *simulator.StateTimeStorage) []float64 {
	events := storage.GetValues("events")
	rates := make([]float64, EventWidth)
	for _, row := range events {
		for i, v := range row {
			rates[i] += v
		}
	}
	n := float64(len(events))
	floor := 1e-6
	for i := range rates {
		rates[i] /= n
		if rates[i] < floor {
			rates[i] = floor
		}
	}
	return rates
}

// ComputeLogCoefficients converts rates to log-space coefficients
// for use with the forward simulation's log-linear rate model.
func ComputeLogCoefficients(rates []float64) (scoreCoeffs, cardCoeffs []float64) {
	scoreCoeffs = make([]float64, ScoreRateWidth)
	cardCoeffs = make([]float64, CardRateWidth)
	for i := 0; i < ScoreRateWidth; i++ {
		scoreCoeffs[i] = math.Log(rates[i])
	}
	for i := 0; i < CardRateWidth; i++ {
		cardCoeffs[i] = math.Log(rates[ScoreRateWidth+i])
	}
	return scoreCoeffs, cardCoeffs
}

// NewMatchRateTrainingPartition creates a partition that fits event rates
// to observed per-minute event count data using gradient descent on the
// Poisson log-likelihood. The gradient_descent state represents the
// current rate estimates (EventWidth values).
//
// The returned partition is an embedded simulation that runs gradient
// descent (ascent on log-likelihood) for descentIterations steps each
// time the outer simulation advances.
//
// descentIterations must be <= windowDepth since the inner simulation
// replays data from the outer partition's state history.
func NewMatchRateTrainingPartition(
	storage *simulator.StateTimeStorage,
	learningRate float64,
	descentIterations int,
	windowDepth int,
) *simulator.PartitionConfig {
	return analysis.NewLikelihoodMeanFunctionFitPartition(
		analysis.AppliedLikelihoodMeanFunctionFit{
			Name: "rate_fit",
			Model: analysis.ParameterisedModelWithGradient{
				Likelihood: &inference.PoissonLikelihoodDistribution{},
				Params:     simulator.NewParams(make(map[string][]float64)),
			},
			Gradient: analysis.LikelihoodMeanGradient{
				Function: inference.MeanGradientFunc,
				Width:    EventWidth,
			},
			Data: analysis.DataRef{PartitionName: "events"},
			Window: analysis.WindowedPartitions{
				Data:  []analysis.DataRef{{PartitionName: "events"}},
				Depth: windowDepth,
			},
			LearningRate:      learningRate,
			DescentIterations: descentIterations,
		},
		storage,
	)
}

// RunMatchRateTraining runs gradient descent training to fit event rates
// and returns the output storage with fitted rate trajectories.
func RunMatchRateTraining(
	storage *simulator.StateTimeStorage,
	initRates []float64,
	learningRate float64,
	descentIterations int,
	windowDepth int,
) *simulator.StateTimeStorage {
	rateFit := NewMatchRateTrainingPartition(
		storage, learningRate, descentIterations, windowDepth,
	)
	// Set the inner gradient_descent partition's initial state.
	rateFit.Params.Set("gradient_descent/init_state_values", initRates)
	// Enable gradient ascent to maximize log-likelihood.
	rateFit.Params.Set("gradient_descent/ascent", []float64{1.0})
	// Set the outer init state's gradient_descent portion.
	copy(rateFit.InitStateValues[len(rateFit.InitStateValues)-EventWidth:], initRates)

	return analysis.AddPartitionsToStateTimeStorage(
		storage,
		[]*simulator.PartitionConfig{rateFit},
		map[string]int{"events": windowDepth},
	)
}

// ExtractFittedRates returns the final fitted rates from training output.
// The rate_fit state is [events..., gradient..., gradient_descent...];
// the last EventWidth values are the gradient_descent output (rates).
func ExtractFittedRates(
	storage *simulator.StateTimeStorage,
) []float64 {
	fitResults := storage.GetValues("rate_fit")
	if len(fitResults) == 0 {
		return nil
	}
	lastState := fitResults[len(fitResults)-1]
	return lastState[len(lastState)-EventWidth:]
}
