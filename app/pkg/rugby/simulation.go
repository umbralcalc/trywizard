package rugby

import (
	"github.com/umbralcalc/stochadex/pkg/discrete"
	"github.com/umbralcalc/stochadex/pkg/simulator"
	"github.com/umbralcalc/trywizard/pkg/match"
)

// MatchesPerRun bounds how many matches the dashboard runs before the
// simulation terminates. At MatchMinutes (80) steps per match and an inline
// driver tick of 30 ms, MatchesPerRun = 200 gives roughly 8 minutes of
// continuous viewing; the histogram will have stabilised well before that.
const MatchesPerRun = 200

// AwayReferenceSubMinutes is the fixed substitution strategy the away
// team plays against. Same minute for all four groups makes the home
// reader's comparisons against a single reference clean.
var AwayReferenceSubMinutes = []float64{55, 55, 55, 55}

// DefaultHomeSubMinutes is the per-position-group default the sliders
// open at. Picked to bracket the conventional "around 50–60 minutes"
// range the plan describes.
var DefaultHomeSubMinutes = []float64{55, 50, 60, 55}

// BuildRugbySimulation constructs the stochadex generator for the rugby
// dashboard. Eleven partitions, wired in declaration order:
//
//	home_sub_snapshot   action-state input from the sliders, latched per match
//	sub_covariates      8-wide binary covariate vector keyed off m=t%80
//	baseline_rates      smoothed per-minute baseline rates, cycled by m
//	score_rates         log-linear scoring rates (existing match package)
//	card_rates          log-linear yellow-card rates (existing)
//	score_events        Cox process for tries+penalties (existing)
//	card_events         Cox process for yellow cards (existing)
//	conversion_events   Bernoulli conversion trials (existing)
//	match_state         match-cycling in-match state derivation
//	outcomes            running histogram + win count over completed matches
//	histogram_bars      (x, y, w, h) projection of bin counts for the canvas
//
// The fitted model is embedded; see embed.go.
func BuildRugbySimulation() *simulator.ConfigGenerator {
	model := LoadFittedModel()

	homeSubSnapshot := &simulator.PartitionConfig{
		Name:      "home_sub_snapshot",
		Iteration: &HomeSubSnapshotIteration{},
		Params: simulator.NewParams(map[string][]float64{
			"action_state_values": cloneFloats(DefaultHomeSubMinutes),
		}),
		InitStateValues:   cloneFloats(DefaultHomeSubMinutes),
		StateHistoryDepth: 1,
		Seed:              0,
	}

	subCovariates := &simulator.PartitionConfig{
		Name:      "sub_covariates",
		Iteration: &SubCovariatesFromActionIteration{},
		Params: simulator.NewParams(map[string][]float64{
			"away_sub_minutes": cloneFloats(AwayReferenceSubMinutes),
		}),
		ParamsFromUpstream: map[string]simulator.NamedUpstreamConfig{
			"home_sub_minutes": {Upstream: "home_sub_snapshot"},
		},
		InitStateValues:   make([]float64, match.SubCovWidth),
		StateHistoryDepth: 1,
		Seed:              0,
	}

	baselineRates := &simulator.PartitionConfig{
		Name:      "baseline_rates",
		Iteration: &BaselineRatesByMinuteIteration{},
		Params: simulator.NewParams(map[string][]float64{
			"rates_by_minute": flattenBaseline(model.BaselineRates),
			"n_minutes":       {float64(len(model.BaselineRates))},
		}),
		InitStateValues:   make([]float64, match.RateEventWidth),
		StateHistoryDepth: 1,
		Seed:              0,
	}

	scoreRates := match.NewScoreRatesPartition(cloneFloats(model.ScoreCoefficients))
	scoreRates.ParamsFromUpstream = map[string]simulator.NamedUpstreamConfig{
		"covariates": {Upstream: "sub_covariates"},
		"baseline":   {Upstream: "baseline_rates"},
	}

	cardRates := match.NewCardRatesPartition(cloneFloats(model.CardCoefficients))
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
		InitStateValues:   make([]float64, match.ScoreRateWidth),
		StateHistoryDepth: 2,
		Seed:              7832,
	}

	cardEvents := &simulator.PartitionConfig{
		Name:      "card_events",
		Iteration: &discrete.CoxProcessIteration{},
		Params:    simulator.NewParams(make(map[string][]float64)),
		ParamsFromUpstream: map[string]simulator.NamedUpstreamConfig{
			"rates": {Upstream: "card_rates"},
		},
		InitStateValues:   make([]float64, match.CardRateWidth),
		StateHistoryDepth: match.YellowCardMinutes + 1,
		Seed:              7833,
	}

	conversionEvents := &simulator.PartitionConfig{
		Name:      "conversion_events",
		Iteration: &match.ConversionIteration{},
		Params: simulator.NewParams(map[string][]float64{
			"conversion_probs": cloneFloats(model.ConversionProbs),
		}),
		ParamsFromUpstream: map[string]simulator.NamedUpstreamConfig{
			"try_values": {Upstream: "score_events"},
		},
		ParamsAsPartitions: map[string][]string{
			"score_events_partition": {"score_events"},
		},
		InitStateValues:   make([]float64, 2),
		StateHistoryDepth: 1,
		Seed:              7834,
	}

	matchState := &simulator.PartitionConfig{
		Name:      "match_state",
		Iteration: &MatchCyclingStateIteration{},
		Params:    simulator.NewParams(make(map[string][]float64)),
		ParamsFromUpstream: map[string]simulator.NamedUpstreamConfig{
			"score_values":      {Upstream: "score_events"},
			"conversion_values": {Upstream: "conversion_events"},
			"card_values":       {Upstream: "card_events"},
		},
		ParamsAsPartitions: map[string][]string{
			"card_partition": {"card_events"},
		},
		InitStateValues:   make([]float64, MatchStateWidth),
		StateHistoryDepth: 1,
		Seed:              0,
	}

	outcomes := &simulator.PartitionConfig{
		Name:      "outcomes",
		Iteration: &OutcomesIteration{},
		Params:    simulator.NewParams(make(map[string][]float64)),
		ParamsFromUpstream: map[string]simulator.NamedUpstreamConfig{
			"match_state": {Upstream: "match_state"},
		},
		InitStateValues:   make([]float64, OutcomesWidth),
		StateHistoryDepth: 1,
		Seed:              0,
	}

	histogramBars := &simulator.PartitionConfig{
		Name:      "histogram_bars",
		Iteration: &HistogramBarsIteration{},
		Params:    simulator.NewParams(make(map[string][]float64)),
		ParamsFromUpstream: map[string]simulator.NamedUpstreamConfig{
			"outcomes_values": {Upstream: "outcomes"},
		},
		InitStateValues:   make([]float64, HistogramBarsLen),
		StateHistoryDepth: 1,
		Seed:              0,
	}

	gen := simulator.NewConfigGenerator()
	for _, p := range []*simulator.PartitionConfig{
		homeSubSnapshot,
		subCovariates,
		baselineRates,
		scoreRates,
		cardRates,
		scoreEvents,
		cardEvents,
		conversionEvents,
		matchState,
		outcomes,
		histogramBars,
	} {
		gen.SetPartition(p)
	}

	gen.SetSimulation(&simulator.SimulationConfig{
		OutputCondition: &simulator.EveryStepOutputCondition{},
		TerminationCondition: &simulator.NumberOfStepsTerminationCondition{
			MaxNumberOfSteps: MatchMinutes * MatchesPerRun,
		},
		TimestepFunction: &simulator.ConstantTimestepFunction{Stepsize: 1.0},
		InitTimeValue:    0.0,
	})
	return gen
}

func cloneFloats(src []float64) []float64 {
	dst := make([]float64, len(src))
	copy(dst, src)
	return dst
}

func flattenBaseline(rows [][]float64) []float64 {
	if len(rows) == 0 {
		return nil
	}
	out := make([]float64, 0, len(rows)*len(rows[0]))
	for _, r := range rows {
		out = append(out, r...)
	}
	return out
}
