package browserd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const DefaultExecutionTimeoutMs = 300000
const MaxExecutionTimeoutMs = 1800000
const MaxWatchTimeoutMs = 30000

type TaskControl struct {
	RequestID          string `json:"request_id"`
	ExecutionTimeoutMs int    `json:"executionTimeoutMs,omitempty"`
}
type TaskSubmission struct {
	BrowserdTaskID string `json:"browserd_task_id"`
}
type TaskState string

const (
	TaskRunning    TaskState = "running"
	TaskCancelling TaskState = "cancelling"
	TaskSucceeded  TaskState = "succeeded"
	TaskFailed     TaskState = "failed"
	TaskCancelled  TaskState = "cancelled"
	TaskTimedOut   TaskState = "timed_out"
)

func (s TaskState) Terminal() bool {
	return s == TaskSucceeded || s == TaskFailed || s == TaskCancelled || s == TaskTimedOut
}
func (s TaskState) Valid() bool { return s == TaskRunning || s == TaskCancelling || s.Terminal() }

type TaskError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type TaskInfo struct {
	RuntimeSessionID string          `json:"runtimeSessionId"`
	Operation        string          `json:"operation"`
	Stage            string          `json:"stage"`
	ElapsedMs        int64           `json:"elapsedMs"`
	Result           json.RawMessage `json:"result,omitempty"`
	Error            *TaskError      `json:"error,omitempty"`
	SessionUsable    bool            `json:"sessionUsable"`
}
type Task struct {
	BrowserdTaskID string    `json:"browserd_task_id"`
	State          TaskState `json:"state"`
	Info           TaskInfo  `json:"info"`
}
type WatchTaskInput struct {
	TimeoutMs int `json:"timeoutMs"`
}

func ValidateTaskControl(in TaskControl) error {
	if strings.TrimSpace(in.RequestID) == "" || len(in.RequestID) > 256 {
		return Error{Code: "INVALID_REQUEST", Message: "request_id must contain 1..256 characters"}
	}
	if in.ExecutionTimeoutMs < 0 || in.ExecutionTimeoutMs > MaxExecutionTimeoutMs {
		return Error{Code: "INVALID_REQUEST", Message: fmt.Sprintf("executionTimeoutMs must be 0..%d (0 selects the default)", MaxExecutionTimeoutMs)}
	}
	return nil
}

func ValidateWaitForCondition(c WaitForCondition) error {
	switch c.Type {
	case "url_contains":
		if strings.TrimSpace(c.URLContains) != "" {
			return nil
		}
	case "url_matches":
		if strings.TrimSpace(c.URLMatches) != "" {
			return nil
		}
	case "text_visible", "text_gone":
		if strings.TrimSpace(c.Text) != "" {
			return nil
		}
	case "element_visible", "element_enabled", "element_editable", "ref_actionable":
		if strings.TrimSpace(c.Ref) != "" || strings.TrimSpace(c.Selector) != "" {
			return nil
		}
	}
	return Error{Code: "INVALID_REQUEST", Message: "wait_for requires a supported condition and its URL, text, ref or selector"}
}
func ValidateTaskID(id string) error {
	if !strings.HasPrefix(id, "bdt_") || len(id) > 256 || strings.ContainsAny(id, "/\\ ?#\t\r\n") {
		return Error{Code: "INVALID_REQUEST", Message: "browserd_task_id must be the exact ID returned by task submission"}
	}
	return nil
}
func (c *Client) submitTask(ctx context.Context, path string, in any) (TaskSubmission, error) {
	var out TaskSubmission
	if err := c.doJSONStatus(ctx, http.MethodPost, path, in, &out, http.StatusAccepted); err != nil {
		return out, err
	}
	if err := ValidateTaskID(out.BrowserdTaskID); err != nil {
		return TaskSubmission{}, Error{Code: "browserd_response_invalid", Message: "task submission did not return browserd_task_id", Cause: err}
	}
	return out, nil
}
func (c *Client) WatchTask(ctx context.Context, id string, timeoutMs int) (Task, error) {
	if err := ValidateTaskID(id); err != nil {
		return Task{}, err
	}
	if timeoutMs < 0 || timeoutMs > MaxWatchTimeoutMs {
		return Task{}, Error{Code: "INVALID_REQUEST", Message: "watch timeoutMs must be 0..30000"}
	}
	var out Task
	if err := c.doJSON(ctx, http.MethodPost, "/v1/tasks/"+url.PathEscape(id)+"/watch", WatchTaskInput{timeoutMs}, &out); err != nil {
		return out, err
	}
	return validateTaskResponse(id, out)
}
func (c *Client) CancelTask(ctx context.Context, id string) (Task, error) {
	if err := ValidateTaskID(id); err != nil {
		return Task{}, err
	}
	var out Task
	if err := c.doJSON(ctx, http.MethodPost, "/v1/tasks/"+url.PathEscape(id)+"/cancel", struct{}{}, &out); err != nil {
		return out, err
	}
	return validateTaskResponse(id, out)
}
func validateTaskResponse(id string, t Task) (Task, error) {
	if t.BrowserdTaskID != id || !t.State.Valid() {
		return Task{}, Error{Code: "browserd_response_invalid", Message: "task response has mismatched ID or invalid state"}
	}
	if t.State == TaskSucceeded && (len(t.Info.Result) == 0 || string(t.Info.Result) == "null") {
		return Task{}, Error{Code: "browserd_response_invalid", Message: "successful task has no result"}
	}
	if (t.State == TaskFailed || t.State == TaskTimedOut) && t.Info.Error == nil {
		return Task{}, Error{Code: "browserd_response_invalid", Message: "failed task has no error"}
	}
	return t, nil
}
func DecodeTaskResult[T any](task Task) (T, error) {
	var out T
	if task.State != TaskSucceeded {
		return out, Error{Code: "TASK_NOT_SUCCEEDED", Message: "task has not succeeded"}
	}
	if err := json.Unmarshal(task.Info.Result, &out); err != nil {
		return out, Error{Code: "browserd_response_invalid", Message: "invalid task result", Cause: err}
	}
	return out, nil
}

func DecodeNavigationResult(task Task, includeSnapshot bool) (NavigateResult, error) {
	out, err := DecodeTaskResult[NavigateResult](task)
	if err != nil {
		if includeSnapshot {
			return NavigateResult{}, Error{Code: "BROWSERD_SNAPSHOT_REQUIRED", Message: "invalid navigation snapshot result", Cause: err}
		}
		return NavigateResult{}, err
	}
	if includeSnapshot {
		err = ValidateSnapshotResult(out.Snapshot)
		if err == nil && out.URL != out.Snapshot.Page["url"] {
			err = fmt.Errorf("navigation URL does not match snapshot URL")
		}
		if err != nil {
			return NavigateResult{}, Error{Code: "BROWSERD_SNAPSHOT_REQUIRED", Message: "navigation did not return the requested valid snapshot", Cause: err}
		}
	}
	return out, nil
}
func (c *Client) WaitForLoadState(ctx context.Context, sessionID string, in PageToolInput) (TaskSubmission, error) {
	if in.Method != "page.waitForLoadState" || in.TimeoutMs != 0 {
		return TaskSubmission{}, Error{Code: "INVALID_REQUEST", Message: "use page.waitForLoadState with executionTimeoutMs"}
	}
	if err := ValidateTaskControl(in.TaskControl); err != nil {
		return TaskSubmission{}, err
	}
	return c.submitTask(ctx, sessionPath(sessionID, "pageTool"), in)
}
