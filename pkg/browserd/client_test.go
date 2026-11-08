package browserd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestClientEndpointsCallCurrentBrowserdPaths(t *testing.T) {
	seen := []string{}
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.RequestURI())
		gotAuth = r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/v1/sessions":
			writeJSON(t, w, map[string]any{"runtimeSessionId": "rt_1", "resolvedVersion": "v1"})
		case "/v1/sessions/rt_1/navigate":
			w.WriteHeader(http.StatusAccepted)
			writeJSON(t, w, map[string]any{"browserd_task_id": "bdt_test"})
		case "/v1/sessions/rt_1/snapshot":
			writeJSON(t, w, map[string]any{"snapshotId": "snap_1", "page": map[string]any{"url": "https://example.com", "groups": map[string]any{}}})
		case "/v1/sessions/rt_1/act":
			writeJSON(t, w, map[string]any{"ok": true, "action": "click", "ref": "e1"})
		case "/v1/sessions/rt_1/wait-for":
			w.WriteHeader(http.StatusAccepted)
			writeJSON(t, w, map[string]any{"browserd_task_id": "bdt_test"})
		case "/v1/sessions/rt_1/pageTool":
			writeJSON(t, w, map[string]any{"value": "Example"})
		case "/v1/sessions/rt_1/evaluate":
			w.WriteHeader(http.StatusAccepted)
			writeJSON(t, w, map[string]any{"browserd_task_id": "bdt_test"})
		case "/v1/sessions/rt_1/screenshot":
			writeJSON(t, w, map[string]any{"screenshotId": "shot_1.png", "s3Path": "/screens/shot_1.png", "contentType": "image/png", "byteLength": 123})
		case "/v1/sessions/rt_1/upload-files":
			writeJSON(t, w, map[string]any{"ok": true, "ref": "e1", "fileNames": []string{"a.txt"}})
		case "/v1/sessions/rt_1/live-view":
			writeJSON(t, w, map[string]any{"handoffId": "live_1", "viewerUrl": "https://browser.example/v/live_1", "permission": "view"})
		case "/v1/sessions/rt_1/handoff/start":
			writeJSON(t, w, map[string]any{"handoffId": "ho_1", "viewerUrl": "https://browser.example/v/ho_1", "permission": "control"})
		case "/v1/sessions/rt_1/handoff/ho_1/complete":
			writeJSON(t, w, map[string]any{"ok": true})
		case "/v1/sessions/rt_1/commit":
			writeJSON(t, w, map[string]any{"newVersion": "v2", "bytes": 10, "durationMs": 20})
		case "/v1/sessions/rt_1":
			writeJSON(t, w, map[string]any{"ok": true})
		default:
			t.Fatalf("unexpected path %s", r.URL.RequestURI())
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, APIKey: "secret"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx := context.Background()
	if _, err := client.CreateSession(ctx, CreateSessionInput{Fingerprint: FingerprintConfig{Seed: "seed"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Navigate(ctx, "rt_1", NavigateInput{TaskControl: TaskControl{RequestID: "test-request"}, URL: "https://example.com"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Snapshot(ctx, "rt_1", SnapshotInput{Mode: "refs"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Act(ctx, "rt_1", ActInput{Action: "click", Ref: "e1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WaitFor(ctx, "rt_1", WaitForInput{TaskControl: TaskControl{RequestID: "test-request"}, Condition: WaitForCondition{Type: "ref_actionable", Ref: "e1"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.PageTool(ctx, "rt_1", PageToolInput{Method: "page.title"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Evaluate(ctx, "rt_1", EvaluateInput{TaskControl: TaskControl{RequestID: "test-request"}, Script: "return true"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Screenshot(ctx, "rt_1", ScreenshotInput{Mode: "viewport", ScreenshotS3Prefix: "/screens/"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.UploadFiles(ctx, "rt_1", UploadFilesInput{Ref: "e1", Files: []UploadFileSource{{LocalPath: "/tmp/a.txt"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.LiveView(ctx, "rt_1", LiveViewInput{Permission: "view"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.StartHandoff(ctx, "rt_1", StartHandoffInput{Permission: "control"}); err != nil {
		t.Fatal(err)
	}
	if err := client.CompleteHandoff(ctx, "rt_1", "ho_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Commit(ctx, "rt_1", CommitInput{}); err != nil {
		t.Fatal(err)
	}
	if err := client.Close(ctx, "rt_1"); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"POST /v1/sessions",
		"POST /v1/sessions/rt_1/navigate",
		"GET /v1/sessions/rt_1/snapshot?mode=refs",
		"POST /v1/sessions/rt_1/act",
		"POST /v1/sessions/rt_1/wait-for",
		"POST /v1/sessions/rt_1/pageTool",
		"POST /v1/sessions/rt_1/evaluate",
		"POST /v1/sessions/rt_1/screenshot",
		"POST /v1/sessions/rt_1/upload-files",
		"POST /v1/sessions/rt_1/live-view",
		"POST /v1/sessions/rt_1/handoff/start",
		"POST /v1/sessions/rt_1/handoff/ho_1/complete",
		"POST /v1/sessions/rt_1/commit",
		"DELETE /v1/sessions/rt_1",
	}
	if !reflect.DeepEqual(seen, want) {
		t.Fatalf("seen paths = %#v", seen)
	}
	if gotAuth != "Bearer secret" {
		t.Fatalf("auth = %q", gotAuth)
	}
}

func TestDoJSONReturnsTypedBrowserdError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": nil, "error": map[string]any{"code": "INVALID_REQUEST", "message": "bad input"}})
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	var out map[string]any
	err = client.doJSON(context.Background(), http.MethodPost, "/v1/test", map[string]any{}, &out)
	var browserErr Error
	if !AsError(err, &browserErr) || browserErr.Code != "INVALID_REQUEST" || browserErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("err = %#v", err)
	}
}

func TestValidateActInputCurrentActions(t *testing.T) {
	valid := []ActInput{
		{Action: "click", Ref: "e1"},
		{Action: "click", X: 10, Y: 10},
		{Action: "doubleClick", Ref: "e1"},
		{Action: "hover", Ref: "e1"},
		{Action: "scroll", DeltaY: 100},
		{Action: "paste", Ref: "e1", Text: "hello"},
		{Action: "type", Ref: "e1", Text: "hello", Submit: true},
		{Action: "fill", Ref: "e1", Value: "hello"},
		{Action: "press", Ref: "e1", Key: "Enter"},
		{Action: "scrollIntoView", Ref: "t1"},
		{Action: "select", Ref: "e1", Values: []string{"a"}},
	}
	for _, input := range valid {
		if err := ValidateActInput(input); err != nil {
			t.Fatalf("%+v should be valid: %v", input, err)
		}
	}
	invalid := []ActInput{
		{Action: ""},
		{Action: "type", Text: "hello"},
		{Action: "press", Ref: "e1", Key: "Return"},
		{Action: "press", Ref: "e1", Key: "Control+s"},
		{Action: "scroll", DeltaX: 0, DeltaY: 0},
		{Action: "paste"},
	}
	for _, input := range invalid {
		if err := ValidateActInput(input); err == nil {
			t.Fatalf("%+v should be invalid", input)
		}
	}
}

func TestScreenshotContractRejectsLegacyShape(t *testing.T) {
	if err := ValidateScreenshotInput(ScreenshotInput{Mode: "selector"}); err == nil {
		t.Fatalf("selector mode must require selector")
	}
	if err := ValidateScreenshotInput(ScreenshotInput{Mode: "viewport", Selector: ".card"}); err == nil {
		t.Fatalf("viewport mode must reject selector")
	}
	if err := ValidateScreenshotInput(ScreenshotInput{Mode: "fullpage", ScreenshotS3Prefix: "/screens/"}); err != nil {
		t.Fatalf("fullpage should be valid: %v", err)
	}
	inputType := reflect.TypeOf(ScreenshotInput{})
	for _, field := range []string{"Ref", "FullPage", "Base64"} {
		if _, ok := inputType.FieldByName(field); ok {
			t.Fatalf("ScreenshotInput must not expose legacy field %s", field)
		}
	}
	resultType := reflect.TypeOf(ScreenshotResult{})
	if _, ok := resultType.FieldByName("Base64"); ok {
		t.Fatalf("ScreenshotResult must not expose base64")
	}
}

func TestActInputMarshalsCurrentBrowserdFields(t *testing.T) {
	raw, err := json.Marshal(ActInput{
		Action:        "paste",
		Ref:           "e1",
		X:             10,
		Y:             20,
		DeltaX:        1,
		DeltaY:        2,
		Text:          "plain",
		HTML:          "<b>plain</b>",
		Key:           "Enter",
		Value:         "v",
		Values:        []string{"a", "b"},
		Clear:         true,
		Submit:        true,
		Button:        "left",
		ClickCount:    2,
		MotionProfile: "humanized",
		TimeoutMs:     3000,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"action":"paste"`, `"deltaX":1`, `"html":"\u003cb\u003eplain\u003c/b\u003e"`, `"submit":true`, `"motionProfile":"humanized"`} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("missing %s in %s", want, raw)
		}
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"data": data, "error": nil}); err != nil {
		t.Fatalf("write json: %v", err)
	}
}
