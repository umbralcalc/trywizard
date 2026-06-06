# trywizard

An event-based rugby match simulator written with the [stochadex framework](https://github.com/umbralcalc/stochadex). The project fits stochastic rate models to real match event data and uses them to run counterfactual simulations — answering questions like "how does substitution timing affect win probability?"

## How it works

### Match simulation

The simulator models a rugby match as a system of 8 coupled stochastic partitions running minute-by-minute over 80 minutes:

- **Baseline rates** are smoothed per-minute event rates computed from multi-game data using adaptive-bandwidth kernel smoothing, split by home/away.
- **Score events** (tries and penalties) and **card events** (yellow cards) are generated as Cox processes whose intensities are driven by a log-linear model: `rate = baseline * exp(intercept + covariate effects)`.
- **Conversions** are Bernoulli draws triggered by each new try.
- **Match state** derives cumulative scores (try = 5pts, conversion = 2pts, penalty = 3pts) and tracks active yellow cards (10-minute duration).

### Rate model & training

Event rates follow a Poisson GLM with log link. Each of the 6 rate channels (home/away tries, penalties, yellow cards) has an intercept plus 8 covariate coefficients encoding substitution status per position group (front row, back row, halves, outside backs) for each team.

Training uses warm-start stochastic gradient descent over 50 epochs across all games, maximising the Poisson log-likelihood with respect to the coefficients. The `PoissonCovariateGLMLikelihood` iteration computes gradients analytically.

### Substitution counterfactuals

With fitted coefficients, the simulator can inject alternative substitution strategies and compare outcomes over Monte Carlo runs. A `SubstitutionStrategy` specifies when each position group is substituted, generating per-minute binary covariate vectors that feed into the rate model. By sweeping substitution timing across position groups and comparing win probabilities, the model quantifies the causal effect of substitution decisions.

## Project structure

```
pkg/match/          Core simulation: rate functions, state logic, training, counterfactuals
pkg/sportdevs/      Rugby data API client (SportDevs)
cfg/                Stochadex YAML configs for simulations
dat/                Event and player CSV data (not publicly shared)
nbs/                Jupyter notebooks for analysis and visualisation
```

### Key files in `pkg/match/`

| File | Purpose |
|------|---------|
| `simulation.go` | Partition wiring for forward simulations |
| `rate_function.go` | Log-linear rate model (`ScoreEventRateFunction`, `CardEventRateFunction`) |
| `state_function.go` | Derives scores, active cards, and half from event streams |
| `baseline.go` | Adaptive-bandwidth kernel smoothing for baseline rates |
| `training.go` | Multi-game warm-start SGD training |
| `poisson_glm.go` | Poisson GLM likelihood and gradients |
| `substitution.go` | Substitution strategy definitions and covariate generation |
| `comparison.go` | Forward simulation runs and win probability computation |
| `conversion.go` | Bernoulli conversion model |
| `data.go` | CSV data loading and transformation |

## Notebooks

- **`simulation_comparison`** — Validates intercept-only vs. covariate-aware models against observed match outcomes.
- **`event_features`** — Visualises smoothed baseline event rates per minute, showing home advantage patterns and late-match yellow card increases.
- **`substitution_counterfactuals`** — Compares substitution strategies (early/standard/late timing, position group sweeps) and their effect on win probability.

## Build & run

```bash
go build ./...
go test -count=1 ./...
go run github.com/umbralcalc/stochadex/cmd/stochadex --config cfg/match_simulation.yaml
```

## Data sources & licence

The **code** in this repository is released under the [MIT License](LICENSE). That licence covers the code only — **not** the match data.

This is an entirely non-commercial project for research and learning. The underlying match data was gathered from third-party providers and remains subject to each provider's own terms — it is not ours to relicense and is **not redistributed here**. The raw CSVs are kept out of the repository (`dat/` is git-ignored); only the fitted model coefficients (`app/pkg/rugby/fittedmodel.json`) — aggregate statistical parameters with no match records — are committed.

Match data was sourced primarily from the [SportDevs](https://www.sportdevs.com/) rugby API (free tier). The following were also consulted as references during development:

- API: [https://www.sportdevs.com/](https://www.sportdevs.com/) (free tier)
- Site: [https://www.rugbypass.com/](https://www.rugbypass.com/)
- Site: [https://www.rugbydatabase.co.nz/](https://www.rugbydatabase.co.nz/)
- Site: [https://www.espn.co.uk/rugby/](https://www.espn.co.uk/rugby/)
- Site: [https://all.rugby/](https://all.rugby/)
