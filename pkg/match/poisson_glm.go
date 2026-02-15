package match

import (
	"math"
	"math/rand/v2"

	"github.com/umbralcalc/stochadex/pkg/inference"
	"github.com/umbralcalc/stochadex/pkg/simulator"
	"gonum.org/v1/gonum/mat"
)

// PoissonCovariateGLMLikelihood implements the stochadex
// LikelihoodDistributionWithGradient interface for a Poisson GLM with
// log link and time-varying covariates.
//
// Each data row passed to EvaluateLogLike / EvaluateLogLikeMeanGrad is
// expected to have layout [count_0, ..., count_{NRates-1}, cov_0, ..., cov_{NCovs-1}].
// The counts are the first NRates elements; covariates follow.
//
// The "mean" (β) is obtained via MeanFromParamsOrPartition (wired to
// the gradient_descent partition state). Rates are computed as:
//
//	μ_i = exp(β_{i,0} + Σⱼ β_{i,j+1} · cov_j)
//
// The gradient returned by EvaluateLogLikeMeanGrad is w.r.t. β (not μ).
type PoissonCovariateGLMLikelihood struct {
	Src          rand.Source
	NRates       int
	NCovs        int
	NBaselineOff int // number of baseline offset values appended after covariates (0 = no baseline)
	beta         *mat.VecDense
}

func (p *PoissonCovariateGLMLikelihood) SetSeed(
	partitionIndex int,
	settings *simulator.Settings,
) {
	p.Src = rand.NewPCG(
		settings.Iterations[partitionIndex].Seed,
		settings.Iterations[partitionIndex].Seed,
	)
}

func (p *PoissonCovariateGLMLikelihood) SetParams(
	params *simulator.Params,
	partitionIndex int,
	stateHistories []*simulator.StateHistory,
	timestepsHistory *simulator.CumulativeTimestepsHistory,
) {
	p.beta = inference.MeanFromParamsOrPartition(
		params, partitionIndex, stateHistories,
	)
}

func (p *PoissonCovariateGLMLikelihood) coeffsPerRate() int {
	return 1 + p.NCovs
}

func (p *PoissonCovariateGLMLikelihood) computeRate(
	rateIdx int, covariates, baseline []float64,
) float64 {
	cpr := p.coeffsPerRate()
	offset := rateIdx * cpr
	logRate := p.beta.AtVec(offset)
	for j := 0; j < p.NCovs; j++ {
		logRate += p.beta.AtVec(offset+1+j) * covariates[j]
	}
	rate := math.Exp(logRate)
	if len(baseline) > rateIdx && baseline[rateIdx] > 0 {
		rate *= baseline[rateIdx]
	}
	return rate
}

// extractDataParts splits a data row into counts, covariates, and baseline.
func (p *PoissonCovariateGLMLikelihood) extractDataParts(
	data []float64,
) (counts, covariates, baseline []float64) {
	counts = data[:p.NRates]
	covariates = data[p.NRates : p.NRates+p.NCovs]
	if p.NBaselineOff > 0 {
		baseline = data[p.NRates+p.NCovs : p.NRates+p.NCovs+p.NBaselineOff]
	}
	return
}

func (p *PoissonCovariateGLMLikelihood) EvaluateLogLike(
	data []float64,
) float64 {
	counts, covariates, baseline := p.extractDataParts(data)
	ll := 0.0
	for i := 0; i < p.NRates; i++ {
		mu := p.computeRate(i, covariates, baseline)
		if mu > 0 {
			ll += counts[i]*math.Log(mu) - mu
		}
	}
	return ll
}

func (p *PoissonCovariateGLMLikelihood) GenerateNewSamples() []float64 {
	// Generate Poisson samples from current rates (zero covariates, no baseline).
	totalWidth := p.NRates + p.NCovs + p.NBaselineOff
	samples := make([]float64, totalWidth)
	covs := make([]float64, p.NCovs)
	rng := rand.New(p.Src)
	for i := 0; i < p.NRates; i++ {
		mu := p.computeRate(i, covs, nil)
		// Simple Poisson sampling via inverse transform.
		L := math.Exp(-mu)
		k := 0
		prod := 1.0
		for {
			prod *= rng.Float64()
			if prod < L {
				break
			}
			k++
		}
		samples[i] = float64(k)
	}
	return samples
}

// EvaluateLogLikeMeanGrad returns the gradient of the Poisson
// log-likelihood w.r.t. the β coefficients.
// For rate i and coefficient j:
//
//	∂L/∂β_{i,0} = (y_i - μ_i)           (intercept, x_0 = 1)
//	∂L/∂β_{i,j} = (y_i - μ_i) · cov_j   (covariates)
func (p *PoissonCovariateGLMLikelihood) EvaluateLogLikeMeanGrad(
	data []float64,
) []float64 {
	counts, covariates, baseline := p.extractDataParts(data)
	cpr := p.coeffsPerRate()
	grad := make([]float64, p.NRates*cpr)
	for i := 0; i < p.NRates; i++ {
		mu := p.computeRate(i, covariates, baseline)
		residual := counts[i] - mu
		offset := i * cpr
		grad[offset] = residual // intercept (x_0 = 1)
		for j := 0; j < p.NCovs; j++ {
			grad[offset+1+j] = residual * covariates[j]
		}
	}
	return grad
}
