package browserd

import (
	"encoding/json"
	"testing"
)

func TestNavigateIncludeSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name, response     string
		include, wantError bool
	}{
		{"valid", `{"url":"https://example.com/final","snapshot":{"snapshotId":"snap_1","page":{"url":"https://example.com/final","groups":{}}}}`, true, false},
		{"legacy opt out", `{"url":"https://example.com/","snapshotCleared":true}`, false, false},
		{"malformed snapshot", `{"url":"https://example.com/","snapshot":[]}`, true, true},
		{"numeric id", `{"url":"https://example.com/","snapshot":{"snapshotId":7,"page":{"url":"https://example.com/","groups":{}}}}`, true, true},
		{"missing", `{"url":"https://example.com/","snapshotCleared":true}`, true, true},
		{"null", `{"url":"https://example.com/","snapshot":null}`, true, true},
		{"empty id", `{"url":"https://example.com/","snapshot":{"page":{"url":"https://example.com/","groups":{}}}}`, true, true},
		{"missing url", `{"url":"https://example.com/","snapshot":{"snapshotId":"s","page":{"groups":{}}}}`, true, true},
		{"invalid url", `{"url":"about:blank","snapshot":{"snapshotId":"s","page":{"url":"about:blank","groups":{}}}}`, true, true},
		{"null groups", `{"url":"https://example.com/","snapshot":{"snapshotId":"s","page":{"url":"https://example.com/","groups":null}}}`, true, true},
		{"array groups", `{"url":"https://example.com/","snapshot":{"snapshotId":"s","page":{"url":"https://example.com/","groups":[]}}}`, true, true},
		{"mismatched url", `{"url":"https://example.com/wrong","snapshot":{"snapshotId":"s","page":{"url":"https://example.com/","groups":{}}}}`, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			task := Task{BrowserdTaskID: "bdt_test", State: TaskSucceeded, Info: TaskInfo{Operation: "navigate", Result: json.RawMessage(tc.response)}}
			result, err := DecodeNavigationResult(task, tc.include)
			if tc.wantError {
				if result.Snapshot != nil || result.URL != "" {
					t.Errorf("failed navigation exposed partial result: %+v", result)
				}
				var e Error
				if !AsError(err, &e) || e.Code != "BROWSERD_SNAPSHOT_REQUIRED" {
					t.Fatalf("expected protocol error, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			} else if tc.include && result.Snapshot.SnapshotID != "snap_1" {
				t.Fatal("snapshot lost")
			}
		})
	}
}
