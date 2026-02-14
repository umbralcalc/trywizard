package match

import (
	"math"

	"github.com/umbralcalc/stochadex/pkg/general"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

const (
	ScoreRateWidth = 4 // home_try, away_try, home_penalty, away_penalty
	CardRateWidth  = 2 // home_yellow, away_yellow

	CoeffsPerRate   = 1 + SubCovWidth                // 9 (intercept + 8 covariates)
	ScoreCoeffWidth = ScoreRateWidth * CoeffsPerRate // 36
	CardCoeffWidth  = CardRateWidth * CoeffsPerRate  // 18
)

// computeRates computes rates from coefficients and optional covariates
// using a log-linear model. If covariates are present, each rate uses
// CoeffsPerRate coefficients: rate_i = exp(β₀ + Σⱼ βⱼ·xⱼ).
// Without covariates, falls back to rate_i = exp(coefficients[i]).
func computeRates(coefficients, covariates []float64, width int) []float64 {
	rates := make([]float64, width)
	if len(covariates) > 0 && len(coefficients) == width*CoeffsPerRate {
		for i := range rates {
			offset := i * CoeffsPerRate
			logRate := coefficients[offset]
			for j := 0; j < len(covariates); j++ {
				logRate += coefficients[offset+1+j] * covariates[j]
			}
			rates[i] = math.Exp(logRate)
		}
	} else {
		for i := range rates {
			rates[i] = math.Exp(coefficients[i])
		}
	}
	return rates
}

// ScoreEventRateFunction computes rates for scoring events using a
// log-linear model. With covariates from upstream: rate_i = exp(β₀ + Σβⱼxⱼ).
// Without covariates: rate_i = exp(coefficients[i]).
func ScoreEventRateFunction(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) []float64 {
	coefficients := params.Get("coefficients")
	covariates, _ := params.GetOk("covariates")
	return computeRates(coefficients, covariates, ScoreRateWidth)
}

// CardEventRateFunction computes rates for card events using a
// log-linear model. With covariates from upstream: rate_i = exp(β₀ + Σβⱼxⱼ).
// Without covariates: rate_i = exp(coefficients[i]).
func CardEventRateFunction(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) []float64 {
	coefficients := params.Get("coefficients")
	covariates, _ := params.GetOk("covariates")
	return computeRates(coefficients, covariates, CardRateWidth)
}

// NewScoreRatesPartition creates a partition that computes scoring event
// rates from coefficients. If withCovariates is true, wires upstream
// covariates from the "sub_covariates" partition.
func NewScoreRatesPartition(coefficients []float64) *simulator.PartitionConfig {
	return &simulator.PartitionConfig{
		Name: "score_rates",
		Iteration: &general.ValuesFunctionIteration{
			Function: ScoreEventRateFunction,
		},
		Params: simulator.NewParams(map[string][]float64{
			"coefficients": coefficients,
		}),
		InitStateValues:   make([]float64, ScoreRateWidth),
		StateHistoryDepth: 1,
		Seed:              0,
	}
}

// NewCardRatesPartition creates a partition that computes card event
// rates from coefficients.
func NewCardRatesPartition(coefficients []float64) *simulator.PartitionConfig {
	return &simulator.PartitionConfig{
		Name: "card_rates",
		Iteration: &general.ValuesFunctionIteration{
			Function: CardEventRateFunction,
		},
		Params: simulator.NewParams(map[string][]float64{
			"coefficients": coefficients,
		}),
		InitStateValues:   make([]float64, CardRateWidth),
		StateHistoryDepth: 1,
		Seed:              0,
	}
}
