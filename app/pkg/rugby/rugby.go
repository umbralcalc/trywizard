// Package rugby is the dexetera dashboard for the rugby manager
// decision-support post. The simulator under the hood is the fitted
// trywizard model; the controls are four sliders that set when each
// position group is substituted; the visualisation is a margin line for
// the in-progress match and a histogram of completed-match margins,
// plus DOM readouts for the live score and rolling win percentage.
//
// See app/cmd/rugby/{train,generate,register_step} for the trainer that
// produces fittedmodel.json, the codegen that emits the widget shell,
// and the wasm entry-point respectively.
package rugby

import (
	"fmt"

	"github.com/umbralcalc/dexetera/pkg/dashboard"
)

const (
	canvasWidth  = 640
	canvasHeight = 400

	// Top strip: in-progress match score margin (line chart).
	chartX      = 60
	chartY      = 30
	chartWidth  = HistCanvasWidth
	chartHeight = 80

	// Histogram baseline (where bars rest) and the tick mark row beneath it.
	axisY0 = HistCanvasY0 + HistCanvasHeight + 4
	tickY1 = axisY0 + 6
)

// NewConfig returns the dashboard.Config for the rugby widget. The
// declaration order of renderers matters: later renderers draw on top.
func NewConfig() *dashboard.Config {
	vb := dashboard.NewVisualizationBuilder().
		WithCanvas(canvasWidth, canvasHeight).
		WithBackground("#fafafa").
		WithUpdateInterval(0).
		// Zero line (score diff = 0) inside the line chart area.
		AddLine("", chartX, chartY+chartHeight/2, chartX+chartWidth, chartY+chartHeight/2, &dashboard.LineOptions{
			Color:       "#cccccc",
			Width:       1,
			DashPattern: []int{3, 3},
		}).
		// Score margin (home − away) line for the in-progress match.
		// Binds to match_state state[0], i.e. MSIdxScoreDiff.
		AddLineChart("match_state", chartX, chartY, chartWidth, chartHeight, &dashboard.ChartOptions{
			Color:     "#3c78d8",
			LineWidth: 2,
		}).
		// Section divider between the score-margin chart and the histogram.
		AddLine("", chartX, chartY+chartHeight+18, chartX+chartWidth, chartY+chartHeight+18, &dashboard.LineOptions{
			Color: "#e3e6ec",
			Width: 1,
		}).
		// Vertical guide at margin = 0 inside the histogram (helps the
		// reader read which side the bars fall on).
		AddLine("",
			HistCanvasX0+HistCanvasWidth/2, HistCanvasY0,
			HistCanvasX0+HistCanvasWidth/2, axisY0, &dashboard.LineOptions{
				Color:       "#cfd4dc",
				Width:       1,
				DashPattern: []int{2, 4},
			}).
		// Histogram bars driven by histogram_bars partition (rectangleSet).
		// Slate grey rather than blue, to read as "outcome distribution"
		// distinct from the line chart's "current trajectory".
		AddRectangleSet("histogram_bars", 0, 0, &dashboard.ShapeOptions{
			FillColor: "#7d8aa1",
			Anchor:    "topLeft",
		}).
		// Histogram baseline.
		AddLine("", HistCanvasX0, axisY0, HistCanvasX0+HistCanvasWidth, axisY0, &dashboard.LineOptions{
			Color: "#2c3e50",
			Width: 1,
		})

	// Axis tick marks at −50, −25, 0, +25, +50 (5 evenly spaced).
	for i := 0; i <= 4; i++ {
		x := HistCanvasX0 + i*HistCanvasWidth/4
		vb = vb.AddLine("", x, axisY0, x, tickY1, &dashboard.LineOptions{
			Color: "#2c3e50",
			Width: 1,
		})
	}

	vis := vb.Build()

	cfg := dashboard.NewConfigBuilder("rugby").
		WithDescription("Rugby manager decision support: pick when each home position group comes off the pitch; the simulator (fitted to thousands of real match events) shows how the win-margin distribution shifts. This is a research model fitted to match events, not a rugby manager decision tool.").
		WithServerPartition("match_state").
		WithServerPartition("outcomes").
		WithServerPartition("histogram_bars").
		WithActionStatePartition("home_sub_snapshot").
		WithVisualization(vis).
		WithSimulation(BuildRugbySimulation)

	for i, s := range []sliderSpec{
		{name: "front_row", label: "Front-row sub (min)"},
		{name: "back_row", label: "Back-row sub (min)"},
		{name: "halves", label: "Halves sub (min)"},
		{name: "outside_backs", label: "Outside-backs sub (min)"},
	} {
		cfg = cfg.WithSlider(dashboard.Slider{
			Name:       s.name,
			Label:      s.label,
			Partition:  "home_sub_snapshot",
			ValueIndex: i,
			Min:        0,
			Max:        80,
			Step:       1,
			Default:    DefaultHomeSubMinutes[i],
			Decimals:   0,
		})
	}

	cfg = cfg.
		WithReadout(dashboard.Readout{
			Partition: "match_state",
			Template: fmt.Sprintf(
				"match {v%d} · min {v%d} · home {v%d} · away {v%d}",
				MSIdxMatchID, MSIdxMatchMinute, MSIdxHomeScore, MSIdxAwayScore,
			),
			Decimals: 0,
		}).
		WithReadout(dashboard.Readout{
			Partition: "outcomes",
			Template: fmt.Sprintf(
				"home win: {v%d}%% · matches: {v%d}",
				OutIdxWinPct, OutIdxMatches,
			),
			Decimals: 1,
		}).
		WithResetButton().
		WithInlineDriver(30)

	return cfg.Build()
}

type sliderSpec struct {
	name, label string
}
