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
			generator := NewMatchSimulationConfigGenerator(
				scoreCoeffs, cardCoeffs, 7832, 80, 1.0,
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
