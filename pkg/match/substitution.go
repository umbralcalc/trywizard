package match

import (
	"github.com/umbralcalc/stochadex/pkg/general"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

// NewSubCovariatesReplayPartition creates a partition that replays
// observed substitution covariates from pre-computed data.
func NewSubCovariatesReplayPartition(covariates [][]float64) *simulator.PartitionConfig {
	return &simulator.PartitionConfig{
		Name: "sub_covariates",
		Iteration: &general.FromStorageIteration{
			Data: covariates,
		},
		Params:            simulator.NewParams(make(map[string][]float64)),
		InitStateValues:   make([]float64, SubCovWidth),
		StateHistoryDepth: 1,
		Seed:              0,
	}
}

// NewSubCovariatesConstantPartition creates a partition that returns
// fixed covariate values at every time step. Pass make([]float64, SubCovWidth)
// for a "no substitutions" baseline.
func NewSubCovariatesConstantPartition(values []float64) *simulator.PartitionConfig {
	return &simulator.PartitionConfig{
		Name: "sub_covariates",
		Iteration: &general.ParamValuesIteration{},
		Params: simulator.NewParams(map[string][]float64{
			"param_values": values,
		}),
		InitStateValues:   make([]float64, SubCovWidth),
		StateHistoryDepth: 1,
		Seed:              0,
	}
}
