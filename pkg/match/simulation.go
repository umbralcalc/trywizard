package match

import (
	"github.com/umbralcalc/stochadex/pkg/discrete"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

// NewMatchSimulationPartitions returns the partition configs for a forward
// match simulation. The partitions are:
//   - score_rates: log-linear rate model for scoring events (width 4)
//   - card_rates: log-linear rate model for card events (width 2)
//   - score_events: Cox process for tries and penalties (width 4)
//   - card_events: Cox process for yellow cards (width 2)
//   - conversion_events: Bernoulli trials per new try (width 2)
//   - match_state: derived match state (scores, active cards, half)
//
// scoreCoefficients has length ScoreRateWidth (4): [try_h, try_a, penalty_h, penalty_a].
// cardCoefficients has length CardRateWidth (2).
// conversionProbabilities has length 2: [home_prob, away_prob].
func NewMatchSimulationPartitions(
	scoreCoefficients []float64,
	cardCoefficients []float64,
	conversionProbabilities []float64,
	seed uint64,
) []*simulator.PartitionConfig {
	scoreRates := NewScoreRatesPartition(scoreCoefficients)
	cardRates := NewCardRatesPartition(cardCoefficients)

	scoreEvents := &simulator.PartitionConfig{
		Name:      "score_events",
		Iteration: &discrete.CoxProcessIteration{},
		Params:    simulator.NewParams(make(map[string][]float64)),
		ParamsFromUpstream: map[string]simulator.NamedUpstreamConfig{
			"rates": {Upstream: "score_rates"},
		},
		InitStateValues:   make([]float64, ScoreRateWidth),
		StateHistoryDepth: 2,
		Seed:              seed,
	}

	cardEvents := &simulator.PartitionConfig{
		Name:      "card_events",
		Iteration: &discrete.CoxProcessIteration{},
		Params:    simulator.NewParams(make(map[string][]float64)),
		ParamsFromUpstream: map[string]simulator.NamedUpstreamConfig{
			"rates": {Upstream: "card_rates"},
		},
		InitStateValues:   make([]float64, CardRateWidth),
		StateHistoryDepth: YellowCardMinutes + 1,
		Seed:              seed + 1,
	}

	conversionEvents := &simulator.PartitionConfig{
		Name:      "conversion_events",
		Iteration: &ConversionIteration{},
		Params: simulator.NewParams(map[string][]float64{
			"conversion_probs": conversionProbabilities,
		}),
		ParamsFromUpstream: map[string]simulator.NamedUpstreamConfig{
			"try_values": {Upstream: "score_events"},
		},
		InitStateValues:   make([]float64, 2),
		StateHistoryDepth: 1,
		Seed:              seed + 2,
	}

	matchState := NewMatchStatePartition()

	return []*simulator.PartitionConfig{
		scoreRates,
		cardRates,
		scoreEvents,
		cardEvents,
		conversionEvents,
		matchState,
	}
}

// NewMatchSimulationConfigGenerator creates a ConfigGenerator for a full
// match simulation with the given coefficients and simulation parameters.
func NewMatchSimulationConfigGenerator(
	scoreCoefficients []float64,
	cardCoefficients []float64,
	conversionProbabilities []float64,
	seed uint64,
	numSteps int,
	stepSize float64,
) *simulator.ConfigGenerator {
	generator := simulator.NewConfigGenerator()
	generator.SetSimulation(&simulator.SimulationConfig{
		OutputCondition: &simulator.EveryStepOutputCondition{},
		OutputFunction:  &simulator.StdoutOutputFunction{},
		TerminationCondition: &simulator.NumberOfStepsTerminationCondition{
			MaxNumberOfSteps: numSteps,
		},
		TimestepFunction: &simulator.ConstantTimestepFunction{Stepsize: stepSize},
		InitTimeValue:    0.0,
	})
	for _, partition := range NewMatchSimulationPartitions(
		scoreCoefficients, cardCoefficients, conversionProbabilities, seed,
	) {
		generator.SetPartition(partition)
	}
	return generator
}
