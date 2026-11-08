package browserd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskClientSubmissionAndWatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/sessions/rt_test/navigate":
			var in map[string]any
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Fatal(err)
			}
			if in["request_id"] != "request-test" || in["executionTimeoutMs"] != float64(60000) {
				t.Errorf("wrong submission: %+v", in)
			}
			if _, ok := in["timeoutMs"]; ok {
				t.Error("old timeoutMs sent")
			}
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"data":{"browserd_task_id":"bdt_test"},"error":null}`))
		case "/v1/tasks/bdt_test/watch":
			w.Write([]byte(`{"data":{"browserd_task_id":"bdt_test","state":"running","info":{"runtimeSessionId":"rt_test","operation":"navigate","stage":"loading"}},"error":null}`))
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	c, err := NewClient(Config{BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	submitted, err := c.Navigate(context.Background(), "rt_test", NavigateInput{URL: "https://example.com", TaskControl: TaskControl{RequestID: "request-test", ExecutionTimeoutMs: 60000}})
	if err != nil || submitted.BrowserdTaskID != "bdt_test" {
		t.Fatalf("submission: %+v %v", submitted, err)
	}
	task, err := c.WatchTask(context.Background(), submitted.BrowserdTaskID, 0)
	if err != nil || task.State != TaskRunning || task.Info.Stage != "loading" {
		t.Fatalf("watch: %+v %v", task, err)
	}
}

func TestTaskClientRejectsLegacySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"url":"https://example.com"},"error":null}`))
	}))
	defer server.Close()
	c, _ := NewClient(Config{BaseURL: server.URL})
	if _, err := c.Navigate(context.Background(), "rt", NavigateInput{URL: "https://example.com", TaskControl: TaskControl{RequestID: "r"}}); err == nil {
		t.Fatal("accepted legacy synchronous response")
	}
}

func TestTaskClientInvalidInputDoesNotSend(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++ }))
	defer server.Close()
	c, _ := NewClient(Config{BaseURL: server.URL})
	if _, err := c.WatchTask(context.Background(), "bdt_test", 30001); err == nil {
		t.Error("accepted excessive watch")
	}
	if _, err := c.WatchTask(context.Background(), "", 0); err == nil {
		t.Error("accepted empty task")
	}
	if _, err := c.Navigate(context.Background(), "rt", NavigateInput{URL: "https://example.com"}); err == nil {
		t.Error("accepted missing request ID")
	}
	if _, err := c.Act(context.Background(), "rt", ActInput{Action: "waitFor", Ref: "e1"}); err == nil {
		t.Error("accepted removed waitFor action")
	}
	if calls != 0 {
		t.Fatalf("sent %d invalid requests", calls)
	}
}

func TestTaskClientRejectsUnknownState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"browserd_task_id":"bdt_test","state":"maybe","info":{}},"error":null}`))
	}))
	defer server.Close()
	c, _ := NewClient(Config{BaseURL: server.URL})
	if _, err := c.WatchTask(context.Background(), "bdt_test", 0); err == nil {
		t.Fatal("accepted unknown task state")
	}
}
