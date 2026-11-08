package browserd

import "time"

type FingerprintConfig struct {
	Seed                string   `json:"seed"`
	Locale              string   `json:"locale"`
	Languages           []string `json:"languages"`
	AcceptLanguage      string   `json:"acceptLanguage"`
	Timezone            string   `json:"timezone"`
	Platform            string   `json:"platform"`
	OS                  string   `json:"os"`
	UserAgent           string   `json:"userAgent"`
	ViewportWidth       int64    `json:"viewportWidth"`
	ViewportHeight      int64    `json:"viewportHeight"`
	ScreenWidth         int64    `json:"screenWidth"`
	ScreenHeight        int64    `json:"screenHeight"`
	DeviceScaleFactor   float64  `json:"deviceScaleFactor"`
	HardwareConcurrency int64    `json:"hardwareConcurrency"`
	DeviceMemory        int64    `json:"deviceMemory"`
	WebGLVendor         string   `json:"webglVendor"`
	WebGLRenderer       string   `json:"webglRenderer"`
}

type CreateSessionInput struct {
	ProfilePath     string            `json:"profilePath,omitempty"`
	FingerprintSeed string            `json:"fingerprintSeed,omitempty"`
	TTLSeconds      int               `json:"ttlSec,omitempty"`
	LeaseID         string            `json:"leaseId,omitempty"`
	Fingerprint     FingerprintConfig `json:"fingerprint"`
	ProxyServer     string            `json:"proxyServer,omitempty"`
}

type Session struct {
	RuntimeSessionID string `json:"runtimeSessionId"`
	ProfilePath      string `json:"profilePath,omitempty"`
	CDPWsURL         string `json:"cdpWsUrl,omitempty"`
	LeaseID          string `json:"leaseId,omitempty"`
	ResolvedVersion  string `json:"resolvedVersion,omitempty"`
}

type CommitInput struct{}

type CommitResult struct {
	NewVersion string `json:"newVersion"`
	Bytes      int64  `json:"bytes"`
	DurationMs int64  `json:"durationMs"`
}

type NavigateInput struct {
	TaskControl
	IncludeSnapshot           bool   `json:"includeSnapshot,omitempty"`
	URL                       string `json:"url"`
	WaitUntil                 string `json:"waitUntil,omitempty"`
	AfterLoadScreenshotS3Path string `json:"afterLoadScreenshotS3Path,omitempty"`
}

type NavigateResult struct {
	Snapshot        *SnapshotResult `json:"snapshot,omitempty"`
	URL             string          `json:"url"`
	Title           string          `json:"title,omitempty"`
	SnapshotCleared bool            `json:"snapshotCleared"`
}

type SnapshotInput struct {
	Mode string `json:"mode,omitempty"`
}

type PageTable struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

// PageSnapshot preserves the versioned server observation verbatim. Use
// DecodeSnapshotPage for a validated in-memory view; business semantics remain with consumers.
type PageSnapshot map[string]any

type SnapshotResult struct {
	SnapshotID string       `json:"snapshotId"`
	Page       PageSnapshot `json:"page"`
}

type ActInput struct {
	ExpectedSnapshotID string   `json:"expectedSnapshotId,omitempty"`
	Action             string   `json:"action"`
	Ref                string   `json:"ref,omitempty"`
	X                  float64  `json:"x,omitempty"`
	Y                  float64  `json:"y,omitempty"`
	DeltaX             float64  `json:"deltaX,omitempty"`
	DeltaY             float64  `json:"deltaY,omitempty"`
	Text               string   `json:"text,omitempty"`
	HTML               string   `json:"html,omitempty"`
	Key                string   `json:"key,omitempty"`
	Value              string   `json:"value,omitempty"`
	Values             []string `json:"values,omitempty"`
	Clear              bool     `json:"clear,omitempty"`
	Submit             bool     `json:"submit,omitempty"`
	Button             string   `json:"button,omitempty"`
	ClickCount         int      `json:"clickCount,omitempty"`
	MotionProfile      string   `json:"motionProfile,omitempty"`
	TimeoutMs          int      `json:"timeoutMs,omitempty"`
}

type ActResult struct {
	OK     bool   `json:"ok"`
	Action string `json:"action"`
	Ref    string `json:"ref,omitempty"`
	URL    string `json:"url,omitempty"`
	Title  string `json:"title,omitempty"`
}

type WaitForCondition struct {
	Type        string   `json:"type"`
	Ref         string   `json:"ref,omitempty"`
	Selector    string   `json:"selector,omitempty"`
	Text        string   `json:"text,omitempty"`
	URLContains string   `json:"urlContains,omitempty"`
	URLMatches  string   `json:"urlMatches,omitempty"`
	Checks      []string `json:"checks,omitempty"`
}

type WaitForInput struct {
	TaskControl
	Condition  WaitForCondition `json:"condition"`
	IntervalMs int              `json:"intervalMs,omitempty"`
	StableMs   int              `json:"stableMs,omitempty"`
}

type WaitForResult struct {
	OK            bool     `json:"ok"`
	ConditionType string   `json:"conditionType"`
	Ref           string   `json:"ref,omitempty"`
	Text          string   `json:"text,omitempty"`
	X             float64  `json:"x,omitempty"`
	Y             float64  `json:"y,omitempty"`
	Checks        []string `json:"checks,omitempty"`
	URL           string   `json:"url,omitempty"`
	Title         string   `json:"title,omitempty"`
}

type ScreenshotInput struct {
	Mode               string `json:"mode,omitempty"`
	Selector           string `json:"selector,omitempty"`
	Format             string `json:"format,omitempty"`
	Quality            int    `json:"quality,omitempty"`
	ScreenshotS3Prefix string `json:"screenshotS3Prefix,omitempty"`
}

type ScreenshotResult struct {
	ScreenshotID string `json:"screenshotId"`
	S3Path       string `json:"s3Path"`
	ContentType  string `json:"contentType"`
	ByteLength   int    `json:"byteLength"`
}

type UploadFileSource struct {
	S3Path    string `json:"s3Path,omitempty"`
	LocalPath string `json:"localPath,omitempty"`
	URL       string `json:"url,omitempty"`
	Filename  string `json:"filename,omitempty"`
}

type UploadFilesInput struct {
	Ref       string             `json:"ref"`
	X         float64            `json:"x,omitempty"`
	Y         float64            `json:"y,omitempty"`
	Files     []UploadFileSource `json:"files"`
	TimeoutMs int                `json:"timeoutMs,omitempty"`
}

type UploadFilesResult struct {
	OK        bool     `json:"ok"`
	Ref       string   `json:"ref"`
	FileNames []string `json:"fileNames"`
}

type PageToolInput struct {
	TaskControl
	Method    string         `json:"method"`
	Payload   map[string]any `json:"payload,omitempty"`
	TimeoutMs int            `json:"timeoutMs,omitempty"`
}

type PageToolResult map[string]any

type PageToolBridgeConfig struct {
	Enabled bool   `json:"enabled"`
	Name    string `json:"name,omitempty"`
}

type EvaluateInput struct {
	TaskControl
	Script             string                `json:"script"`
	Args               []any                 `json:"args,omitempty"`
	World              string                `json:"world,omitempty"`
	AllowDuringHandoff bool                  `json:"allowDuringHandoff,omitempty"`
	PageToolBridge     *PageToolBridgeConfig `json:"pageToolBridge,omitempty"`
}

type EvaluateResult struct {
	Result any    `json:"result"`
	URL    string `json:"url"`
	Title  string `json:"title"`
}

type LiveViewInput struct {
	Permission string `json:"permission,omitempty"`
	TTLSeconds int    `json:"ttlSeconds,omitempty"`
}

type LiveViewResult struct {
	HandoffID  string    `json:"handoffId"`
	ViewerURL  string    `json:"viewerUrl"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Permission string    `json:"permission"`
}

type StartHandoffInput struct {
	Permission string `json:"permission,omitempty"`
	TTLSeconds int    `json:"ttlSeconds,omitempty"`
}

type HandoffResult struct {
	HandoffID  string    `json:"handoffId"`
	ViewerURL  string    `json:"viewerUrl"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Permission string    `json:"permission"`
}
