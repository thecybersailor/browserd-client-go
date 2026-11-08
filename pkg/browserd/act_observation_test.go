package browserd

import (
	"encoding/json"
	"testing"
)

func TestActExecutesWithoutObservationBinding(t *testing.T) {
	for _, in := range []ActInput{{Action: "click", Ref: "e1"}, {Action: "click", X: 10, Y: 10}} {
		if err := ValidateActInput(in); err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err = json.Unmarshal(raw, &payload); err != nil {
			t.Fatal(err)
		}
		if _, exists := payload["expectedSnapshotId"]; exists {
			t.Fatalf("unexpected observation binding: %s", raw)
		}
	}
}
