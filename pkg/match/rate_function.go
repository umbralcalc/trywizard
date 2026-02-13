package match

import (
	"math"

	"github.com/umbralcalc/stochadex/pkg/general"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

const (
	ScoreRateWidth = 4 // home_try, away_try, home_penalty, away_penalty
	CardRateWidth  = 2 // home_yellow, away_yellow
)

// ScoreEventRateFunction computes rates for scoring events using a
// log-linear model: rate_i = exp(coefficients[i]).
func ScoreEventRateFunction(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) []float64 {
	coefficients := params.Get("coefficients")
	rates := make([]float64, ScoreRateWidth)
	for i := range rates {
		rates[i] = math.Exp(coefficients[i])
	}
	return rates
}

// CardEventRateFunction computes rates for card events using a
// log-linear model: rate_i = exp(coefficients[i]).
func CardEventRateFunction(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) []float64 {
	coefficients := params.Get("coefficients")
	rates := make([]float64, CardRateWidth)
	for i := range rates {
		rates[i] = math.Exp(coefficients[i])
	}
	return rates
}

// NewScoreRatesPartition creates a partition that computes scoring event
// rates from coefficients.
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
