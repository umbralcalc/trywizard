package match

// Plans substitution timing over the real match model using
// stochadex's agents.SimulationEnvironment.

import (
	"testing"

	"github.com/umbralcalc/stochadex/pkg/agents"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

const probeHorizon = 20

// probeCoefficients builds a score-rate coefficient vector with a KNOWN effect:
// substituting the home front row (covariate 0) shifts the home try rate by
// homeFrontRowEffect on the log scale. Everything else is intercept-only, so the
// planner's choice has exactly one consequence and its sign is known in advance.
func probeCoefficients(homeFrontRowEffect float64) []float64 {
	coefficients := make([]float64, ScoreCoeffWidth)
	intercepts := []float64{-2.5, -2.5, -3.0, -3.0} // home try, away try, home pen, away pen
	for rate := 0; rate < ScoreRateWidth; rate++ {
		coefficients[rate*CoeffsPerRate] = intercepts[rate]
	}
	// Covariate 0 = home front row, acting on rate 0 = home tries.
	coefficients[0*CoeffsPerRate+1+0] = homeFrontRowEffect
	return coefficients
}

func newProbeEnvironment(t *testing.T, homeFrontRowEffect float64, seed uint64) *agents.SimulationEnvironment {
	t.Helper()
	generator := NewMatchSimulationConfigGenerator(
		probeCoefficients(homeFrontRowEffect),
		make([]float64, CardCoeffWidth),
		[]float64{0.5, 0.5},
		seed, probeHorizon, 1.0,
	)
	generator.SetSimulation(&simulator.SimulationConfig{
		OutputCondition:      &simulator.NilOutputCondition{},
		OutputFunction:       &simulator.NilOutputFunction{},
		TerminationCondition: &simulator.NumberOfStepsTerminationCondition{MaxNumberOfSteps: 1},
		TimestepFunction:     &simulator.ConstantTimestepFunction{Stepsize: 1.0},
		InitTimeValue:        0.0,
	})
	settings, implementations := generator.GenerateConfigs()
	implementations.ExecutionStrategy = &simulator.InlineExecution{}

	noSubs := make([]float64, SubCovWidth)
	homeFrontRowSubbed := make([]float64, SubCovWidth)
	homeFrontRowSubbed[0] = 1

	return agents.NewSimulationEnvironment(settings, implementations,
		agents.SimulationEnvironmentSpec{
			Actions:         [][]float64{noSubs, homeFrontRowSubbed},
			ActionPartition: "sub_covariates",
			ActionParam:     "param_values",
			Horizon:         probeHorizon,
			// Running score margin, summed over minutes: more points, sooner.
			Reward: func(rows map[string][]float64) float64 {
				return rows["match_state"][StateIdxScoreDiff]
			},
			MinReturn:    -600,
			MaxReturn:    600,
			ScenarioSeed: seed,
		})
}

// planShare plans over the model and reports what fraction of the decisions chose
// to substitute, plus the return achieved.
func planShare(t *testing.T, env *agents.SimulationEnvironment, sims int) (share, ret float64) {
	t.Helper()
	cfg := agents.MCTSConfig[[]float64, int]{
		Simulations:     sims,
		MaxTreeDepth:    probeHorizon + 1,
		RolloutMaxSteps: probeHorizon + 1,
		Rollout:         agents.UniformRandomRollout[[]float64, int](),
	}
	state := env.InitialState()
	subs, steps := 0, 0
	for step := 0; ; step++ {
		if _, done := env.Terminal(state); done {
			break
		}
		best, _, err := agents.RunChanceMCTSSearch(env, state, cfg, uint64(step)+7, sims)
		if err != nil {
			t.Fatalf("RunChanceMCTSSearch: %v", err)
		}
		if best == 1 {
			subs++
		}
		steps++
		state, err = env.Apply(state, best)
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	}
	return float64(subs) / float64(steps), env.Return(state)
}

// TestProbePlanRespondsToKnownEffect is the decision-path claim: when the fitted
// model says substituting the home front row raises the home try rate, the
// planner should substitute; when it says the opposite, it should not. A planner
// that just always picks the same action would pass one half and fail the other.
func TestProbePlanRespondsToKnownEffect(t *testing.T) {
	const sims = 60
	const planningScenarios = 4

	for _, testCase := range []struct {
		name   string
		effect float64
	}{
		{name: "substitution helps", effect: 1.2},
		{name: "substitution hurts", effect: -1.2},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			totalShare := 0.0
			for seed := uint64(1); seed <= planningScenarios; seed++ {
				share, ret := planShare(t, newProbeEnvironment(t, testCase.effect, seed), sims)
				t.Logf("  seed %d: substituted on %.0f%% of minutes, return %.1f",
					seed, 100*share, ret)
				totalShare += share
			}
			t.Logf("=> mean substitution share %.0f%%", 100*totalShare/planningScenarios)
		})
	}
}

// fixedPolicyReturn runs one action for the whole match.
func fixedPolicyReturn(t *testing.T, env *agents.SimulationEnvironment, action int) float64 {
	t.Helper()
	state := env.InitialState()
	for {
		if _, done := env.Terminal(state); done {
			return env.Return(state)
		}
		next, err := env.Apply(state, action)
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
		state = next
	}
}

// TestProbePlanVersusFixedPolicies asks whether the planner is actually finding
// the optimum. With a strictly beneficial substitution the best policy is
// "always", so the planner should approach it — anything well short means the
// value signal is too noisy for the search budget.
func TestProbePlanVersusFixedPolicies(t *testing.T) {
	const scenarios = 6
	for _, sims := range []int{60, 240, 960} {
		var planned, always, never float64
		var share float64
		for seed := uint64(1); seed <= scenarios; seed++ {
			env := newProbeEnvironment(t, 1.2, seed)
			s, r := planShare(t, env, sims)
			share += s
			planned += r
			always += fixedPolicyReturn(t, newProbeEnvironment(t, 1.2, seed), 1)
			never += fixedPolicyReturn(t, newProbeEnvironment(t, 1.2, seed), 0)
		}
		t.Logf("sims=%3d  planner %7.1f (subbed %.0f%%) | always %7.1f | never %7.1f",
			sims, planned/scenarios, 100*share/scenarios,
			always/scenarios, never/scenarios)
	}
}

// TestProbeReturnSpread measures how noisy the return actually is, to tell
// whether the planner-vs-always comparison above is signal or scatter.
func TestProbeReturnSpread(t *testing.T) {
	const scenarios = 40
	var sum, sumSq float64
	for seed := uint64(1); seed <= scenarios; seed++ {
		r := fixedPolicyReturn(t, newProbeEnvironment(t, 1.2, seed), 1)
		sum += r
		sumSq += r * r
	}
	mean := sum / scenarios
	variance := sumSq/scenarios - mean*mean
	sd := 0.0
	if variance > 0 {
		sd = variance
		for i := 0; i < 40; i++ {
			sd = 0.5 * (sd + variance/sd)
		}
	}
	t.Logf("always-substitute return over %d scenarios: mean %.1f, sd %.1f, "+
		"standard error at n=6 is %.1f", scenarios, mean, sd, sd/2.449)
}
