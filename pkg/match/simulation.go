package match

import (
	"github.com/umbralcalc/stochadex/pkg/discrete"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

// NewMatchSimulationPartitionsWithCovariates returns the partition configs
// for a forward match simulation with substitution covariates and optional
// time-varying baseline rates. 8 partitions:
//   - baseline_rates: smoothed baseline rates (replayed or zero-constant)
//   - sub_covariates: substitution covariate data (replayed or constant)
//   - score_rates: log-linear rate model for scoring events (width 4)
//   - card_rates: log-linear rate model for card events (width 2)
//   - score_events: Cox process for tries and penalties (width 4)
//   - card_events: Cox process for yellow cards (width 2)
//   - conversion_events: Bernoulli trials per new try (width 2)
//   - match_state: derived match state (scores, active cards, half)
//
// scoreCoefficients has length ScoreCoeffWidth (36) with CoeffsPerRate stride.
// cardCoefficients has length CardCoeffWidth (18) with CoeffsPerRate stride.
// conversionProbabilities has length 2: [home_prob, away_prob].
func NewMatchSimulationPartitionsWithCovariates(
	scoreCoefficients []float64,
	cardCoefficients []float64,
	conversionProbabilities []float64,
	baselinePartition *simulator.PartitionConfig,
	subCovPartition *simulator.PartitionConfig,
	seed uint64,
) []*simulator.PartitionConfig {
	scoreRates := NewScoreRatesPartition(scoreCoefficients)
	scoreRates.ParamsFromUpstream = map[string]simulator.NamedUpstreamConfig{
		"covariates": {Upstream: "sub_covariates"},
		"baseline":   {Upstream: "baseline_rates"},
	}

	cardRates := NewCardRatesPartition(cardCoefficients)
	cardRates.ParamsFromUpstream = map[string]simulator.NamedUpstreamConfig{
		"covariates": {Upstream: "sub_covariates"},
		"baseline":   {Upstream: "baseline_rates"},
	}

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
		ParamsAsPartitions: map[string][]string{
			"score_events_partition": {"score_events"},
		},
		InitStateValues:   make([]float64, 2),
		StateHistoryDepth: 1,
		Seed:              seed + 2,
	}

	matchState := NewMatchStatePartition()

	return []*simulator.PartitionConfig{
		baselinePartition,
		subCovPartition,
		scoreRates,
		cardRates,
		scoreEvents,
		cardEvents,
		conversionEvents,
		matchState,
	}
}

// NewMatchSimulationPartitions returns the partition configs for a forward
// match simulation without covariates or baseline. Backward-compatible
// wrapper that uses zero-constant covariate and baseline partitions.
//
// scoreCoefficients has length ScoreRateWidth (4): intercept-only.
// cardCoefficients has length CardRateWidth (2): intercept-only.
func NewMatchSimulationPartitions(
	scoreCoefficients []float64,
	cardCoefficients []float64,
	conversionProbabilities []float64,
	seed uint64,
) []*simulator.PartitionConfig {
	baseline := NewBaselineRatesConstantPartition()
	subCov := NewSubCovariatesConstantPartition(make([]float64, SubCovWidth))
	return NewMatchSimulationPartitionsWithCovariates(
		scoreCoefficients, cardCoefficients, conversionProbabilities,
		baseline, subCov, seed,
	)
}

// NewMatchSimulationConfigGeneratorWithCovariates creates a ConfigGenerator
// for a full match simulation with substitution covariates and optional
// time-varying baseline rates.
func NewMatchSimulationConfigGeneratorWithCovariates(
	scoreCoefficients []float64,
	cardCoefficients []float64,
	conversionProbabilities []float64,
	baselinePartition *simulator.PartitionConfig,
	subCovPartition *simulator.PartitionConfig,
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
	for _, partition := range NewMatchSimulationPartitionsWithCovariates(
		scoreCoefficients, cardCoefficients, conversionProbabilities,
		baselinePartition, subCovPartition, seed,
	) {
		generator.SetPartition(partition)
	}
	return generator
}

// NewMatchSimulationConfigGenerator creates a ConfigGenerator for a full
// match simulation without covariates or baseline. Backward-compatible wrapper.
func NewMatchSimulationConfigGenerator(
	scoreCoefficients []float64,
	cardCoefficients []float64,
	conversionProbabilities []float64,
	seed uint64,
	numSteps int,
	stepSize float64,
) *simulator.ConfigGenerator {
	baseline := NewBaselineRatesConstantPartition()
	subCov := NewSubCovariatesConstantPartition(make([]float64, SubCovWidth))
	return NewMatchSimulationConfigGeneratorWithCovariates(
		scoreCoefficients, cardCoefficients, conversionProbabilities,
		baseline, subCov, seed, numSteps, stepSize,
	)
}
