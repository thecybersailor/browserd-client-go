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

func TestLinkImageColumnsSurviveNavigateAndSnapshot(t *testing.T) {
	imageURL := "https://cdn.example/product.webp?token=" + strings.Repeat("x", 300)
	groups := map[string]any{"links": map[string]any{
		"columns": []any{"ref", "tag", "text", "href", "image_url"},
		"rows":    []any{[]any{"e1", "A", "", "/item?variant=1", imageURL}, []any{"e2", "A", "Item $25", "/item?variant=1", ""}},
	}}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		snapshot := map[string]any{"snapshotId": "s1", "page": map[string]any{"url": "https://example.com/catalog", "groups": groups}}
		var data any
		switch r.Method + " " + r.URL.Path {
		case "POST /v1/sessions/rt_1/navigate":
			w.WriteHeader(http.StatusAccepted)
			data = map[string]any{"browserd_task_id": "bdt_test"}
		case "POST /v1/tasks/bdt_test/watch":
			data = map[string]any{"browserd_task_id": "bdt_test", "state": "succeeded", "info": map[string]any{"operation": "navigate", "result": map[string]any{"url": "https://example.com/catalog", "snapshot": snapshot}}}
		case "GET /v1/sessions/rt_1/snapshot":
			data = snapshot
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data, "error": nil})
	}))
	defer server.Close()
	c, err := NewClient(Config{BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	submission, err := c.Navigate(context.Background(), "rt_1", NavigateInput{TaskControl: TaskControl{RequestID: "r"}, URL: "https://example.com/catalog", IncludeSnapshot: true})
	if err != nil {
		t.Fatal(err)
	}
	task, err := c.WatchTask(context.Background(), submission.BrowserdTaskID, 0)
	if err != nil {
		t.Fatal(err)
	}
	nav, err := DecodeNavigationResult(task, true)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := c.Snapshot(context.Background(), "rt_1", SnapshotInput{})
	if err != nil {
		t.Fatal(err)
	}
	for _, page := range []PageSnapshot{nav.Snapshot.Page, snapshot.Page} {
		if !reflect.DeepEqual(page["groups"], groups) {
			t.Fatalf("image-only links, full image URLs or column identity lost: %+v", page)
		}
	}
	if calls != 3 {
		t.Fatalf("unexpected retry or fallback: %d calls", calls)
	}
}
