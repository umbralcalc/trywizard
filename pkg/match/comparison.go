package match

import (
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

// ScorePoint holds a home and away score at a point in time.
type ScorePoint struct{ Home, Away float64 }

// SimulationResult holds the output of a single forward simulation.
type SimulationResult struct {
	// ScoreTrajectory is the home/away score at each simulated minute.
	ScoreTrajectory []ScorePoint
	// EventTotals is the final cumulative event count per type (length EventWidth).
	EventTotals []float64
}

// extractResult extracts a SimulationResult from a completed simulation store.
func extractResult(store *simulator.StateTimeStorage) SimulationResult {
	states := store.GetValues("match_state")
	trajectory := make([]ScorePoint, len(states))
	for j, s := range states {
		trajectory[j] = ScorePoint{
			Home: s[StateIdxHomeScore],
			Away: s[StateIdxAwayScore],
		}
	}

	totals := make([]float64, EventWidth)
	scoreEvents := store.GetValues("score_events")
	convEvents := store.GetValues("conversion_events")
	cardEvents := store.GetValues("card_events")
	if len(scoreEvents) > 0 {
		last := scoreEvents[len(scoreEvents)-1]
		totals[IdxHomeTry] = last[0]
		totals[IdxAwayTry] = last[1]
		totals[IdxHomePenalty] = last[2]
		totals[IdxAwayPenalty] = last[3]
	}
	if len(convEvents) > 0 {
		last := convEvents[len(convEvents)-1]
		totals[IdxHomeConv] = last[0]
		totals[IdxAwayConv] = last[1]
	}
	if len(cardEvents) > 0 {
		last := cardEvents[len(cardEvents)-1]
		totals[IdxHomeYellow] = last[0]
		totals[IdxAwayYellow] = last[1]
	}

	return SimulationResult{
		ScoreTrajectory: trajectory,
		EventTotals:     totals,
	}
}

// RunMatchSimulations runs nSims forward simulations with intercept-only
// coefficients and returns the results. Each simulation uses seed
// baseSeed+i for i in [0, nSims).
// If baselineRates is non-nil, it is replayed as time-varying baseline rates.
func RunMatchSimulations(
	scoreCoeffs, cardCoeffs, convProbs []float64,
	baselineRates [][]float64,
	nSims, nSteps int,
	baseSeed uint64,
) []SimulationResult {
	results := make([]SimulationResult, nSims)
	for i := 0; i < nSims; i++ {
		store := simulator.NewStateTimeStorage()
		var baselinePartition *simulator.PartitionConfig
		if baselineRates != nil {
			baselinePartition = NewBaselineRatesReplayPartition(baselineRates)
		} else {
			baselinePartition = NewBaselineRatesConstantPartition()
		}
		generator := NewMatchSimulationConfigGeneratorWithCovariates(
			scoreCoeffs, cardCoeffs, convProbs,
			baselinePartition,
			NewSubCovariatesConstantPartition(make([]float64, SubCovWidth)),
			baseSeed+uint64(i), nSteps, 1.0,
		)
		generator.SetSimulation(&simulator.SimulationConfig{
			OutputCondition: &simulator.EveryStepOutputCondition{},
			OutputFunction:  &simulator.StateTimeStorageOutputFunction{Store: store},
			TerminationCondition: &simulator.NumberOfStepsTerminationCondition{
				MaxNumberOfSteps: nSteps,
			},
			TimestepFunction: &simulator.ConstantTimestepFunction{Stepsize: 1.0},
			InitTimeValue:    0.0,
		})
		settings, implementations := generator.GenerateConfigs()
		coordinator := simulator.NewPartitionCoordinator(settings, implementations)
		coordinator.Run()
		results[i] = extractResult(store)
	}
	return results
}

// RunMatchSimulationsWithCovariates runs nSims forward simulations with
// covariate-aware coefficients, a substitution covariate partition, and
// optional time-varying baseline rates.
func RunMatchSimulationsWithCovariates(
	scoreCoeffs, cardCoeffs, convProbs []float64,
	baselinePartition *simulator.PartitionConfig,
	subCovPartition *simulator.PartitionConfig,
	nSims, nSteps int,
	baseSeed uint64,
) []SimulationResult {
	results := make([]SimulationResult, nSims)
	for i := 0; i < nSims; i++ {
		store := simulator.NewStateTimeStorage()
		generator := NewMatchSimulationConfigGeneratorWithCovariates(
			scoreCoeffs, cardCoeffs, convProbs,
			baselinePartition, subCovPartition,
			baseSeed+uint64(i), nSteps, 1.0,
		)
		generator.SetSimulation(&simulator.SimulationConfig{
			OutputCondition: &simulator.EveryStepOutputCondition{},
			OutputFunction:  &simulator.StateTimeStorageOutputFunction{Store: store},
			TerminationCondition: &simulator.NumberOfStepsTerminationCondition{
				MaxNumberOfSteps: nSteps,
			},
			TimestepFunction: &simulator.ConstantTimestepFunction{Stepsize: 1.0},
			InitTimeValue:    0.0,
		})
		settings, implementations := generator.GenerateConfigs()
		coordinator := simulator.NewPartitionCoordinator(settings, implementations)
		coordinator.Run()
		results[i] = extractResult(store)
	}
	return results
}

// ComputeScoreTrajectory builds a cumulative home/away score trajectory
// from the per-minute event counts in storage.
func ComputeScoreTrajectory(storage *simulator.StateTimeStorage) []ScorePoint {
	events := storage.GetValues("events")
	trajectory := make([]ScorePoint, len(events))
	var cumHome, cumAway float64
	for i, ev := range events {
		cumHome += ev[IdxHomeTry]*5.0 + ev[IdxHomeConv]*2.0 + ev[IdxHomePenalty]*3.0
		cumAway += ev[IdxAwayTry]*5.0 + ev[IdxAwayConv]*2.0 + ev[IdxAwayPenalty]*3.0
		trajectory[i] = ScorePoint{Home: cumHome, Away: cumAway}
	}
	return trajectory
}

// ComputeEventTotals sums per-minute event counts into total counts
// per event type.
func ComputeEventTotals(storage *simulator.StateTimeStorage) []float64 {
	events := storage.GetValues("events")
	totals := make([]float64, EventWidth)
	for _, ev := range events {
		for j := 0; j < EventWidth; j++ {
			totals[j] += ev[j]
		}
	}
	return totals
}
