package rugby

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed fittedmodel.json
var fittedModelBytes []byte

// FittedModel is the pre-trained model the widget runs against. Populated
// once at build time by `go run ./app/cmd/rugby/train`. The JSON is
// embedded into the wasm so the page boots without a network fetch.
type FittedModel struct {
	ScoreCoefficients []float64   `json:"scoreCoefficients"`
	CardCoefficients  []float64   `json:"cardCoefficients"`
	BaselineRates     [][]float64 `json:"baselineRates"`
	ConversionProbs   []float64   `json:"conversionProbs"`
}

// LoadFittedModel decodes the embedded fittedmodel.json. It panics on a
// malformed embedded payload, since that would mean an inconsistent build
// rather than a runtime condition we can recover from.
func LoadFittedModel() *FittedModel {
	var m FittedModel
	if err := json.Unmarshal(fittedModelBytes, &m); err != nil {
		panic(fmt.Sprintf("rugby: malformed embedded fittedmodel.json: %v", err))
	}
	return &m
}
