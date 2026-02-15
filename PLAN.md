## Substitution counterfactual analysis

The goal of this workstream is to answer: **given a match situation, what substitution strategy maximises the probability of winning?**

### Context

Rugby union allows each team 8 substitutes per match. The bench must include replacements for all 3 front row positions (2 props + hooker). Managers choose when (and whether) to use their remaining 5 tactical subs. Injuries can force substitutions at any time, consuming the budget unpredictably.

The existing match simulator already models substitution effects:

- **Substitution covariates**: 8 binary indicators (4 position groups × 2 teams) that flip to 1 when a player in that group is substituted.
- **Poisson GLM rate model**: `rate_i = baseline_i × exp(β₀ + Σ βⱼ xⱼ)` where the `βⱼ` coefficients capture how substitutions shift scoring and card event rates.
- **Cox process events**: tries, penalties, and cards are generated from these time-varying rates.

The counterfactual idea is to replay matches under alternative substitution schedules and Monte Carlo simulate win probabilities.

### Plan

#### Phase 1 — Robust coefficient estimation

Train the covariate rate model across all 30 games (not just one) to get stable `β` estimates for substitution effects. The smoothed baseline rates already aggregate across games, so the per-game intercept adjustments and covariate coefficients can be learned jointly.

#### Phase 2 — Strategy-driven simulation

Build a substitution strategy type that defines a schedule of (minute, position_group) tuples:

```go
type SubstitutionStrategy struct {
    // Each entry is the minute at which to substitute in that group.
    // A value of 0 or >80 means "don't substitute this group".
    // Entries are subject to the constraint: total non-zero entries <= 8.
    HomeSubs [4]int // [front_row, back_row, halves, outside_backs]
    AwaySubs [4]int
}
```

A new partition (or iteration) converts this schedule into the per-minute covariate matrix that the rate model consumes, replacing the current `FromStorageIteration` replay of observed covariates.

Given a match's trained coefficients and baseline rates, the pipeline is:

1. Define a `SubstitutionStrategy`
2. Generate the covariate time series from the strategy
3. Configure and run N simulations (e.g. 500)
4. Compute P(win), P(draw), P(loss) from final score distributions

#### Phase 3 — Strategy sweep and optimisation

Sweep substitution timing for each position group while holding others fixed. This produces curves of win probability vs substitution minute per group, revealing optimal timing.

Extend to joint optimisation: given 8 sub slots and the budget constraint, search for the strategy that maximises win probability. This could be exhaustive (the space is manageable if we discretise to 5-minute intervals) or use a simple optimiser.

#### Phase 4 — Injury risk and adaptive policies

Model injuries as a stochastic process (Poisson with a per-minute rate) that forces substitutions and consumes budget. The counterfactual question becomes: **given that injuries randomly consume some subs, what tactical policy maximises win probability in expectation?**

This introduces a tradeoff: aggressive early subs improve rates now but leave the team exposed if injuries hit later. Conservative policies waste potential rate improvements but preserve flexibility.

Implementation:
- Add an injury event partition (Poisson process per position group)
- Track remaining sub budget as state
- The tactical policy becomes conditional: "sub group X at minute T _if_ budget remains >= Y"

#### Phase 5 — Richer covariates

Move beyond binary position-group indicators:
- **Sub count per group**: captures diminishing returns from multiple subs in the same group
- **Cumulative subs used**: captures overall freshness/fatigue budget effects
- **Score difference interaction**: the value of a sub may depend on whether the team is leading or trailing
- **Player-level quality**: individual player ratings so that _who_ comes on matters, not just _when_

Each extension requires retraining the rate model with additional covariate columns.

### Current architecture reference

```
baseline_rates ──────────────────────┐
sub_covariates ──┐                   │
                 ├→ score_rates ─→ score_events ─→ conversion_events ─┐
                 ├→ card_rates  ─→ card_events  ──────────────────────┤
                 │                                                     │
                 └─────────────────────────────────────────────────────┴→ match_state
```

For counterfactuals, the `sub_covariates` partition switches from replaying observed data to generating covariates from a `SubstitutionStrategy`.
