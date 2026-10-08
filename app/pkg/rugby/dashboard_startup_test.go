package rugby

import (
	"testing"

	"github.com/umbralcalc/dexetera/pkg/dashboard"
	"github.com/umbralcalc/dexetera/pkg/simio"
	"github.com/umbralcalc/stochadex/pkg/simulator"
)

// TestDashboardStarts makes the checks dexetera's RegisterStep makes when the
// page loads: every action partition exists and declares action_state_values,
// at the width its sliders send. Run natively, a misconfigured page fails
// here rather than in the reader's browser.
func TestDashboardStarts(t *testing.T) {
	cfg := NewConfig()
	settings, implementations := cfg.SimulationGenerator().GenerateConfigs()
	if err := dashboard.CheckActionWidths(cfg, settings); err != nil {
		t.Fatal(err)
	}
	implementations.OutputFunction = &simulator.NilOutputFunction{}
	coordinator := simulator.NewPartitionCoordinator(settings, implementations)
	if _, err := simio.NewActionDispatcher(coordinator, cfg.ActionStatePartitionNames); err != nil {
		t.Fatal(err)
	}
}
