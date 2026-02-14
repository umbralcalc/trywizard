package match

import (
	"math"

	"github.com/umbralcalc/stochadex/pkg/analysis"
	"github.com/umbralcalc/stochadex/pkg/inference"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

// RateEventWidth is the number of rate-based event types (excluding
// conversions, which are modeled as Bernoulli trials, not Poisson rates).
// Order: [home_try, away_try, home_penalty, away_penalty, home_yellow, away_yellow].
const RateEventWidth = 6

// rateEventIndices maps from rate index to full event vector index.
var rateEventIndices = []int{
	IdxHomeTry, IdxAwayTry,
	IdxHomePenalty, IdxAwayPenalty,
	IdxHomeYellow, IdxAwayYellow,
}

// ComputeMLERates computes the maximum likelihood Poisson rates from
// observed per-minute event counts for rate-based events only (tries,
// penalties, yellow cards). Conversions are excluded since they are
// Bernoulli trials, not Poisson processes.
// Returns RateEventWidth (6) values.
// A floor is applied to prevent zero rates (which cause numerical issues).
func ComputeMLERates(storage *simulator.StateTimeStorage) []float64 {
	events := storage.GetValues("events")
	rates := make([]float64, RateEventWidth)
	for _, row := range events {
		for i, idx := range rateEventIndices {
			rates[i] += row[idx]
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
// Takes RateEventWidth (6) rates, returns score (4) and card (2) coefficients.
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

// RateEventsStorage creates a StateTimeStorage containing only rate-based
// event columns (tries, penalties, yellows) for use with gradient descent
// training. Conversions are excluded.
func RateEventsStorage(storage *simulator.StateTimeStorage) *simulator.StateTimeStorage {
	events := storage.GetValues("events")
	times := storage.GetTimes()
	rateStorage := simulator.NewStateTimeStorage()
	for i, row := range events {
		rateRow := make([]float64, RateEventWidth)
		for j, idx := range rateEventIndices {
			rateRow[j] = row[idx]
		}
		rateStorage.ConcurrentAppend("events", times[i], rateRow)
	}
	return rateStorage
}

// NewMatchRateTrainingPartition creates a partition that fits event rates
// to observed per-minute event count data using gradient descent on the
// Poisson log-likelihood. The gradient_descent state represents the
// current rate estimates (RateEventWidth values).
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
				Width:    RateEventWidth,
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
// The input storage should contain only rate-based event columns
// (use RateEventsStorage to filter the full event data).
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
	copy(rateFit.InitStateValues[len(rateFit.InitStateValues)-RateEventWidth:], initRates)

	return analysis.AddPartitionsToStateTimeStorage(
		storage,
		[]*simulator.PartitionConfig{rateFit},
		map[string]int{"events": windowDepth},
	)
}

// ExtractFittedRates returns the final fitted rates from training output.
// The rate_fit state is [events..., gradient..., gradient_descent...];
// the last RateEventWidth values are the gradient_descent output (rates).
func ExtractFittedRates(
	storage *simulator.StateTimeStorage,
) []float64 {
	fitResults := storage.GetValues("rate_fit")
	if len(fitResults) == 0 {
		return nil
	}
	lastState := fitResults[len(fitResults)-1]
	return lastState[len(lastState)-RateEventWidth:]
}

// CovariateDataWidth is the width of the combined data partition used for
// covariate-aware training: [rate_event_counts(6), covariates(8)].
const CovariateDataWidth = RateEventWidth + SubCovWidth

// TotalCoeffWidth is the total number of β coefficients across all rate
// types: 6 rates × 9 coefficients each = 54.
const TotalCoeffWidth = RateEventWidth * CoeffsPerRate

// RateEventsWithCovariatesStorage creates a StateTimeStorage containing
// combined [rate_event_counts, sub_covariates] rows for covariate-aware
// training. The partition is named "events_with_covariates".
func RateEventsWithCovariatesStorage(
	storage *simulator.StateTimeStorage,
) *simulator.StateTimeStorage {
	events := storage.GetValues("events")
	covariates := storage.GetValues("sub_covariates")
	times := storage.GetTimes()
	combined := simulator.NewStateTimeStorage()
	for i, row := range events {
		combRow := make([]float64, CovariateDataWidth)
		for j, idx := range rateEventIndices {
			combRow[j] = row[idx]
		}
		if i < len(covariates) {
			copy(combRow[RateEventWidth:], covariates[i])
		}
		combined.ConcurrentAppend("events_with_covariates", times[i], combRow)
	}
	return combined
}

// InitCoefficientsFromRates creates an initial coefficient vector for the
// covariate model from MLE rates. Intercepts are set to log(rate), all
// covariate coefficients are set to 0.
func InitCoefficientsFromRates(rates []float64) []float64 {
	coeffs := make([]float64, TotalCoeffWidth)
	for i := 0; i < RateEventWidth; i++ {
		coeffs[i*CoeffsPerRate] = math.Log(rates[i])
	}
	return coeffs
}

// SplitCoefficients splits a flat coefficient vector (length TotalCoeffWidth=54)
// into score coefficients (length ScoreCoeffWidth=36) and card coefficients
// (length CardCoeffWidth=18).
func SplitCoefficients(coeffs []float64) (scoreCoeffs, cardCoeffs []float64) {
	scoreCoeffs = make([]float64, ScoreCoeffWidth)
	cardCoeffs = make([]float64, CardCoeffWidth)
	copy(scoreCoeffs, coeffs[:ScoreCoeffWidth])
	copy(cardCoeffs, coeffs[ScoreCoeffWidth:])
	return scoreCoeffs, cardCoeffs
}

// NewMatchCovariateRateTrainingPartition creates a partition that fits
// event rate coefficients (intercepts + covariate effects) using gradient
// descent on the Poisson GLM log-likelihood. The gradient_descent state
// represents the current coefficient estimates (TotalCoeffWidth values).
//
// The input storage must contain an "events_with_covariates" partition
// (use RateEventsWithCovariatesStorage to build it).
func NewMatchCovariateRateTrainingPartition(
	storage *simulator.StateTimeStorage,
	learningRate float64,
	descentIterations int,
	windowDepth int,
) *simulator.PartitionConfig {
	return analysis.NewLikelihoodMeanFunctionFitPartition(
		analysis.AppliedLikelihoodMeanFunctionFit{
			Name: "covariate_rate_fit",
			Model: analysis.ParameterisedModelWithGradient{
				Likelihood: &PoissonCovariateGLMLikelihood{
					NRates: RateEventWidth,
					NCovs:  SubCovWidth,
				},
				Params: simulator.NewParams(make(map[string][]float64)),
			},
			Gradient: analysis.LikelihoodMeanGradient{
				Function: inference.MeanGradientFunc,
				Width:    TotalCoeffWidth,
			},
			Data: analysis.DataRef{PartitionName: "events_with_covariates"},
			Window: analysis.WindowedPartitions{
				Data:  []analysis.DataRef{{PartitionName: "events_with_covariates"}},
				Depth: windowDepth,
			},
			LearningRate:      learningRate,
			DescentIterations: descentIterations,
		},
		storage,
	)
}

// RunMatchCovariateRateTraining runs gradient descent training to fit
// event rate coefficients (intercepts + covariate effects) and returns
// the output storage with fitted coefficient trajectories.
// The input storage should contain "events_with_covariates"
// (use RateEventsWithCovariatesStorage to build it).
func RunMatchCovariateRateTraining(
	storage *simulator.StateTimeStorage,
	initCoeffs []float64,
	learningRate float64,
	descentIterations int,
	windowDepth int,
) *simulator.StateTimeStorage {
	fit := NewMatchCovariateRateTrainingPartition(
		storage, learningRate, descentIterations, windowDepth,
	)
	fit.Params.Set("gradient_descent/init_state_values", initCoeffs)
	fit.Params.Set("gradient_descent/ascent", []float64{1.0})
	copy(fit.InitStateValues[len(fit.InitStateValues)-TotalCoeffWidth:], initCoeffs)

	return analysis.AddPartitionsToStateTimeStorage(
		storage,
		[]*simulator.PartitionConfig{fit},
		map[string]int{"events_with_covariates": windowDepth},
	)
}

// ExtractFittedCoefficients returns the final fitted coefficients from
// covariate-aware training output. The covariate_rate_fit state is
// [events_with_covariates..., gradient..., gradient_descent...];
// the last TotalCoeffWidth values are the gradient_descent output.
func ExtractFittedCoefficients(
	storage *simulator.StateTimeStorage,
) []float64 {
	fitResults := storage.GetValues("covariate_rate_fit")
	if len(fitResults) == 0 {
		return nil
	}
	lastState := fitResults[len(fitResults)-1]
	return lastState[len(lastState)-TotalCoeffWidth:]
}
