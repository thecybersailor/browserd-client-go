package browserd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestClientPreservesTransportCause(t *testing.T) {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded, errors.New("transport failed")} {
		client, err := NewClient(Config{BaseURL: "https://browser.example", HTTPClient: &http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, cause }),
		}})
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Navigate(context.Background(), "rt_1", NavigateInput{TaskControl: TaskControl{RequestID: "test-request"}, URL: "https://example.com"})
		var typed Error
		if !AsError(err, &typed) || typed.Code != "browserd_unavailable" || !errors.Is(err, cause) {
			t.Fatalf("transport cause lost: %v", err)
		}
	}
}

func TestClientPreservesResponseDecodeCause(t *testing.T) {
	client, err := NewClient(Config{BaseURL: "https://browser.example", HTTPClient: &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("invalid json")), Header: make(http.Header)}, nil
		}),
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Navigate(context.Background(), "rt_1", NavigateInput{TaskControl: TaskControl{RequestID: "test-request"}, URL: "https://example.com"})
	var typed Error
	var syntax *json.SyntaxError
	if !AsError(err, &typed) || typed.Code != "browserd_response_invalid" || !errors.As(err, &syntax) {
		t.Fatalf("decode cause lost: %v", err)
	}
}
