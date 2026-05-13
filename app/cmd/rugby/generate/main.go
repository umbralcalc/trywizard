// generate emits the rugby widget shell (widget.html, test.html, build.sh)
// into app/rugby/. Re-run whenever the dashboard.Config in pkg/rugby
// changes shape (controls, partitions, visualisation).
//
//	cd app && go run ./cmd/rugby/generate
//
// After codegen, two CSS rules in the emitted HTML are rewritten so the
// live controls (slider accent + readout text) use the explainer-series'
// magenta-for-actions colour instead of dexetera's default blue.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/umbralcalc/dexetera/pkg/dashboard"
	"github.com/umbralcalc/trywizard/app/pkg/rugby"
)

// actionColor is the magenta from the Acting on Simulated Systems collection
// — used to signal "this is what the reader controls". Replaces dexetera's
// default blue (#3c78d8) on the slider track and the slider readout text.
const actionColor = "#b0447a"

func main() {
	runtimeURL := flag.String("runtime-url", "",
		"absolute URL the blog will serve dexetera's runtime/ folder from "+
			"(e.g. https://example.com/assets/dexetera/runtime/). "+
			"Leave empty for local preview via test.html.")
	wasmURL := flag.String("wasm-url", "",
		"absolute URL the blog will serve main.wasm from. "+
			"Leave empty for local preview.")
	flag.Parse()

	cfg := rugby.NewConfig()
	dashboard.MustGenerateWidget(cfg, dashboard.WidgetOptions{
		RuntimeBaseURL: *runtimeURL,
		WasmURL:        *wasmURL,
	})

	for _, name := range []string{"widget.html", "test.html"} {
		path := filepath.Join(cfg.Name, name)
		if err := recolorControls(path); err != nil {
			fmt.Fprintf(os.Stderr, "recolor %s: %v\n", path, err)
			os.Exit(1)
		}
	}
}

// recolorControls rewrites the two CSS rules dexetera emits for the live
// controls so the slider track and readout pick up the action colour
// rather than the default simulation-output blue. It also injects DOM
// captions around the canvas so the reader can see at a glance what the
// top strip vs the histogram represent — dexetera's text renderer
// hardcodes a white fill that is invisible on our light canvas, so canvas
// labels aren't an option.
//
// All replacements are anchored on enough surrounding context to avoid
// touching unrelated occurrences of the same colour or tag.
func recolorControls(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := string(data)
	replacements := [][2]string{
		{"accent-color: #3c78d8", "accent-color: " + actionColor},
		{".slider-readout { grid-area: readout; text-align: right; color: #3c78d8;",
			".slider-readout { grid-area: readout; text-align: right; color: " + actionColor + ";"},
		// Inject canvas captions: a small label above the canvas
		// describing what the top strip is, and one below describing the
		// histogram axis.
		{`<canvas width="640" height="400"></canvas>`,
			`<p class="canvas-caption canvas-caption-top">Score margin (home − away) for the in-progress match</p>` +
				`<canvas width="640" height="400"></canvas>` +
				`<p class="canvas-caption canvas-caption-axis"><span>−50</span><span>−25</span><span>0</span><span>+25</span><span>+50</span></p>` +
				`<p class="canvas-caption">Win-margin distribution across completed matches</p>`},
		// Small additional CSS for the injected captions. Inserted just
		// before the closing </style> of the widget's scoped stylesheet.
		{`</style>`,
			`#{{.WidgetID}} .canvas-caption { margin: 0; font-size: 0.85rem; color: #2c3e50; opacity: 0.75; text-align: center; }` +
				`#{{.WidgetID}} .canvas-caption-top { margin-bottom: 0.1em; }` +
				`#{{.WidgetID}} .canvas-caption-axis { display: flex; justify-content: space-between; padding: 0 60px; margin: 0.1em 0 0.3em; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; opacity: 0.6; }` +
				`</style>`},
	}
	// The last replacement uses a templated id; resolve it to the widget's
	// actual id by reading it from the existing DOM.
	widgetID := extractWidgetID(out)
	for i, r := range replacements {
		replacements[i][0] = strings.ReplaceAll(r[0], "{{.WidgetID}}", widgetID)
		replacements[i][1] = strings.ReplaceAll(r[1], "{{.WidgetID}}", widgetID)
	}
	for _, r := range replacements {
		if !strings.Contains(out, r[0]) {
			return fmt.Errorf("expected fragment not found: %q", r[0])
		}
		out = strings.Replace(out, r[0], r[1], 1)
	}
	return os.WriteFile(path, []byte(out), 0644)
}

// extractWidgetID picks the widget root's id out of the generated HTML so
// the captions we inject can scope to the same element as the rest of the
// dexetera CSS. The id appears verbatim as `id="..."` on the outer div.
func extractWidgetID(html string) string {
	const marker = `id="`
	i := strings.Index(html, marker)
	if i < 0 {
		return "dexetera"
	}
	i += len(marker)
	end := strings.Index(html[i:], `"`)
	if end < 0 {
		return "dexetera"
	}
	return html[i : i+end]
}
