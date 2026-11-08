# browserd-client-go

Go module: `github.com/thecybersailor/browserd-client-go`.

Import `github.com/thecybersailor/browserd-client-go/pkg/browserd`. The old module path is not supported.

`Navigate`, `Evaluate`, `WaitFor`, and `WaitForLoadState` submit asynchronous work and return `TaskSubmission{BrowserdTaskID}`. Supply a stable `TaskControl.RequestID`; `ExecutionTimeoutMs` defaults to 300000 and is bounded by 1800000. These methods never return the former synchronous action result.

Use `WatchTask(ctx, id, timeoutMs)` with a 0..30000 ms observation window. A watch timeout returns the current running state without cancelling execution. `CancelTask` explicitly requests cancellation. Only `TaskSucceeded` contains a successful immutable result; use `DecodeTaskResult[T]` or `DecodeNavigationResult`. Cancellation may invalidate the browser session; inspect `Info.SessionUsable`.

`CreateSession`, `Snapshot`, `Act`, `Screenshot`, `UploadFiles`, handoff, commit, cookies and profile transfer are synchronous controlled operations. `Act` rejects `waitFor`; use `WaitFor` with a condition. Short `PageTool` calls use `PageTool`; load-state waiting uses `WaitForLoadState` and the task lifecycle.

The client rejects missing task IDs, unknown states, legacy synchronous submissions and invalid time budgets. It preserves typed upstream error codes and transport causes. It does not automatically retry, guess alternate fields, or fall back to another protocol.

Validation:

```sh
go test ./...
go vet ./...
```

For decisions based on a captured page, set `ActInput.ExpectedSnapshotID` to the
`SnapshotResult.SnapshotID` from that observation. Guarded actions use an observed
`Ref` (or a scroll delta); coordinates are rejected. The server checks document,
node identity, relevant page state and target actionability before input.
`STALE_OBSERVATION` and `TARGET_NOT_ACTIONABLE` mean the input was rejected before
execution: take a new observation and decide again. Other failures can have an
unknown execution result and must not be automatically replayed.
