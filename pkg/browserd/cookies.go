package browserd

import (
	"context"
	"encoding/json"
	"net/http"
)

type CookieInput struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	URL      string `json:"url"`
	HTTPOnly bool   `json:"httpOnly,omitempty"`
}
type CookieConfig struct {
	ProfilePath string        `json:"profilePath,omitempty"`
	Revision    int64         `json:"revision,omitempty"`
	Cookies     []CookieInput `json:"cookies"`
}
type ImportCookiesResult struct {
	ImportedCount int `json:"importedCount"`
}

func (c *Client) ImportCookies(ctx context.Context, id string, in CookieConfig) (ImportCookiesResult, error) {
	var out ImportCookiesResult
	err := c.doJSON(ctx, http.MethodPost, sessionPath(id, "cookies/import"), in, &out)
	return out, err
}

type ImportProfileCookiesResult struct {
	UpdatedSessions int `json:"updatedSessions"`
}

func (c *Client) ImportProfileCookies(ctx context.Context, in CookieConfig) (ImportProfileCookiesResult, error) {
	var out ImportProfileCookiesResult
	err := c.doJSON(ctx, http.MethodPost, "/v1/profiles/cookies", in, &out)
	return out, err
}

type ProfileTransferInput struct {
	ProfilePath     string            `json:"profilePath"`
	Fingerprint     FingerprintConfig `json:"fingerprint"`
	FingerprintSeed string            `json:"fingerprintSeed"`
	Origins         []string          `json:"origins,omitempty"`
	Bundle          json.RawMessage   `json:"bundle,omitempty"`
	CookieRevision  int64             `json:"cookieRevision,omitempty"`
}

func (c *Client) ExportProfile(ctx context.Context, in ProfileTransferInput) (json.RawMessage, error) {
	var out json.RawMessage
	err := c.doJSON(ctx, http.MethodPost, "/v1/profiles/export", in, &out)
	return out, err
}
func (c *Client) ImportProfile(ctx context.Context, in ProfileTransferInput) (json.RawMessage, error) {
	var out json.RawMessage
	err := c.doJSON(ctx, http.MethodPost, "/v1/profiles/import", in, &out)
	return out, err
}
