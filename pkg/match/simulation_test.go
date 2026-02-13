package match

import (
	"testing"

	"github.com/umbralcalc/stochadex/pkg/discrete"
	"github.com/umbralcalc/stochadex/pkg/general"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

func TestMatchSimulation(t *testing.T) {
	t.Run(
		"test that the match simulation runs",
		func(t *testing.T) {
			scoreCoeffs := []float64{-3.0, -3.0, -3.5, -3.5}
			cardCoeffs := []float64{-4.5, -4.5}
			convProbs := []float64{0.5, 0.5}
			generator := NewMatchSimulationConfigGenerator(
				scoreCoeffs, cardCoeffs, convProbs, 7832, 80, 1.0,
			)
			store := simulator.NewStateTimeStorage()
			generator.SetSimulation(&simulator.SimulationConfig{
				OutputCondition: &simulator.EveryStepOutputCondition{},
				OutputFunction:  &simulator.StateTimeStorageOutputFunction{Store: store},
				TerminationCondition: &simulator.NumberOfStepsTerminationCondition{
					MaxNumberOfSteps: 80,
				},
				TimestepFunction: &simulator.ConstantTimestepFunction{Stepsize: 1.0},
				InitTimeValue:    0.0,
			})
			settings, implementations := generator.GenerateConfigs()
			coordinator := simulator.NewPartitionCoordinator(settings, implementations)
			coordinator.Run()

			matchStates := store.GetValues("match_state")
			if len(matchStates) == 0 {
				t.Fatal("expected non-empty match state output")
			}
			for i, state := range matchStates {
				if state[StateIdxHomeScore] < 0 {
					t.Errorf("step %d: home score negative: %f", i, state[StateIdxHomeScore])
				}
				if state[StateIdxAwayScore] < 0 {
					t.Errorf("step %d: away score negative: %f", i, state[StateIdxAwayScore])
				}
			}

			// Conversions should not exceed tries.
			scoreEvents := store.GetValues("score_events")
			convEvents := store.GetValues("conversion_events")
			if len(scoreEvents) > 0 && len(convEvents) > 0 {
				lastScore := scoreEvents[len(scoreEvents)-1]
				lastConv := convEvents[len(convEvents)-1]
				if lastConv[0] > lastScore[0] {
					t.Errorf("home conversions (%f) exceed home tries (%f)",
						lastConv[0], lastScore[0])
				}
				if lastConv[1] > lastScore[1] {
					t.Errorf("away conversions (%f) exceed away tries (%f)",
						lastConv[1], lastScore[1])
				}
			}
		},
	)
	t.Run(
		"test that the match simulation runs with harnesses",
		func(t *testing.T) {
			settings := simulator.LoadSettingsFromYaml(
				"./simulation_settings.yaml",
			)
			iterations := []simulator.Iteration{
				&general.ValuesFunctionIteration{
					Function: ScoreEventRateFunction,
				},
				&general.ValuesFunctionIteration{
					Function: CardEventRateFunction,
				},
				&discrete.CoxProcessIteration{},
				&discrete.CoxProcessIteration{},
				&ConversionIteration{},
				&general.ValuesFunctionIteration{
					Function: MatchStateFunction,
				},
			}
			store := simulator.NewStateTimeStorage()
			implementations := &simulator.Implementations{
				Iterations:      iterations,
				OutputCondition: &simulator.EveryStepOutputCondition{},
				OutputFunction:  &simulator.StateTimeStorageOutputFunction{Store: store},
				TerminationCondition: &simulator.NumberOfStepsTerminationCondition{
					MaxNumberOfSteps: 80,
				},
				TimestepFunction: &simulator.ConstantTimestepFunction{Stepsize: 1.0},
			}
			if err := simulator.RunWithHarnesses(settings, implementations); err != nil {
				t.Errorf("test harness failed: %v", err)
			}
		},
	)
}
