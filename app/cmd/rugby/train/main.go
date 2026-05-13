// train fits the rugby simulation's coefficients, baseline rates, and
// average conversion probabilities, then writes them to
// app/pkg/rugby/fittedmodel.json for the wasm widget to embed.
//
// Run once after data changes:
//
//	go run ./app/cmd/rugby/train
//
// The trainer is offline because RunMultiGameBaselineCovariateTrainingFull
// takes minutes; doing it in-browser at widget load is not viable.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/umbralcalc/trywizard/pkg/match"
)

type fittedModel struct {
	ScoreCoefficients []float64   `json:"scoreCoefficients"`
	CardCoefficients  []float64   `json:"cardCoefficients"`
	BaselineRates     [][]float64 `json:"baselineRates"`
	ConversionProbs   []float64   `json:"conversionProbs"`
}

func main() {
	eventsPath := flag.String("events", "dat/events.csv", "path to events.csv")
	playersPath := flag.String("players", "dat/players.csv", "path to players.csv")
	outPath := flag.String("out", "app/pkg/rugby/fittedmodel.json", "output JSON path")
	learningRate := flag.Float64("lr", 0.001, "training learning rate")
	windowDepth := flag.Int("window", 50, "training window depth")
	descentIters := flag.Int("inner", 1, "gradient descent inner iterations per step")
	epochs := flag.Int("epochs", 50, "training epochs")
	flag.Parse()

	fmt.Printf("training coefficients (lr=%g, window=%d, inner=%d, epochs=%d)...\n",
		*learningRate, *windowDepth, *descentIters, *epochs)
	coeffs, err := match.RunMultiGameBaselineCovariateTrainingFull(
		*eventsPath, *playersPath,
		*learningRate, *windowDepth, *descentIters, *epochs,
	)
	if err != nil {
		fatal("training failed: %v", err)
	}
	scoreCoeffs, cardCoeffs := match.SplitCoefficients(coeffs)

	fmt.Println("computing smoothed baseline rates...")
	baselineRates, err := match.ComputeSmoothedBaselineRates(*eventsPath, *playersPath)
	if err != nil {
		fatal("baseline rates: %v", err)
	}

	fmt.Println("averaging per-game conversion probabilities...")
	games, err := match.ListGames(*playersPath)
	if err != nil {
		fatal("list games: %v", err)
	}
	var sumHome, sumAway float64
	nGames := 0
	for _, g := range games {
		s, err := match.TransformEventsToStateTimeStorage(*eventsPath, g.GameID, g.HomeTeamID)
		if err != nil {
			continue
		}
		cp := match.ComputeConversionProbabilities(s)
		sumHome += cp[0]
		sumAway += cp[1]
		nGames++
	}
	if nGames == 0 {
		fatal("no games available for conversion probability averaging")
	}
	convProbs := []float64{sumHome / float64(nGames), sumAway / float64(nGames)}

	out := fittedModel{
		ScoreCoefficients: scoreCoeffs,
		CardCoefficients:  cardCoeffs,
		BaselineRates:     baselineRates,
		ConversionProbs:   convProbs,
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0755); err != nil {
		fatal("mkdir output dir: %v", err)
	}
	f, err := os.Create(*outPath)
	if err != nil {
		fatal("create %s: %v", *outPath, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&out); err != nil {
		fatal("encode JSON: %v", err)
	}
	fmt.Printf("wrote %s (%d baseline minutes, %d games for conv probs)\n",
		*outPath, len(baselineRates), nGames)
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
