package match

import (
	"testing"

	"gonum.org/v1/gonum/floats/scalar"
)

func TestTransformEventsToStateTimeStorage(t *testing.T) {
	// Game 600009: home team is 25900 (Gloucester Rugby)
	storage, err := TransformEventsToStateTimeStorage("../../dat/events.csv", 600009, 25900)
	if err != nil {
		t.Fatalf("failed to transform events: %v", err)
	}

	events := storage.GetValues("events")
	times := storage.GetTimes()

	if len(events) == 0 {
		t.Fatal("expected non-empty events storage")
	}
	if len(events) != len(times) {
		t.Fatalf("events length %d != times length %d", len(events), len(times))
	}

	// Each row should have EventWidth columns.
	for i, row := range events {
		if len(row) != EventWidth {
			t.Errorf("minute %d: expected width %d, got %d", i, EventWidth, len(row))
		}
	}

	// Verify specific known events from the CSV:
	// Minute 4: home try (Jack Cotgreave, team 25900)
	t.Run("minute 4 home try", func(t *testing.T) {
		if !scalar.EqualWithinAbsOrRel(events[4][IdxHomeTry], 1.0, 1e-10, 1e-10) {
			t.Errorf("expected 1 home try at minute 4, got %f", events[4][IdxHomeTry])
		}
	})

	// Minute 8: away try (Tom Pearson, team 25907)
	t.Run("minute 8 away try", func(t *testing.T) {
		if !scalar.EqualWithinAbsOrRel(events[8][IdxAwayTry], 1.0, 1e-10, 1e-10) {
			t.Errorf("expected 1 away try at minute 8, got %f", events[8][IdxAwayTry])
		}
	})

	// Minute 5: home conversion (Santiago Carreras, team 25900)
	t.Run("minute 5 home conversion", func(t *testing.T) {
		if !scalar.EqualWithinAbsOrRel(events[5][IdxHomeConv], 1.0, 1e-10, 1e-10) {
			t.Errorf("expected 1 home conversion at minute 5, got %f", events[5][IdxHomeConv])
		}
	})

	// Minute 38: home yellow card (Afolabi Fasogbon, team 25900)
	t.Run("minute 38 home yellow card", func(t *testing.T) {
		if !scalar.EqualWithinAbsOrRel(events[38][IdxHomeYellow], 1.0, 1e-10, 1e-10) {
			t.Errorf("expected 1 home yellow at minute 38, got %f", events[38][IdxHomeYellow])
		}
	})

	// Game 600009 has no penalty goals, so penalty columns should be zero.
	t.Run("no penalties in game 600009", func(t *testing.T) {
		totalHomePenalties := 0.0
		totalAwayPenalties := 0.0
		for _, row := range events {
			totalHomePenalties += row[IdxHomePenalty]
			totalAwayPenalties += row[IdxAwayPenalty]
		}
		if totalHomePenalties != 0 {
			t.Errorf("expected 0 home penalties, got %f", totalHomePenalties)
		}
		if totalAwayPenalties != 0 {
			t.Errorf("expected 0 away penalties, got %f", totalAwayPenalties)
		}
	})

	// Verify total counts across all minutes match match summary.
	t.Run("total event counts", func(t *testing.T) {
		totalHomeTries := 0.0
		totalAwayTries := 0.0
		totalHomeConv := 0.0
		totalHomeYellow := 0.0
		for _, row := range events {
			totalHomeTries += row[IdxHomeTry]
			totalAwayTries += row[IdxAwayTry]
			totalHomeConv += row[IdxHomeConv]
			totalHomeYellow += row[IdxHomeYellow]
		}
		if totalHomeTries < 1 {
			t.Errorf("expected at least 1 home try, got %f", totalHomeTries)
		}
		if totalAwayTries < 1 {
			t.Errorf("expected at least 1 away try, got %f", totalAwayTries)
		}
		if totalHomeConv < 1 {
			t.Errorf("expected at least 1 home conversion, got %f", totalHomeConv)
		}
		if totalHomeYellow < 1 {
			t.Errorf("expected at least 1 home yellow, got %f", totalHomeYellow)
		}
	})
}

func TestComputeConversionProbabilities(t *testing.T) {
	t.Run("test conversion probabilities for known match", func(t *testing.T) {
		storage, err := TransformEventsToStateTimeStorage(
			"../../dat/events.csv", 600009, 25900,
		)
		if err != nil {
			t.Fatalf("failed to load data: %v", err)
		}

		probs := ComputeConversionProbabilities(storage)
		if len(probs) != 2 {
			t.Fatalf("expected 2 probabilities, got %d", len(probs))
		}

		// Probabilities should be between 0 and 1.
		for i, p := range probs {
			if p < 0 || p > 1 {
				t.Errorf("prob[%d] = %f, expected in [0, 1]", i, p)
			}
		}

		t.Logf("conversion probabilities: home=%f, away=%f", probs[0], probs[1])
	})
}
