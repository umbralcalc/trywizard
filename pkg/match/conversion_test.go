package match

import (
	"testing"

	"github.com/umbralcalc/stochadex/pkg/discrete"
	"github.com/umbralcalc/stochadex/pkg/general"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

func TestConversionIteration(t *testing.T) {
	t.Run(
		"test conversion iteration with harness",
		func(t *testing.T) {
			settings := simulator.LoadSettingsFromYaml(
				"./conversion_settings.yaml",
			)
			iterations := []simulator.Iteration{
				&general.ParamValuesIteration{},
				&discrete.CoxProcessIteration{},
				&ConversionIteration{},
			}
			implementations := &simulator.Implementations{
				Iterations:      iterations,
				OutputCondition: &simulator.EveryStepOutputCondition{},
				OutputFunction:  &simulator.NilOutputFunction{},
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
	t.Run(
		"test conversions do not exceed tries",
		func(t *testing.T) {
			convProbs := []float64{0.7, 0.7}
			scoreCoeffs := []float64{-2.5, -2.5, -10.0, -10.0}
			cardCoeffs := []float64{-10.0, -10.0}

			results := RunMatchSimulations(
				scoreCoeffs, cardCoeffs, convProbs, 20, 80, 5000,
			)
			for i, r := range results {
				if r.EventTotals[IdxHomeConv] > r.EventTotals[IdxHomeTry] {
					t.Errorf("sim %d: home conv (%f) > home tries (%f)",
						i, r.EventTotals[IdxHomeConv], r.EventTotals[IdxHomeTry])
				}
				if r.EventTotals[IdxAwayConv] > r.EventTotals[IdxAwayTry] {
					t.Errorf("sim %d: away conv (%f) > away tries (%f)",
						i, r.EventTotals[IdxAwayConv], r.EventTotals[IdxAwayTry])
				}
			}
		},
	)
}
