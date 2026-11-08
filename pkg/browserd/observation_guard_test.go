package browserd

import (
	"encoding/json"
	"testing"
)

func TestActObservationGuardIsExplicitAndRejectsCoordinates(t *testing.T) {
	in := ActInput{Action: "click", Ref: "e1", ExpectedSnapshotID: "snap_observed"}
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
	if payload["expectedSnapshotId"] != "snap_observed" {
		t.Fatalf("missing observation binding: %s", raw)
	}
	in.Ref = ""
	in.X = 10
	in.Y = 10
	if err = ValidateActInput(in); err == nil {
		t.Fatal("guarded actions must not use a coordinate target")
	}
}
